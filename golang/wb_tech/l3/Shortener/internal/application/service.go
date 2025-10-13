package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/entity"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/Shortener/internal/domain/interfaces"
)

type shortenerService struct {
	urlsRepo      interfaces.ShortURLRepository
	analyticsRepo interfaces.AnalyticsRepository
}

func NewShortenerService(urls interfaces.ShortURLRepository, analytics interfaces.AnalyticsRepository) interfaces.ShortenerService {
	return &shortenerService{urlsRepo: urls, analyticsRepo: analytics}
}

func (s *shortenerService) Create(ctx context.Context, originalURL string) (entity.ShortURL, error) {
	if err := validateURL(originalURL); err != nil {
		return entity.ShortURL{}, err
	}

	code := generateCode(7)
	su := entity.ShortURL{
		ID:          uuid.NewString(),
		OriginalURL: originalURL,
		ShortCode:   code,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.urlsRepo.Save(ctx, su); err != nil {
		return entity.ShortURL{}, err
	}
	return su, nil
}

func (s *shortenerService) Resolve(ctx context.Context, shortCode string, userAgent string) (string, error) {
	su, err := s.urlsRepo.GetByCode(ctx, shortCode)
	if err != nil {
		return "", err
	}
	click := entity.ClickEvent{
		ID:        uuid.NewString(),
		ShortCode: shortCode,
		UserAgent: userAgent,
		CreatedAt: time.Now().UTC(),
	}
	_ = s.analyticsRepo.SaveClick(ctx, click)
	return su.OriginalURL, nil
}

func (s *shortenerService) Analytics(ctx context.Context, shortCode string) (map[string]any, error) {
	clicks, err := s.analyticsRepo.GetClicks(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	byDay, _ := s.analyticsRepo.CountByDay(ctx, shortCode)
	byMonth, _ := s.analyticsRepo.CountByMonth(ctx, shortCode)
	byUA, _ := s.analyticsRepo.CountByUserAgent(ctx, shortCode)

	byDayStr := make(map[string]int)
	for t, count := range byDay {
		byDayStr[t.Format("2006-01-02")] = count
	}

	result := map[string]any{
		"total":        len(clicks),
		"by_day":       byDayStr,
		"by_month":     byMonth,
		"by_userAgent": byUA,
		"events":       clicks,
	}
	return result, nil
}

func generateCode(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	s := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	s = strings.Map(func(r rune) rune {
		if r == '-' || r == '_' {
			return r
		}
		if r >= 'A' && r <= 'Z' {
			return r
		}
		if r >= 'a' && r <= 'z' {
			return r
		}
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
	if len(s) > n {
		return s[:n]
	}
	return s
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if parsedURL.Scheme == "" {
		return errors.New("URL must include scheme (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL scheme must be http or https")
	}

	if parsedURL.Host == "" {
		return errors.New("URL must include host")
	}

	return nil
}
