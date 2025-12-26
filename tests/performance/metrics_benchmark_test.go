package performance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkPerformanceMetricsEndpoint(b *testing.B) {
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/performance",
		nil,
	)

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

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)
	}
}
