package component_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// MockNotificationRepo for component tests.
type MockNotificationRepo struct {
	mock.Mock
}

func (m *MockNotificationRepo) GetNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.Notification, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf("mock error: %w", err)
	}

	notifications, _ := args.Get(0).([]dto.Notification)

	return notifications, args.Int(1), nil
}

func (m *MockNotificationRepo) CountNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)

	err := args.Error(1)
	if err != nil {
		return 0, fmt.Errorf("mock error: %w", err)
	}

	return args.Int(0), nil
}

func TestGetNotificationsComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	svc := service.NewNotificationService(mockRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	notificationID := uuid.New()
	now := time.Now()

	notifications := []dto.Notification{
		{
			NotificationID:   notificationID.String(),
			UserID:           userID.String(),
			Title:            "New follower",
			Message:          "John started following you",
			NotificationType: "follow",
			IsRead:           false,
			IsDeleted:        false,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	mockRepo.On("GetNotifications", mock.Anything, userID, 20, 0).Return(notifications, 5, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                         `json:"success"`
		Data    dto.NotificationListResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	resp := apiResp.Data
	assert.Equal(t, 5, resp.TotalCount)
	assert.Equal(t, 20, resp.Limit)
	assert.Equal(t, 0, resp.Offset)
	require.Len(t, resp.Notifications, 1)
	assert.Equal(t, "New follower", resp.Notifications[0].Title)
	assert.Equal(t, "follow", resp.Notifications[0].NotificationType)
}

func TestGetNotificationsComponent_CountOnly(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	svc := service.NewNotificationService(mockRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	mockRepo.On("CountNotifications", mock.Anything, userID).Return(42, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications?count_only=true", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                          `json:"success"`
		Data    dto.NotificationCountResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	assert.Equal(t, 42, apiResp.Data.TotalCount)

	// Verify that list-only fields are not present
	body := rr.Body.String()
	assert.NotContains(t, body, "notifications")
	assert.NotContains(t, body, `"limit"`)
	assert.NotContains(t, body, `"offset"`)
}

func TestGetNotificationsComponent_Pagination(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	svc := service.NewNotificationService(mockRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	mockRepo.On("GetNotifications", mock.Anything, userID, 10, 5).Return([]dto.Notification{}, 50, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications?limit=10&offset=5", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                         `json:"success"`
		Data    dto.NotificationListResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	resp := apiResp.Data
	assert.Equal(t, 50, resp.TotalCount)
	assert.Equal(t, 10, resp.Limit)
	assert.Equal(t, 5, resp.Offset)
	assert.Empty(t, resp.Notifications)
}

func TestGetNotificationsComponent_Unauthorized(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	svc := service.NewNotificationService(mockRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications", nil)
	// No X-User-Id header

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var apiResp struct {
		Success bool      `json:"success"`
		Error   dto.Error `json:"error"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.False(t, apiResp.Success)

	assert.Equal(t, "UNAUTHORIZED", apiResp.Error.Code)
}

func TestGetNotificationsComponent_ValidationError(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	svc := service.NewNotificationService(mockRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications?limit=invalid", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var apiResp struct {
		Success bool      `json:"success"`
		Error   dto.Error `json:"error"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.False(t, apiResp.Success)

	assert.Equal(t, "VALIDATION_ERROR", apiResp.Error.Code)
}

func TestGetNotificationsComponent_EmptyNotifications(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	svc := service.NewNotificationService(mockRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	// Return nil slice - service should convert to empty array
	mockRepo.On("GetNotifications", mock.Anything, userID, 20, 0).Return(nil, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify that notifications is an empty array, not null
	body := rr.Body.String()
	assert.Contains(t, body, `"notifications":[]`)
}
