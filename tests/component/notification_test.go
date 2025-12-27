package component_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *MockNotificationRepo) DeleteNotifications(
	ctx context.Context,
	userID uuid.UUID,
	notificationIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID, notificationIDs)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf("mock error: %w", err)
	}

	deletedIDs, _ := args.Get(0).([]uuid.UUID)

	return deletedIDs, nil
}

func (m *MockNotificationRepo) MarkNotificationRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, userID, notificationID)

	err := args.Error(1)
	if err != nil {
		return false, fmt.Errorf("mock error: %w", err)
	}

	return args.Bool(0), nil
}

func (m *MockNotificationRepo) MarkAllNotificationsRead(
	ctx context.Context,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf("mock error: %w", err)
	}

	readIDs, _ := args.Get(0).([]uuid.UUID)

	return readIDs, nil
}

func TestGetNotificationsComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

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
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	mockRepo.On("CountNotifications", mock.Anything, userID).Return(42, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications?countOnly=true", nil)
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
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

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
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

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
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

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
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

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

func TestDeleteNotificationsComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()

	mockRepo.On("DeleteNotifications", mock.Anything, userID, mock.Anything).
		Return([]uuid.UUID{notificationID1, notificationID2}, nil)

	reqBody := fmt.Sprintf(`{"notificationIds": ["%s", "%s"]}`, notificationID1.String(), notificationID2.String())
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/notifications", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                           `json:"success"`
		Data    dto.NotificationDeleteResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	resp := apiResp.Data
	assert.Equal(t, "Notifications deleted successfully", resp.Message)
	assert.Len(t, resp.DeletedNotificationIDs, 2)
}

func TestDeleteNotificationsComponent_PartialSuccess(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()

	// Only one notification is deleted
	mockRepo.On("DeleteNotifications", mock.Anything, userID, mock.Anything).
		Return([]uuid.UUID{notificationID1}, nil)

	reqBody := fmt.Sprintf(`{"notificationIds": ["%s", "%s"]}`, notificationID1.String(), notificationID2.String())
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/notifications", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusPartialContent, rr.Code)

	var apiResp struct {
		Success bool                           `json:"success"`
		Data    dto.NotificationDeleteResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	resp := apiResp.Data
	assert.Equal(t, "Some notifications deleted successfully", resp.Message)
	assert.Len(t, resp.DeletedNotificationIDs, 1)
}

func TestDeleteNotificationsComponent_NotFound(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	notificationID := uuid.New()

	// No notifications found
	mockRepo.On("DeleteNotifications", mock.Anything, userID, mock.Anything).
		Return([]uuid.UUID{}, nil)

	reqBody := fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID.String())
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/notifications", strings.NewReader(reqBody))
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var apiResp struct {
		Success bool      `json:"success"`
		Error   dto.Error `json:"error"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.False(t, apiResp.Success)

	assert.Equal(t, "NOT_FOUND", apiResp.Error.Code)
}

func TestDeleteNotificationsComponent_Unauthorized(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	notificationID := uuid.New()

	reqBody := fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID.String())
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user-management/notifications", strings.NewReader(reqBody))
	// No X-User-Id header
	req.Header.Set("Content-Type", "application/json")

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

//nolint:funlen // Table-driven test with multiple test cases
func TestMarkNotificationReadComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupMock      func(*MockNotificationRepo, uuid.UUID, uuid.UUID)
		userID         *uuid.UUID // nil means no header
		notificationID string
		expectedStatus int
		expectedBody   []string
	}{
		{
			name: "Success",
			setupMock: func(m *MockNotificationRepo, userID, notifID uuid.UUID) {
				m.On("MarkNotificationRead", mock.Anything, userID, notifID).Return(true, nil)
			},
			userID:         func() *uuid.UUID { id := uuid.New(); return &id }(),
			notificationID: uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`, `"message":"Notification marked as read successfully"`},
		},
		{
			name: "NotFound",
			setupMock: func(m *MockNotificationRepo, userID, notifID uuid.UUID) {
				m.On("MarkNotificationRead", mock.Anything, userID, notifID).Return(false, nil)
			},
			userID:         func() *uuid.UUID { id := uuid.New(); return &id }(),
			notificationID: uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{`"success":false`, `"NOT_FOUND"`},
		},
		{
			name:           "Unauthorized",
			setupMock:      func(_ *MockNotificationRepo, _, _ uuid.UUID) {},
			userID:         nil,
			notificationID: uuid.New().String(),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{`"success":false`, `"UNAUTHORIZED"`},
		},
		{
			name:           "InvalidNotificationId",
			setupMock:      func(_ *MockNotificationRepo, _, _ uuid.UUID) {},
			userID:         func() *uuid.UUID { id := uuid.New(); return &id }(),
			notificationID: "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockNotificationRepo)
			mockUserRepo := new(MockUserRepo)
			svc := service.NewNotificationService(mockRepo, mockUserRepo)

			c := &app.Container{
				NotificationService: svc,
				Config:              config.Instance,
			}
			c.HealthService = service.NewHealthService(nil, nil)

			srv := server.NewServerWithContainer(c)
			handler := srv.Handler

			var userID, notifID uuid.UUID
			if tt.userID != nil {
				userID = *tt.userID
			}

			parsedID, parseErr := uuid.Parse(tt.notificationID)
			if parseErr == nil {
				notifID = parsedID
			}

			tt.setupMock(mockRepo, userID, notifID)

			reqPath := fmt.Sprintf("/api/v1/user-management/notifications/%s/read", tt.notificationID)
			req := httptest.NewRequest(http.MethodPut, reqPath, nil)

			if tt.userID != nil {
				req.Header.Set("X-User-Id", tt.userID.String())
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			body := rr.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestMarkAllNotificationsReadComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()
	notificationID3 := uuid.New()

	mockRepo.On("MarkAllNotificationsRead", mock.Anything, userID).
		Return([]uuid.UUID{notificationID1, notificationID2, notificationID3}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/user-management/notifications/read-all", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                            `json:"success"`
		Data    dto.NotificationReadAllResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	resp := apiResp.Data
	assert.Equal(t, "All notifications marked as read successfully", resp.Message)
	assert.Len(t, resp.ReadNotificationIDs, 3)
	assert.Contains(t, resp.ReadNotificationIDs, notificationID1.String())
	assert.Contains(t, resp.ReadNotificationIDs, notificationID2.String())
	assert.Contains(t, resp.ReadNotificationIDs, notificationID3.String())
}

func TestMarkAllNotificationsReadComponent_EmptyResult(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	// No unread notifications
	mockRepo.On("MarkAllNotificationsRead", mock.Anything, userID).
		Return([]uuid.UUID{}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/user-management/notifications/read-all", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify response contains empty array
	body := rr.Body.String()
	assert.Contains(t, body, `"success":true`)
	assert.Contains(t, body, `"readNotificationIds":[]`)
}

//nolint:dupl // Similar pattern to other unauthorized tests but for different endpoint
func TestMarkAllNotificationsReadComponent_Unauthorized(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	req := httptest.NewRequest(http.MethodPut, "/api/v1/user-management/notifications/read-all", nil)
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

func TestGetNotificationPreferencesComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	mockUserRepo.On("FindNotificationPreferencesByUserID", mock.Anything, userID).
		Return(&dto.NotificationPreferences{EmailNotifications: true}, nil)
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, userID).
		Return(&dto.PrivacyPreferences{ProfileVisibility: "public"}, nil)
	mockUserRepo.On("FindDisplayPreferencesByUserID", mock.Anything, userID).
		Return(&dto.DisplayPreferences{Theme: "dark"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user-management/notifications/preferences", nil)
	req.Header.Set("X-User-Id", userID.String())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                       `json:"success"`
		Data    dto.UserPreferenceResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	require.NotNil(t, apiResp.Data.Preferences.NotificationPreferences, "NotificationPreferences shouldn't be nil")
	assert.True(t, apiResp.Data.Preferences.NotificationPreferences.EmailNotifications)

	require.NotNil(t, apiResp.Data.Preferences.PrivacyPreferences, "PrivacyPreferences shouldn't be nil")
	assert.Equal(t, "public", apiResp.Data.Preferences.PrivacyPreferences.ProfileVisibility)

	require.NotNil(t, apiResp.Data.Preferences.DisplayPreferences, "DisplayPreferences shouldn't be nil")
	assert.Equal(t, "dark", apiResp.Data.Preferences.DisplayPreferences.Theme)
}

func TestUpdateNotificationPreferencesComponent_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockNotificationRepo)
	mockUserRepo := new(MockUserRepo)
	svc := service.NewNotificationService(mockRepo, mockUserRepo)

	c := &app.Container{
		NotificationService: svc,
		Config:              config.Instance,
	}
	c.HealthService = service.NewHealthService(nil, nil)

	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	userID := uuid.New()

	// 1. Defined Initial State (Domain Objects)
	initialNotif := &dto.NotificationPreferences{
		EmailNotifications: false, // Will be updated to true
		PushNotifications:  true,  // Will be updated to false
		LikeNotifications:  true,  // Will stay true (not in request)
	}
	initialPrivacy := &dto.PrivacyPreferences{
		ProfileVisibility: "public", // Will be updated
		ShowEmail:         true,     // Will stay true
	}
	initialDisplay := &dto.DisplayPreferences{
		Theme:    "dark", // Will be updated
		Language: "en",
	}

	// 2. Define Request (Partial Update)
	emailTrue := true
	pushFalse := false
	profileFollowers := "followers_only"
	themeLight := "light"

	reqBodyObj := dto.UpdateUserPreferenceRequest{
		NotificationPreferences: &dto.UpdateNotificationPreferencesRequest{
			EmailNotifications: &emailTrue,
			PushNotifications:  &pushFalse,
			// LikeNotifications omitted, should remain true
		},
		PrivacyPreferences: &dto.UpdatePrivacyPreferencesRequest{
			ProfileVisibility: &profileFollowers,
		},
		DisplayPreferences: &dto.UpdateDisplayPreferencesRequest{
			Theme: &themeLight,
		},
	}
	reqBytes, _ := json.Marshal(reqBodyObj)

	// 3. Define Expected Merged State
	expectedNotif := &dto.NotificationPreferences{
		EmailNotifications: true,
		PushNotifications:  false,
		LikeNotifications:  true, // Preserved
	}
	expectedPrivacy := &dto.PrivacyPreferences{
		ProfileVisibility: "followers_only",
		ShowEmail:         true, // Preserved
	}
	expectedDisplay := &dto.DisplayPreferences{
		Theme:    "light",
		Language: "en", // Preserved
	}

	// 4. Setup Mocks
	// Service first fetches current prefs
	mockUserRepo.On("FindNotificationPreferencesByUserID", mock.Anything, userID).Return(initialNotif, nil)
	mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, userID).Return(initialPrivacy, nil)
	mockUserRepo.On("FindDisplayPreferencesByUserID", mock.Anything, userID).Return(initialDisplay, nil)

	// Service then updates with merged prefs
	// Check that correct merged objects are passed
	mockUserRepo.On("UpdateNotificationPreferences", mock.Anything, userID, expectedNotif).Return(nil)
	mockUserRepo.On("UpdatePrivacyPreferences", mock.Anything, userID, expectedPrivacy).Return(nil)
	mockUserRepo.On("UpdateDisplayPreferences", mock.Anything, userID, expectedDisplay).Return(nil)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/user-management/notifications/preferences",
		bytes.NewReader(reqBytes),
	)
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var apiResp struct {
		Success bool                       `json:"success"`
		Data    dto.UserPreferenceResponse `json:"data"`
	}

	err := json.Unmarshal(rr.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)

	// Assert response contains merged values
	assert.True(t, apiResp.Data.Preferences.NotificationPreferences.EmailNotifications)
	assert.False(t, apiResp.Data.Preferences.NotificationPreferences.PushNotifications)
	assert.True(t, apiResp.Data.Preferences.NotificationPreferences.LikeNotifications)
	assert.Equal(t, "followers_only", apiResp.Data.Preferences.PrivacyPreferences.ProfileVisibility)
	assert.True(t, apiResp.Data.Preferences.PrivacyPreferences.ShowEmail)
	assert.Equal(t, "light", apiResp.Data.Preferences.DisplayPreferences.Theme)
}
