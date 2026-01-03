package dependency_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/spf13/viper"
)

var testHandler http.Handler

// MockHealthChecker mocks health checking.
type MockHealthChecker struct{}

func (m *MockHealthChecker) Health(ctx context.Context) map[string]string {
	return map[string]string{
		"status": "up",
	}
}

func (m *MockHealthChecker) Close() error {
	return nil
}

// MockMetricsService mocks metrics service.
type MockMetricsService struct{}

func (m *MockMetricsService) GetPerformanceMetrics(ctx context.Context) (*dto.PerformanceMetricsResponse, error) {
	return &dto.PerformanceMetricsResponse{
		RequestCounts: dto.RequestCounts{TotalRequests: 100},
		Database:      dto.DatabaseMetrics{ActiveConnections: 5},
	}, nil
}

func (m *MockMetricsService) GetCacheMetrics(ctx context.Context) (*dto.CacheMetricsResponse, error) {
	return &dto.CacheMetricsResponse{
		HitRate: 0.5,
	}, nil
}

func (m *MockMetricsService) GetSystemMetrics(ctx context.Context) (*dto.SystemMetricsResponse, error) {
	return &dto.SystemMetricsResponse{
		UptimeSeconds: 3600,
	}, nil
}

func (m *MockMetricsService) GetDetailedHealthMetrics(ctx context.Context) (*dto.DetailedHealthMetricsResponse, error) {
	return &dto.DetailedHealthMetricsResponse{
		OverallStatus: "healthy",
		Timestamp:     time.Now(),
	}, nil
}

// MockAdminService mocks admin service.
type MockAdminService struct{}

func (m *MockAdminService) ClearCache(ctx context.Context, keyPattern string) (*dto.CacheClearResponse, error) {
	return &dto.CacheClearResponse{
		Message:      "Cache cleared successfully",
		Pattern:      keyPattern,
		ClearedCount: 10,
	}, nil
}

func TestMain(m *testing.M) {
	// Point viper to the project root config directory
	viper.AddConfigPath("../../config")

	// Load the real configuration from files
	cfg := config.Load()

	// Create container with injected mock dependencies for health checks
	mockHealth := &MockHealthChecker{}
	container, _ := app.NewContainer(app.ContainerConfig{
		Config:   cfg,
		Database: mockHealth,
		Cache:    mockHealth,
	})

	// Inject mock metrics service
	container.MetricsService = &MockMetricsService{}
	container.AdminService = &MockAdminService{}

	// Initialize the router with container
	srv := server.NewServerWithContainer(container)
	testHandler = srv.Handler

	code := m.Run()
	os.Exit(code)
}
