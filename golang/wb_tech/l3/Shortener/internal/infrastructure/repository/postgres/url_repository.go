package postgres

import (
	"context"
	"database/sql"

	"github.com/wb-go/wbf/dbpg"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/infrastructure/repository/postgres/dto"
	"go.uber.org/zap"
)

type URLRepository struct {
	db     *dbpg.DB
	logger *zap.Logger
}

func NewURLRepository(db *dbpg.DB, logger *zap.Logger) *URLRepository {
	return &URLRepository{db: db, logger: logger}

}

func (r *URLRepository) Save(ctx context.Context, u entity.ShortURL) error {
	urlDTO := dto.ShortURLDTO{
		ID:          u.ID,
		OriginalURL: u.OriginalURL,
		ShortCode:   u.ShortCode,
		CreatedAt:   u.CreatedAt,
	}

	_, err := r.db.ExecContext(ctx, `
        insert into short_urls(id, original_url, short_code, created_at)
        values ($1, $2, $3, $4)
        on conflict (short_code) do nothing
    `, urlDTO.ID, urlDTO.OriginalURL, urlDTO.ShortCode, urlDTO.CreatedAt)
	if err != nil {
		r.logger.Error("failed to save short URL", zap.Error(err))
	}
	return err
}

func (r *URLRepository) GetByCode(ctx context.Context, code string) (entity.ShortURL, error) {
	var urlDTO dto.ShortURLDTO
	row := r.db.QueryRowContext(ctx, `
        select id, original_url, short_code, created_at
        from short_urls where short_code=$1
    `, code)
	err := row.Scan(&urlDTO.ID, &urlDTO.OriginalURL, &urlDTO.ShortCode, &urlDTO.CreatedAt)
	if err == sql.ErrNoRows {
		r.logger.Error("short URL not found", zap.String("short_code", code))
		return entity.ShortURL{}, err
	}

	u := entity.ShortURL{
		ID:          urlDTO.ID,
		OriginalURL: urlDTO.OriginalURL,
		ShortCode:   urlDTO.ShortCode,
		CreatedAt:   urlDTO.CreatedAt,
	}

	return u, err
}
