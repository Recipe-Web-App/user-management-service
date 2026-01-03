//nolint:dupl // Benchmark tests have intentional duplicate setup patterns for isolation
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

func seedBenchmarkUserForSocial(b *testing.B, db *sql.DB, userID uuid.UUID, username string) {
	b.Helper()

	ctx := context.Background()

	queryUser := `
		INSERT INTO recipe_manager.users
			(user_id, username, email, full_name, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := db.ExecContext(ctx, queryUser,
		userID,
		username,
		username+"@example.com",
		username+" Full Name",
		"not_a_real_hash",
		true,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		b.Fatalf("failed to seed user %s: %v", username, err)
	}

	queryPrivacy := `
		INSERT INTO recipe_manager.user_privacy_preferences (user_id, profile_visibility, contact_info_visibility)
		VALUES ($1, 'PUBLIC', 'PUBLIC')
		ON CONFLICT (user_id) DO UPDATE SET profile_visibility = 'PUBLIC', contact_info_visibility = 'PUBLIC'
	`

	_, err = db.ExecContext(ctx, queryPrivacy, userID)
	if err != nil {
		_, _ = db.ExecContext(ctx, "DELETE FROM recipe_manager.users WHERE user_id = $1", userID)

		b.Fatalf("failed to seed privacy for %s: %v", username, err)
	}

	b.Cleanup(func() {
		cleanupCtx := context.Background()
		// Delete follows first to avoid foreign key constraint issues
		_, _ = db.ExecContext(cleanupCtx,
			"DELETE FROM recipe_manager.user_follows WHERE follower_id = $1 OR followee_id = $1", userID)
		_, _ = db.ExecContext(cleanupCtx,
			"DELETE FROM recipe_manager.user_privacy_preferences WHERE user_id = $1", userID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM recipe_manager.users WHERE user_id = $1", userID)
	})
}

func seedFollowRelationship(b *testing.B, db *sql.DB, followerID, followeeID uuid.UUID) {
	b.Helper()

	ctx := context.Background()

	query := `
		INSERT INTO recipe_manager.user_follows (follower_id, followee_id, followed_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	_, err := db.ExecContext(ctx, query, followerID, followeeID, time.Now())
	if err != nil {
		b.Fatalf("failed to seed follow relationship: %v", err)
	}

	b.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = db.ExecContext(cleanupCtx,
			"DELETE FROM recipe_manager.user_follows WHERE follower_id = $1 AND followee_id = $2",
			followerID, followeeID)
	})
}

func BenchmarkGetFollowing(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user whose following list we'll fetch
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "social_target_"+targetUserID.String()[:8])

	// Create users for target to follow
	for i := range 10 {
		followeeID := uuid.New()
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followeeID, fmt.Sprintf("followee_%d_%s", i, followeeID.String()[:8]))
		seedFollowRelationship(b, dbSvc.GetDB(), targetUserID, followeeID)
	}

	// Create requester user
	requesterID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, "requester_"+requesterID.String()[:8])

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/following", targetUserID)
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

func BenchmarkGetFollowingConcurrent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user whose following list we'll fetch
	targetUserID := uuid.New()
	targetUsername := "social_target_conc_" + targetUserID.String()[:8]
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, targetUsername)

	// Create users for target to follow
	for i := range 10 {
		followeeID := uuid.New()
		followeeName := fmt.Sprintf("followee_conc_%d_%s", i, followeeID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followeeID, followeeName)
		seedFollowRelationship(b, dbSvc.GetDB(), targetUserID, followeeID)
	}

	// Create requester user
	requesterID := uuid.New()
	requesterName := "requester_conc_" + requesterID.String()[:8]
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, requesterName)

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/following", targetUserID)

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

func BenchmarkGetFollowingCountOnly(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "social_count_"+targetUserID.String()[:8])

	// Create users for target to follow
	for i := range 20 {
		followeeID := uuid.New()
		followeeName := fmt.Sprintf("followee_count_%d_%s", i, followeeID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followeeID, followeeName)
		seedFollowRelationship(b, dbSvc.GetDB(), targetUserID, followeeID)
	}

	// Create requester user
	requesterID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, "requester_count_"+requesterID.String()[:8])

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/following?countOnly=true", targetUserID)
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

func BenchmarkGetFollowingLargePagination(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "social_large_"+targetUserID.String()[:8])

	// Create more users for target to follow (testing larger result sets)
	for i := range 50 {
		followeeID := uuid.New()
		followeeName := fmt.Sprintf("followee_large_%d_%s", i, followeeID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followeeID, followeeName)
		seedFollowRelationship(b, dbSvc.GetDB(), targetUserID, followeeID)
	}

	// Create requester user
	requesterID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, "requester_large_"+requesterID.String()[:8])

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/following?limit=100", targetUserID)
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

func BenchmarkGetFollowers(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user whose followers list we'll fetch
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "followers_target_"+targetUserID.String()[:8])

	// Create users who follow the target
	for i := range 10 {
		followerID := uuid.New()
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, fmt.Sprintf("follower_%d_%s", i, followerID.String()[:8]))
		seedFollowRelationship(b, dbSvc.GetDB(), followerID, targetUserID)
	}

	// Create requester user
	requesterID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, "requester_followers_"+requesterID.String()[:8])

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/followers", targetUserID)
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

func BenchmarkGetFollowersConcurrent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user whose followers list we'll fetch
	targetUserID := uuid.New()
	targetUsername := "followers_target_conc_" + targetUserID.String()[:8]
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, targetUsername)

	// Create users who follow the target
	for i := range 10 {
		followerID := uuid.New()
		followerName := fmt.Sprintf("follower_conc_%d_%s", i, followerID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, followerName)
		seedFollowRelationship(b, dbSvc.GetDB(), followerID, targetUserID)
	}

	// Create requester user
	requesterID := uuid.New()
	requesterName := "requester_followers_conc_" + requesterID.String()[:8]
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, requesterName)

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/followers", targetUserID)

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

func BenchmarkGetFollowersCountOnly(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "followers_count_"+targetUserID.String()[:8])

	// Create users who follow the target
	for i := range 20 {
		followerID := uuid.New()
		followerName := fmt.Sprintf("follower_count_%d_%s", i, followerID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, followerName)
		seedFollowRelationship(b, dbSvc.GetDB(), followerID, targetUserID)
	}

	// Create requester user
	requesterID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, "requester_followers_count_"+requesterID.String()[:8])

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/followers?countOnly=true", targetUserID)
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

func BenchmarkGetFollowersLargePagination(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "followers_large_"+targetUserID.String()[:8])

	// Create more users who follow the target (testing larger result sets)
	for i := range 50 {
		followerID := uuid.New()
		followerName := fmt.Sprintf("follower_large_%d_%s", i, followerID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, followerName)
		seedFollowRelationship(b, dbSvc.GetDB(), followerID, targetUserID)
	}

	// Create requester user
	requesterID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), requesterID, "requester_followers_large_"+requesterID.String()[:8])

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/followers?limit=100", targetUserID)
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

func BenchmarkFollowUser(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create follower user
	followerID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, "follow_follower_"+followerID.String()[:8])

	// Create many target users to follow (one per iteration)
	targetUserIDs := make([]uuid.UUID, b.N)
	for i := range b.N {
		targetID := uuid.New()
		targetUserIDs[i] = targetID
		targetName := fmt.Sprintf("follow_target_%d_%s", i, targetID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetID, targetName)
	}

	b.ResetTimer()

	for i := range b.N {
		reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/follow/%s", followerID, targetUserIDs[i])
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
		req.Header.Set("X-User-Id", followerID.String())

		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkFollowUserConcurrent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user that many followers will follow
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "follow_target_conc_"+targetUserID.String()[:8])

	// Create many follower users
	followerIDs := make([]uuid.UUID, 100)

	for i := range 100 {
		followerID := uuid.New()
		followerIDs[i] = followerID
		followerName := fmt.Sprintf("follow_follower_conc_%d_%s", i, followerID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, followerName)
	}

	followerIdx := 0

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Cycle through followers
			idx := followerIdx % len(followerIDs)
			followerIdx++
			followerID := followerIDs[idx]

			reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/follow/%s", followerID, targetUserID)
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
			req.Header.Set("X-User-Id", followerID.String())

			rr := httptest.NewRecorder()
			benchmarkHandler.ServeHTTP(rr, req)

			// 200 OK is expected (idempotent)
			if rr.Code != http.StatusOK {
				b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
			}
		}
	})
}

func BenchmarkFollowUserIdempotent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create follower and target users
	followerID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, "follow_idem_follower_"+followerID.String()[:8])

	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "follow_idem_target_"+targetUserID.String()[:8])

	// Create initial follow relationship
	seedFollowRelationship(b, dbSvc.GetDB(), followerID, targetUserID)

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/follow/%s", followerID, targetUserID)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
	req.Header.Set("X-User-Id", followerID.String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		// Should still return 200 OK (idempotent via ON CONFLICT DO NOTHING)
		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkUnfollowUser(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create follower user
	followerID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, "unfollow_follower_"+followerID.String()[:8])

	// Create many target users to unfollow (one per iteration)
	targetUserIDs := make([]uuid.UUID, b.N)
	for i := range b.N {
		targetID := uuid.New()
		targetUserIDs[i] = targetID
		targetName := fmt.Sprintf("unfollow_target_%d_%s", i, targetID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetID, targetName)
		// Create initial follow relationship to unfollow
		seedFollowRelationship(b, dbSvc.GetDB(), followerID, targetID)
	}

	b.ResetTimer()

	for i := range b.N {
		reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/follow/%s", followerID, targetUserIDs[i])
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, reqPath, nil)
		req.Header.Set("X-User-Id", followerID.String())

		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}

func BenchmarkUnfollowUserConcurrent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create target user that many followers will unfollow
	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "unfollow_target_conc_"+targetUserID.String()[:8])

	// Create many follower users with existing follow relationships
	followerIDs := make([]uuid.UUID, 100)

	for i := range 100 {
		followerID := uuid.New()
		followerIDs[i] = followerID
		followerName := fmt.Sprintf("unfollow_follower_conc_%d_%s", i, followerID.String()[:8])
		seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, followerName)
		seedFollowRelationship(b, dbSvc.GetDB(), followerID, targetUserID)
	}

	followerIdx := 0

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Cycle through followers
			idx := followerIdx % len(followerIDs)
			followerIdx++
			followerID := followerIDs[idx]

			reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/follow/%s", followerID, targetUserID)
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, reqPath, nil)
			req.Header.Set("X-User-Id", followerID.String())

			rr := httptest.NewRecorder()
			benchmarkHandler.ServeHTTP(rr, req)

			// 200 OK is expected (idempotent)
			if rr.Code != http.StatusOK {
				b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
			}
		}
	})
}

func BenchmarkUnfollowUserIdempotent(b *testing.B) {
	if benchmarkContainer == nil || benchmarkContainer.Database == nil {
		b.Fatal("benchmark container or database is nil")
	}

	dbSvc, ok := benchmarkContainer.Database.(*database.Service)
	if !ok {
		b.Fatal("failed to cast database service")
	}

	cfg := benchmarkContainer.Config.Postgres
	b.Logf("DEBUG: DB Host=%s Port=%d User='%s' DBName=%s", cfg.Host, cfg.Port, cfg.User, cfg.Database)

	// Create follower and target users
	followerID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), followerID, "unfollow_idem_follower_"+followerID.String()[:8])

	targetUserID := uuid.New()
	seedBenchmarkUserForSocial(b, dbSvc.GetDB(), targetUserID, "unfollow_idem_target_"+targetUserID.String()[:8])

	// No initial follow relationship - testing idempotent behavior when not following

	reqPath := fmt.Sprintf("/api/v1/user-management/users/%s/follow/%s", followerID, targetUserID)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, reqPath, nil)
	req.Header.Set("X-User-Id", followerID.String())

	for b.Loop() {
		rr := httptest.NewRecorder()
		benchmarkHandler.ServeHTTP(rr, req)

		// Should still return 200 OK (idempotent - unfollow when not following)
		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d, body: %s", rr.Code, rr.Body.String())
		}
	}
}
