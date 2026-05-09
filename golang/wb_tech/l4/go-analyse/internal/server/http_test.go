package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-analyse/internal/metrics"
)

func TestHealthHandler(t *testing.T) {
	handler := NewHandler(metrics.NewRegistry(100))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("body = %q, want JSON health response", body)
	}
}

func TestMetricsHandlerReturnsPrometheusText(t *testing.T) {
	handler := NewHandler(metrics.NewRegistry(100))

	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Fatalf("content-type = %q, want Prometheus text response", contentType)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "go_analyse_gc_configured_percent") {
		t.Fatalf("metrics response missing custom GC metric\n%s", body)
	}

	if !strings.Contains(body, "go_goroutines") {
		t.Fatalf("metrics response missing built-in Go collector metric\n%s", body)
	}
}

func TestPprofIndexIsMounted(t *testing.T) {
	handler := NewHandler(metrics.NewRegistry(100))

	request := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if !strings.Contains(recorder.Body.String(), "profile") {
		t.Fatalf("pprof index body does not look like pprof index")
	}
}
