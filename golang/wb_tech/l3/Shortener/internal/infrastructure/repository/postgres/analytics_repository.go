package postgres

import (
	"context"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/infrastructure/repository/postgres/dto"
	"go.uber.org/zap"
)

type AnalyticsRepository struct {
	db     *dbpg.DB
	logger *zap.Logger
}

func NewAnalyticsRepository(db *dbpg.DB, logger *zap.Logger) *AnalyticsRepository {
	return &AnalyticsRepository{db: db, logger: logger}
}

func (r *AnalyticsRepository) SaveClick(ctx context.Context, click entity.ClickEvent) error {
	clickDTO := dto.ClickEventDTO{
		ID:        click.ID,
		ShortCode: click.ShortCode,
		UserAgent: click.UserAgent,
		CreatedAt: click.CreatedAt,
	}

	_, err := r.db.ExecContext(ctx, `
        insert into click_events(id, short_code, user_agent, created_at)
        values ($1, $2, $3, $4)
    `, clickDTO.ID, clickDTO.ShortCode, clickDTO.UserAgent, clickDTO.CreatedAt)
	if err != nil {
		r.logger.Error("failed to save click event", zap.Error(err))
	}
	return err
}

func (r *AnalyticsRepository) GetClicks(ctx context.Context, code string) ([]entity.ClickEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
        select id, short_code, user_agent, created_at
        from click_events
        where short_code=$1
        order by created_at asc
    `, code)
	if err != nil {
		r.logger.Error("failed to get clicks", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	var res []entity.ClickEvent
	for rows.Next() {
		var clickDTO dto.ClickEventDTO
		if err := rows.Scan(&clickDTO.ID, &clickDTO.ShortCode, &clickDTO.UserAgent, &clickDTO.CreatedAt); err != nil {
			r.logger.Error("failed to scan click event", zap.Error(err))
			return nil, err
		}

		c := entity.ClickEvent{
			ID:        clickDTO.ID,
			ShortCode: clickDTO.ShortCode,
			UserAgent: clickDTO.UserAgent,
			CreatedAt: clickDTO.CreatedAt,
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

func (r *AnalyticsRepository) CountByDay(ctx context.Context, code string) (map[time.Time]int, error) {
	rows, err := r.db.QueryContext(ctx, `
        select date_trunc('day', created_at at time zone 'utc') as day, count(*)
        from click_events
        where short_code=$1
        group by day
        order by day
    `, code)
	if err != nil {
		r.logger.Error("failed to count by day", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	out := make(map[time.Time]int)
	for rows.Next() {
		var t time.Time
		var c int
		if err := rows.Scan(&t, &c); err != nil {
			r.logger.Error("failed to scan day", zap.Error(err))
			return nil, err
		}
		out[t.UTC()] = c
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) CountByMonth(ctx context.Context, code string) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
        select to_char(date_trunc('month', created_at at time zone 'utc'), 'YYYY-MM') as month, count(*)
        from click_events
        where short_code=$1
        group by month
        order by month
    `, code)
	if err != nil {
		r.logger.Error("failed to count by month", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var m string
		var c int
		if err := rows.Scan(&m, &c); err != nil {
			r.logger.Error("failed to scan month", zap.Error(err))
			return nil, err
		}
		out[m] = c
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) CountByUserAgent(ctx context.Context, code string) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `
        select user_agent, count(*)
        from click_events
        where short_code=$1
        group by user_agent
        order by user_agent
    `, code)
	if err != nil {
		r.logger.Error("failed to count by user agent", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var ua string
		var c int
		if err := rows.Scan(&ua, &c); err != nil {
			r.logger.Error("failed to scan user agent", zap.Error(err))
			return nil, err
		}
		out[ua] = c
	}
	return out, rows.Err()
}
