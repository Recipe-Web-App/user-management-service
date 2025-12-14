package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHealthChecker is a mock implementation of repository.HealthChecker.
type mockHealthChecker struct {
	healthStatus map[string]string
	closeErr     error
}

func (m *mockHealthChecker) Health(_ context.Context) map[string]string {
	return m.healthStatus
}

func (m *mockHealthChecker) Close() error {
	return m.closeErr
}

func TestHealthService_GetHealth(t *testing.T) {
	t.Parallel()

	svc := NewHealthService(nil, nil)
	status := svc.GetHealth(context.Background())

	assert.Equal(t, "UP", status.Status)
}

func TestHealthService_GetReadiness_AllUp(t *testing.T) {
	t.Parallel()

	mockDB := &mockHealthChecker{
		healthStatus: map[string]string{"status": "up", "message": "database is healthy"},
	}
	mockCache := &mockHealthChecker{
		healthStatus: map[string]string{"status": "up", "message": "redis is healthy"},
	}

	svc := NewHealthService(mockDB, mockCache)
	status := svc.GetReadiness(context.Background())

	assert.Equal(t, "READY", status.Status)
	assert.Equal(t, "up", status.Database["status"])
	assert.Equal(t, "up", status.Redis["status"])
}

func TestHealthService_GetReadiness_DatabaseDown(t *testing.T) {
	t.Parallel()

	mockDB := &mockHealthChecker{
		healthStatus: map[string]string{"status": "down", "error": "connection refused"},
	}
	mockCache := &mockHealthChecker{
		healthStatus: map[string]string{"status": "up", "message": "redis is healthy"},
	}

	svc := NewHealthService(mockDB, mockCache)
	status := svc.GetReadiness(context.Background())

	assert.Equal(t, "DEGRADED", status.Status)
	assert.Equal(t, "down", status.Database["status"])
	assert.Equal(t, "up", status.Redis["status"])
}

func TestHealthService_GetReadiness_CacheDown(t *testing.T) {
	t.Parallel()

	mockDB := &mockHealthChecker{
		healthStatus: map[string]string{"status": "up", "message": "database is healthy"},
	}
	mockCache := &mockHealthChecker{
		healthStatus: map[string]string{"status": "down", "error": "connection refused"},
	}

	svc := NewHealthService(mockDB, mockCache)
	status := svc.GetReadiness(context.Background())

	assert.Equal(t, "DEGRADED", status.Status)
	assert.Equal(t, "up", status.Database["status"])
	assert.Equal(t, "down", status.Redis["status"])
}

func TestHealthService_GetReadiness_NilDependencies(t *testing.T) {
	t.Parallel()

	svc := NewHealthService(nil, nil)
	status := svc.GetReadiness(context.Background())

	assert.Equal(t, "DEGRADED", status.Status)
	assert.Equal(t, "down", status.Database["status"])
	assert.Equal(t, "database not configured", status.Database["message"])
	assert.Equal(t, "down", status.Redis["status"])
	assert.Equal(t, "cache not configured", status.Redis["message"])
}
