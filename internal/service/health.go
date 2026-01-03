// Package service contains business logic for the application.
package service

import (
	"context"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

// HealthService handles health-related business logic.
type HealthService struct {
	db    repository.HealthChecker
	cache repository.HealthChecker
}

// NewHealthService creates a new health service.
func NewHealthService(db, cache repository.HealthChecker) *HealthService {
	return &HealthService{
		db:    db,
		cache: cache,
	}
}

// HealthStatus represents the overall health status.
type HealthStatus struct {
	Status   string            `json:"status"`
	Database map[string]string `json:"database,omitempty"`
	Redis    map[string]string `json:"redis,omitempty"`
}

// GetHealth returns simple health status (liveness).
func (s *HealthService) GetHealth(_ context.Context) HealthStatus {
	return HealthStatus{Status: "UP"}
}

// GetReadiness returns detailed readiness status.
func (s *HealthService) GetReadiness(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Status: "READY",
	}

	if s.db != nil {
		status.Database = s.db.Health(ctx)
	} else {
		status.Database = map[string]string{"status": "down", "message": "database not configured"}
	}

	if s.cache != nil {
		status.Redis = s.cache.Health(ctx)
	} else {
		status.Redis = map[string]string{"status": "down", "message": "cache not configured"}
	}

	// Determine overall status
	dbStatus := status.Database["status"]
	redisStatus := status.Redis["status"]

	if dbStatus != "up" || redisStatus != "up" {
		status.Status = "DEGRADED"
	}

	return status
}
