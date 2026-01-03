package performance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
)

func BenchmarkGetUserStats(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Seed some data to make stats meaningful
	for range 5 {
		userID := uuid.New()
		seedBenchmarkUser(b, dbSvc.GetDB(), userID)
	}

	requesterID := uuid.New()
	reqPath := "/api/v1/user-management/admin/users/stats"
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	req.Header.Set("X-User-Id", requesterID.String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkGetUserStatsConcurrent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Seed some data
	for range 5 {
		userID := uuid.New()
		seedBenchmarkUser(b, dbSvc.GetDB(), userID)
	}

	requesterID := uuid.New()
	reqPath := "/api/v1/user-management/admin/users/stats"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
			req.Header.Set("X-User-Id", requesterID.String())
			rr := httptest.NewRecorder()
			benchmarkHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
			}
		}
	})
}

func BenchmarkClearCache(b *testing.B) {
	if benchmarkContainer == nil {
		b.Fatal("benchmark container is nil")
	}

	reqPath := "/api/v1/user-management/admin/cache/clear"

	requesterID := uuid.New()

	for b.Loop() {
		reqBody := strings.NewReader(`{"keyPattern": "*"}`)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, reqBody)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Id", requesterID.String())

		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}
