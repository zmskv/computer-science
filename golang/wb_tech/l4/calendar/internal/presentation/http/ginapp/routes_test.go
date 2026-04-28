package ginapp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"calendar/internal/application"
	"calendar/internal/infrastructure/logging"
	"calendar/internal/infrastructure/repository"
	"calendar/internal/presentation/http/ginapp/middleware"

	"github.com/gin-gonic/gin"
)

func TestCalendarHTTPCreateAndQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.NewAsyncLogger(io.Discard, 16)
	defer logger.Close()

	repo := repository.NewInMemoryRepo()
	service := application.NewService(repo, nil, logger, application.Config{
		ArchiveInterval: time.Hour,
		ArchiveAfter:    time.Hour,
	})
	service.Start(ctx)

	router := gin.New()
	router.Use(middleware.LoggerMiddleware(logger))
	InitRoutes(router, service, logger)

	body := map[string]any{
		"user_id": 1,
		"date":    "2026-04-28T10:00:00Z",
		"event":   "Team sync",
	}
	payload, _ := json.Marshal(body)

	createReq := httptest.NewRequest(http.MethodPost, "/calendar/create_event", bytes.NewReader(payload))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)

	if createResp.Code != http.StatusCreated {
		t.Fatalf("create_event status = %d, want %d, body = %s", createResp.Code, http.StatusCreated, createResp.Body.String())
	}

	queryReq := httptest.NewRequest(http.MethodGet, "/calendar/events_for_day?user_id=1&date=2026-04-28", nil)
	queryResp := httptest.NewRecorder()
	router.ServeHTTP(queryResp, queryReq)

	if queryResp.Code != http.StatusOK {
		t.Fatalf("events_for_day status = %d, want %d, body = %s", queryResp.Code, http.StatusOK, queryResp.Body.String())
	}

	var response struct {
		Result []struct {
			Title string `json:"event"`
		} `json:"result"`
	}
	if err := json.Unmarshal(queryResp.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(response.Result) != 1 || response.Result[0].Title != "Team sync" {
		t.Fatalf("events_for_day result = %+v, want one Team sync event", response.Result)
	}
}
