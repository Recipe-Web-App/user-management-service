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

	var apiResp dto.UserStatsResponse

	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	assert.Equal(t, expectedStats, &apiResp)

	mockRepo.AssertExpectations(t)
}
