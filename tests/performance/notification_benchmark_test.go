package performance_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
)

func seedBenchmarkNotifications(b *testing.B, db *sql.DB, userID uuid.UUID, count int) {
	b.Helper()

	ctx := context.Background()

	// First ensure the user exists
	queryUser := `
		INSERT INTO recipe_manager.users
			(user_id, username, email, full_name, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO NOTHING
	`

	_, err := db.ExecContext(ctx, queryUser,
		userID,
		"notif_user_"+userID.String()[:8],
		fmt.Sprintf("notif_%s@example.com", userID.String()[:8]),
		"Notification Benchmark User",
		"not_a_real_hash",
		true,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		b.Fatalf("failed to seed user: %v", err)
	}

	// Insert notifications
	queryNotification := `
		INSERT INTO recipe_manager.notifications
			(notification_id, user_id, title, message, notification_type, is_read, is_deleted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	notificationIDs := make([]uuid.UUID, count)

	for i := range count {
		notificationID := uuid.New()
		notificationIDs[i] = notificationID

		_, err := db.ExecContext(ctx, queryNotification,
			notificationID,
			userID,
			fmt.Sprintf("Test Notification %d", i),
			fmt.Sprintf("This is test notification message number %d", i),
			"follow",
			false,
			false,
			now.Add(time.Duration(-i)*time.Minute),
			now.Add(time.Duration(-i)*time.Minute),
		)
		if err != nil {
			b.Fatalf("failed to seed notification %d: %v", i, err)
		}
	}

	b.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM recipe_manager.notifications WHERE user_id = $1", userID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM recipe_manager.users WHERE user_id = $1", userID)
	})
}

func BenchmarkGetNotifications(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Skip("benchmark container or database is nil - skipping benchmark")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	userID := uuid.New()

	// Seed 50 notifications for this user
	seedBenchmarkNotifications(b, dbSvc.GetDB(), userID, 50)

	reqPath := "/api/v1/user-management/notifications"
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	req.Header.Set("X-User-Id", userID.String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkGetNotificationsWithPagination(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Skip("benchmark container or database is nil - skipping benchmark")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	userID := uuid.New()

	// Seed 100 notifications for this user
	seedBenchmarkNotifications(b, dbSvc.GetDB(), userID, 100)

	reqPath := "/api/v1/user-management/notifications?limit=10&offset=50"
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	req.Header.Set("X-User-Id", userID.String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkGetNotificationsCountOnly(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Skip("benchmark container or database is nil - skipping benchmark")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	userID := uuid.New()

	// Seed 50 notifications for this user
	seedBenchmarkNotifications(b, dbSvc.GetDB(), userID, 50)

	reqPath := "/api/v1/user-management/notifications?count_only=true"
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	req.Header.Set("X-User-Id", userID.String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}
