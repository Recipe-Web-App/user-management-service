package performance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func BenchmarkPerformanceMetricsEndpoint(b *testing.B) {
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/performance",
		nil,
	)
	req.Header.Set("X-User-Id", uuid.New().String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)
	}
}

func BenchmarkCacheMetricsEndpoint(b *testing.B) {
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/cache",
		nil,
	)
	req.Header.Set("X-User-Id", uuid.New().String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)
	}
}

func BenchmarkSystemMetricsEndpoint(b *testing.B) {
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/system",
		nil,
	)
	req.Header.Set("X-User-Id", uuid.New().String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)
	}
}

func BenchmarkDetailedHealthMetricsEndpoint(b *testing.B) {
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/health/detailed",
		nil,
	)
	req.Header.Set("X-User-Id", uuid.New().String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)
	}
}
