package performance_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
)

func seedBenchmarkUser(b *testing.B, db *sql.DB, userID uuid.UUID) {
	b.Helper()

	ctx := context.Background()

	queryUser := `
		INSERT INTO recipe_manager.users
			(user_id, username, email, full_name, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := db.ExecContext(ctx, queryUser,
		userID,
		fmt.Sprintf("perf_user_%s", userID),
		fmt.Sprintf("perf_%s@example.com", userID),
		"Perf User",
		"not_a_real_hash",
		true,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		b.Fatalf("failed to seed user: %v", err)
	}

	queryPrivacy := `
		INSERT INTO recipe_manager.user_privacy_preferences (user_id, profile_visibility, contact_info_visibility)
		VALUES ($1, 'PUBLIC', 'PUBLIC')
		ON CONFLICT (user_id) DO UPDATE SET profile_visibility = 'PUBLIC', contact_info_visibility = 'PUBLIC'
	`

	_, err = db.ExecContext(ctx, queryPrivacy, userID)
	if err != nil {
		_, _ = db.ExecContext(ctx, "DELETE FROM recipe_manager.users WHERE user_id = $1", userID)

		b.Fatalf("failed to seed privacy: %v", err)
	}

	b.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM recipe_manager.user_privacy_preferences WHERE user_id = $1", userID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM recipe_manager.users WHERE user_id = $1", userID)
	})
}

func BenchmarkGetUserProfile(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	targetUserID := uuid.New()
	requesterID := uuid.New()

	seedBenchmarkUser(b, dbSvc.GetDB(), targetUserID)

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/profile", targetUserID)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	req.Header.Set("X-User-Id", requesterID.String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rr.Code)
		}
	}
}

func BenchmarkUpdateUserProfile(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	userID := uuid.New()

	seedBenchmarkUser(b, dbSvc.GetDB(), userID)

	reqPath := "/api/v1/user-management/users/profile"
	reqBody := `{"bio": "Updated bio for benchmark test"}`

	for b.Loop() {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, reqPath, strings.NewReader(reqBody))
		req.Header.Set("X-User-Id", userID.String())
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkRequestAccountDeletion(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	userID := uuid.New()

	seedBenchmarkUser(b, dbSvc.GetDB(), userID)

	reqPath := "/api/v1/user-management/users/account/delete-request"

	for b.Loop() {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
		req.Header.Set("X-User-Id", userID.String())

		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkRequestAccountDeletionConcurrent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	userID := uuid.New()

	seedBenchmarkUser(b, dbSvc.GetDB(), userID)

	reqPath := "/api/v1/user-management/users/account/delete-request"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
			req.Header.Set("X-User-Id", userID.String())

			rr := httptest.NewRecorder()
			benchmarkHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
			}
		}
	})
}

func BenchmarkSearchUsers(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Seed a few users for search results
	for range 5 {
		userID := uuid.New()
		seedBenchmarkUser(b, dbSvc.GetDB(), userID)
	}

	requesterID := uuid.New()
	seedBenchmarkUser(b, dbSvc.GetDB(), requesterID)

	reqPath := "/api/v1/user-management/users/search?query=perf&limit=10"
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

func BenchmarkSearchUsersConcurrent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Seed a few users for search results
	for range 5 {
		userID := uuid.New()
		seedBenchmarkUser(b, dbSvc.GetDB(), userID)
	}

	requesterID := uuid.New()
	seedBenchmarkUser(b, dbSvc.GetDB(), requesterID)

	reqPath := "/api/v1/user-management/users/search?query=perf&limit=10"

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
