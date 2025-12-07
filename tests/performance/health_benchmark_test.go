package performance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkHealthEndpoint(b *testing.B) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-management/health", nil)

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)
	}
}

func BenchmarkReadyEndpoint(b *testing.B) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-management/ready", nil)

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)
	}
}
