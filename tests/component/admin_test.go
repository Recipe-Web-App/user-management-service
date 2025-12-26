package component_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetUserStatsComponent_Success(t *testing.T) {
	t.Parallel()

	// Setup Mocks
	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)

	// Create Service
	userService := service.NewUserService(mockRepo, mockTokenStore)

	// Create Container
	c := &app.Container{
		UserService: userService,
		Config:      config.Instance,
	}
	// Setup Health so router doesn't panic
	c.HealthService = service.NewHealthService(nil, nil)

	// Create Server
	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	// Mock Expectation
	expectedStats := &dto.UserStatsResponse{
		TotalUsers:        100,
		ActiveUsers:       80,
		InactiveUsers:     20,
		NewUsersToday:     5,
		NewUsersThisWeek:  15,
		NewUsersThisMonth: 30,
	}

	mockRepo.On("GetUserStats", mock.Anything).Return(expectedStats, nil)

	// Execute
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/admin/users/stats", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp struct {
		Success bool                  `json:"success"`
		Data    dto.UserStatsResponse `json:"data"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)
	assert.Equal(t, expectedStats, &apiResp.Data)

	mockRepo.AssertExpectations(t)
}

func TestAdminEndpoints_Removed(t *testing.T) {
	t.Parallel()

	// Setup
	c := &app.Container{
		Config: config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	// Create dummy service to satisfy AdminHandler constructor
	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	userService := service.NewUserService(mockRepo, mockTokenStore)
	c.UserService = userService

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	// List of removed endpoints
	removedPaths := []string{
		"/api/v1/user-management/admin/redis/session-stats",
		"/api/v1/user-management/admin/health",
	}

	for _, path := range removedPaths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Should return 404 (Not Found)
		assert.Equal(t, http.StatusNotFound, w.Code, "Expected 404 for removed endpoint: %s", path)
	}

	// ForceLogout (POST)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-management/admin/users/123/force-logout", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// ClearSessions (DELETE)
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/admin/redis/sessions", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
