package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/application"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/SalesTracker/internal/domain/entity"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrItemNotFound = errors.New("item not found")
)

type Repository struct {
	db *dbpg.DB
}

func NewRepository(ctx context.Context, masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*Repository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, fmt.Errorf("create postgres connection: %w", err)
	}

	if err := db.Master.PingContext(ctx); err != nil {
		_ = db.Master.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() {
	if r.db == nil {
		return
	}

	if r.db.Master != nil {
		_ = r.db.Master.Close()
	}

	for _, slave := range r.db.Slaves {
		if slave != nil {
			_ = slave.Close()
		}
	}
}

func (r *Repository) CreateItem(ctx context.Context, item entity.Item) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO items (id, item_type, amount, category, description, occurred_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		item.ID,
		item.Type,
		item.Amount,
		item.Category,
		item.Description,
		item.OccurredAt.UTC(),
		item.CreatedAt.UTC(),
		item.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	return nil
}

func (r *Repository) GetItem(ctx context.Context, itemID string) (entity.Item, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, item_type, amount, category, description, occurred_at, created_at, updated_at
		FROM items
		WHERE id = $1`,
		itemID,
	)

	item, err := scanItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Item{}, application.ErrItemNotFound
		}
		return entity.Item{}, err
	}

	return item, nil
}

func (r *Repository) ListItems(ctx context.Context, filter entity.ItemFilter) ([]entity.Item, error) {
	whereClause, args := buildWhereClause(commonFilters{
		From:     filter.From,
		To:       filter.To,
		Type:     filter.Type,
		Category: filter.Category,
	})

	query := fmt.Sprintf(
		`SELECT id, item_type, amount, category, description, occurred_at, created_at, updated_at
		FROM items
		%s
		ORDER BY %s %s, created_at DESC`,
		whereClause,
		sortColumn(filter.SortBy),
		strings.ToUpper(string(filter.SortOrder)),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]entity.Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (r *Repository) UpdateItem(ctx context.Context, item entity.Item) error {
	result, err := r.db.Master.ExecContext(
		ctx,
		`UPDATE items
		SET item_type = $2,
			amount = $3,
			category = $4,
			description = $5,
			occurred_at = $6,
			updated_at = $7
		WHERE id = $1`,
		item.ID,
		item.Type,
		item.Amount,
		item.Category,
		item.Description,
		item.OccurredAt.UTC(),
		item.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read update result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

func (r *Repository) DeleteItem(ctx context.Context, itemID string) error {
	result, err := r.db.Master.ExecContext(
		ctx,
		`DELETE FROM items
		WHERE id = $1`,
		itemID,
	)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrItemNotFound
	}

	return nil
}

func (r *Repository) GetAnalytics(ctx context.Context, filter entity.AnalyticsFilter) (entity.AnalyticsResult, error) {
	whereClause, args := buildWhereClause(commonFilters{
		From:     filter.From,
		To:       filter.To,
		Type:     filter.Type,
		Category: filter.Category,
	})

	result := entity.AnalyticsResult{
		Points: []entity.AnalyticsPoint{},
	}

	summaryQuery := fmt.Sprintf(
		`SELECT
			COALESCE(SUM(amount), 0),
			COALESCE(AVG(amount), 0),
			COUNT(*)::bigint,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount), 0),
			COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount), 0)
		FROM items
		%s`,
		whereClause,
	)

	if err := r.db.QueryRowContext(ctx, summaryQuery, args...).Scan(
		&result.Summary.Sum,
		&result.Summary.Avg,
		&result.Summary.Count,
		&result.Summary.Median,
		&result.Summary.Percentile90,
	); err != nil {
		return entity.AnalyticsResult{}, fmt.Errorf("get analytics summary: %w", err)
	}

	if filter.GroupBy == entity.AnalyticsGroupNone {
		return result, nil
	}

	groupExpr := analyticsGroupExpression(filter.GroupBy)
	pointsQuery := fmt.Sprintf(
		`SELECT
			%s AS group_label,
			COALESCE(SUM(amount), 0),
			COALESCE(AVG(amount), 0),
			COUNT(*)::bigint
		FROM items
		%s
		GROUP BY 1
		ORDER BY 1 ASC`,
		groupExpr,
		whereClause,
	)

	rows, err := r.db.QueryContext(ctx, pointsQuery, args...)
	if err != nil {
		return entity.AnalyticsResult{}, fmt.Errorf("get analytics points: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var point entity.AnalyticsPoint
		if err := rows.Scan(&point.Group, &point.Sum, &point.Avg, &point.Count); err != nil {
			return entity.AnalyticsResult{}, fmt.Errorf("scan analytics point: %w", err)
		}

		point.Label = point.Group
		result.Points = append(result.Points, point)
	}

	if err := rows.Err(); err != nil {
		return entity.AnalyticsResult{}, fmt.Errorf("iterate analytics points: %w", err)
	}

	return result, nil
}

type commonFilters struct {
	From     *time.Time
	To       *time.Time
	Type     entity.ItemType
	Category string
}

func buildWhereClause(filters commonFilters) (string, []any) {
	conditions := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if filters.From != nil {
		args = append(args, filters.From.UTC())
		conditions = append(conditions, fmt.Sprintf("occurred_at >= $%d", len(args)))
	}

	if filters.To != nil {
		args = append(args, filters.To.UTC())
		conditions = append(conditions, fmt.Sprintf("occurred_at <= $%d", len(args)))
	}

	if filters.Type != "" {
		args = append(args, filters.Type)
		conditions = append(conditions, fmt.Sprintf("item_type = $%d", len(args)))
	}

	if filters.Category != "" {
		args = append(args, filters.Category)
		conditions = append(conditions, fmt.Sprintf("category = $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func sortColumn(field entity.SortField) string {
	switch field {
	case entity.SortFieldAmount:
		return "amount"
	case entity.SortFieldCategory:
		return "category"
	case entity.SortFieldType:
		return "item_type"
	case entity.SortFieldCreatedAt:
		return "created_at"
	default:
		return "occurred_at"
	}
}

func analyticsGroupExpression(group entity.AnalyticsGroup) string {
	switch group {
	case entity.AnalyticsGroupWeek:
		return "TO_CHAR(DATE_TRUNC('week', occurred_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD')"
	case entity.AnalyticsGroupMonth:
		return "TO_CHAR(DATE_TRUNC('month', occurred_at AT TIME ZONE 'UTC'), 'YYYY-MM')"
	case entity.AnalyticsGroupCategory:
		return "category"
	default:
		return "TO_CHAR(DATE_TRUNC('day', occurred_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD')"
	}
}

func scanItem(scanner interface {
	Scan(dest ...any) error
}) (entity.Item, error) {
	var item entity.Item

	if err := scanner.Scan(
		&item.ID,
		&item.Type,
		&item.Amount,
		&item.Category,
		&item.Description,
		&item.OccurredAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return entity.Item{}, err
	}

	item.OccurredAt = item.OccurredAt.UTC()
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()

	return item, nil
}
