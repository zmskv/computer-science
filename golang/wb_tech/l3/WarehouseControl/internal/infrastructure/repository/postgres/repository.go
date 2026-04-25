package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wb-go/wbf/dbpg"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/WarehouseControl/internal/domain/entity"
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

func (r *Repository) CreateItem(ctx context.Context, actor entity.Actor, item entity.Item) error {
	return r.withActorTx(ctx, actor, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO items (id, name, sku, quantity, location, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			item.ID,
			item.Name,
			item.SKU,
			item.Quantity,
			item.Location,
			item.Description,
			item.CreatedAt.UTC(),
			item.UpdatedAt.UTC(),
		)
		if err != nil {
			return fmt.Errorf("create item: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetItem(ctx context.Context, itemID string) (entity.Item, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, sku, quantity, location, description, created_at, updated_at
		FROM items
		WHERE id = $1`,
		itemID,
	)

	item, err := scanItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Item{}, entity.ErrItemNotFound
		}
		return entity.Item{}, fmt.Errorf("get item: %w", err)
	}

	return item, nil
}

func (r *Repository) ListItems(ctx context.Context, filter entity.ItemFilter) ([]entity.Item, error) {
	whereClause, args := buildItemsWhereClause(filter)

	rows, err := r.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`SELECT id, name, sku, quantity, location, description, created_at, updated_at
			FROM items
			%s
			ORDER BY updated_at DESC, created_at DESC`,
			whereClause,
		),
		args...,
	)
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

func (r *Repository) UpdateItem(ctx context.Context, actor entity.Actor, item entity.Item) error {
	return r.withActorTx(ctx, actor, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(
			ctx,
			`UPDATE items
			SET name = $2,
				sku = $3,
				quantity = $4,
				location = $5,
				description = $6,
				updated_at = $7
			WHERE id = $1`,
			item.ID,
			item.Name,
			item.SKU,
			item.Quantity,
			item.Location,
			item.Description,
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
			return entity.ErrItemNotFound
		}

		return nil
	})
}

func (r *Repository) DeleteItem(ctx context.Context, actor entity.Actor, itemID string) error {
	return r.withActorTx(ctx, actor, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM items WHERE id = $1`, itemID)
		if err != nil {
			return fmt.Errorf("delete item: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("read delete result: %w", err)
		}
		if rowsAffected == 0 {
			return entity.ErrItemNotFound
		}

		return nil
	})
}

func (r *Repository) ListHistory(ctx context.Context, filter entity.HistoryFilter) ([]entity.HistoryEntry, error) {
	whereClause, args := buildHistoryWhereClause(filter)

	rows, err := r.db.QueryContext(
		ctx,
		fmt.Sprintf(
			`SELECT id, item_id, action, changed_by, changed_role, changed_at, old_data, new_data
			FROM item_history
			%s
			ORDER BY changed_at DESC, id DESC`,
			whereClause,
		),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	defer rows.Close()

	history := make([]entity.HistoryEntry, 0)
	for rows.Next() {
		entry, err := scanHistoryEntry(rows)
		if err != nil {
			return nil, err
		}

		history = append(history, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history: %w", err)
	}

	return history, nil
}

func (r *Repository) withActorTx(ctx context.Context, actor entity.Actor, fn func(tx *sql.Tx) error) (err error) {
	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.current_user', $1, true)`, actor.Username); err != nil {
		return fmt.Errorf("set current user: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.current_role', $1, true)`, string(actor.Role)); err != nil {
		return fmt.Errorf("set current role: %w", err)
	}

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func buildItemsWhereClause(filter entity.ItemFilter) (string, []any) {
	if strings.TrimSpace(filter.Query) == "" {
		return "", nil
	}

	search := "%" + strings.ToLower(strings.TrimSpace(filter.Query)) + "%"

	return `WHERE LOWER(name) LIKE $1 OR LOWER(sku) LIKE $1 OR LOWER(location) LIKE $1`, []any{search}
}

func buildHistoryWhereClause(filter entity.HistoryFilter) (string, []any) {
	conditions := make([]string, 0, 5)
	args := make([]any, 0, 5)

	if filter.ItemID != "" {
		args = append(args, filter.ItemID)
		conditions = append(conditions, fmt.Sprintf("item_id = $%d", len(args)))
	}
	if filter.Username != "" {
		args = append(args, filter.Username)
		conditions = append(conditions, fmt.Sprintf("changed_by = $%d", len(args)))
	}
	if filter.Action != "" {
		args = append(args, filter.Action)
		conditions = append(conditions, fmt.Sprintf("action = $%d", len(args)))
	}
	if filter.From != nil {
		args = append(args, filter.From.UTC())
		conditions = append(conditions, fmt.Sprintf("changed_at >= $%d", len(args)))
	}
	if filter.To != nil {
		args = append(args, filter.To.UTC())
		conditions = append(conditions, fmt.Sprintf("changed_at <= $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func scanItem(scanner interface {
	Scan(dest ...any) error
}) (entity.Item, error) {
	var item entity.Item

	if err := scanner.Scan(
		&item.ID,
		&item.Name,
		&item.SKU,
		&item.Quantity,
		&item.Location,
		&item.Description,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return entity.Item{}, err
	}

	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()

	return item, nil
}

func scanHistoryEntry(scanner interface {
	Scan(dest ...any) error
}) (entity.HistoryEntry, error) {
	var entry entity.HistoryEntry
	var oldData []byte
	var newData []byte

	if err := scanner.Scan(
		&entry.ID,
		&entry.ItemID,
		&entry.Action,
		&entry.ChangedBy,
		&entry.ChangedRole,
		&entry.ChangedAt,
		&oldData,
		&newData,
	); err != nil {
		return entity.HistoryEntry{}, err
	}

	entry.ChangedAt = entry.ChangedAt.UTC()
	entry.OldData = unmarshalJSONMap(oldData)
	entry.NewData = unmarshalJSONMap(newData)

	return entry, nil
}

func unmarshalJSONMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}

	data := make(map[string]any)
	if err := json.Unmarshal(raw, &data); err != nil {
		return map[string]any{}
	}

	return data
}
