package component_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
)

// authMockMetricsService is a mock implementation for testing.
type authMockMetricsService struct{}

func (m *authMockMetricsService) GetPerformanceMetrics(_ context.Context) (*dto.PerformanceMetricsResponse, error) {
	return &dto.PerformanceMetricsResponse{
		RequestCounts: dto.RequestCounts{TotalRequests: 100},
		ResponseTimes: dto.ResponseTimes{AverageMs: 42.0},
	}, nil
}

func (m *authMockMetricsService) GetCacheMetrics(_ context.Context) (*dto.CacheMetricsResponse, error) {
	return &dto.CacheMetricsResponse{HitRate: 0.9}, nil
}

func (m *authMockMetricsService) GetSystemMetrics(_ context.Context) (*dto.SystemMetricsResponse, error) {
	return &dto.SystemMetricsResponse{
		System:  dto.SystemInfo{CPUUsagePercent: 25.5},
		Process: dto.ProcessInfo{NumThreads: 10},
	}, nil
}

func (m *authMockMetricsService) GetDetailedHealthMetrics(
	_ context.Context,
) (*dto.DetailedHealthMetricsResponse, error) {
	return &dto.DetailedHealthMetricsResponse{OverallStatus: "healthy"}, nil
}

func TestAuthMiddleware_HealthRoutes_NoAuthRequired(t *testing.T) {
	t.Parallel()

	c := &app.Container{
		Config:        config.Instance,
		HealthService: service.NewHealthService(nil, nil),
	}

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{
			name:           "health endpoint without auth",
			endpoint:       "/api/v1/user-management/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ready endpoint without auth",
			endpoint:       "/api/v1/user-management/ready",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestAuthMiddleware_ProtectedRoutes_RequireAuth(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)

	userService := service.NewUserService(mockRepo, mockTokenStore, nil)

	c := &app.Container{
		UserService:   userService,
		Config:        config.Instance,
		HealthService: service.NewHealthService(nil, nil),
	}

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	tests := []struct {
		name           string
		endpoint       string
		method         string
		expectedStatus int
	}{
		{
			name:           "admin stats without auth",
			endpoint:       "/api/v1/user-management/admin/users/stats",
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "users list without auth",
			endpoint:       "/api/v1/user-management/users",
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "metrics without auth",
			endpoint:       "/api/v1/user-management/metrics/performance",
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.endpoint, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestAuthMiddleware_ProtectedRoutes_WithValidAuth(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	mockMetrics := &authMockMetricsService{}

	userService := service.NewUserService(mockRepo, mockTokenStore, nil)

	c := &app.Container{
		UserService:    userService,
		MetricsService: mockMetrics,
		Config:         config.Instance,
		HealthService:  service.NewHealthService(nil, nil),
	}

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	tests := []struct {
		name           string
		endpoint       string
		method         string
		expectedStatus int
	}{
		{
			name:           "metrics with valid auth",
			endpoint:       "/api/v1/user-management/metrics/performance",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "metrics system with valid auth",
			endpoint:       "/api/v1/user-management/metrics/system",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.endpoint, nil)
			req.Header.Set("X-User-Id", uuid.New().String())

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestAuthMiddleware_InvalidUserID(t *testing.T) {
	t.Parallel()

	c := &app.Container{
		Config:        config.Instance,
		HealthService: service.NewHealthService(nil, nil),
	}

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	tests := []struct {
		name           string
		userIDHeader   string
		expectedStatus int
	}{
		{
			name:           "empty user id",
			userIDHeader:   "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid uuid format",
			userIDHeader:   "not-a-uuid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "partial uuid",
			userIDHeader:   "12345678-1234-1234-1234",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/admin/users/stats", nil)
			if tc.userIDHeader != "" {
				req.Header.Set("X-User-Id", tc.userIDHeader)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
