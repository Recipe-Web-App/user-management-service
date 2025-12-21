package dependency_test

import (
	"context"
	"encoding/json"
	"errors"
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
)

// errNotificationTest is used for testing repository errors.
var errNotificationTest = errors.New("database connection failed")

const (
	notificationsBaseURL = "/api/v1/user-management/notifications"
)

// MockNotificationRepository is a mock implementation of repository.NotificationRepository.
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) GetNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.Notification, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf("get notifications: %w", err)
	}

	notifications, _ := args.Get(0).([]dto.Notification)

	return notifications, args.Int(1), nil
}

func (m *MockNotificationRepository) CountNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)

	err := args.Error(1)
	if err != nil {
		return 0, fmt.Errorf("count notifications: %w", err)
	}

	return args.Int(0), nil
}

type notificationTestFixture struct {
	handler  http.Handler
	mockRepo *MockNotificationRepository
	userID   uuid.UUID
}

func setupNotificationTest(t *testing.T) *notificationTestFixture {
	t.Helper()

	mockRepo := new(MockNotificationRepository)
	cfg := &config.Config{}

	container, err := app.NewContainer(app.ContainerConfig{
		Config:           cfg,
		NotificationRepo: mockRepo,
	})
	require.NoError(t, err)

	srv := server.NewServerWithContainer(container)

	return &notificationTestFixture{
		handler:  srv.Handler,
		mockRepo: mockRepo,
		userID:   uuid.New(),
	}
}

func newNotificationRequest(t *testing.T, userID uuid.UUID, queryParams string) *http.Request {
	t.Helper()

	reqPath := notificationsBaseURL + queryParams
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set("X-User-Id", userID.String())

	return req
}

func createTestNotification(userID uuid.UUID) dto.Notification {
	now := time.Now()

	return dto.Notification{
		NotificationID:   uuid.New().String(),
		UserID:           userID.String(),
		Title:            "Test Notification",
		Message:          "This is a test notification message",
		NotificationType: "follow",
		IsRead:           false,
		IsDeleted:        false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

//nolint:funlen // Integration test with many test cases
func TestGetNotifications(t *testing.T) {
	t.Parallel()

	t.Run("Success_ReturnsList", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notification := createTestNotification(fix.userID)
		notifications := []dto.Notification{notification}

		fix.mockRepo.On("GetNotifications", mock.Anything, fix.userID, 20, 0).
			Return(notifications, 10, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, ""))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.NotificationListResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.True(t, resp.Success)
		assert.Equal(t, 10, resp.Data.TotalCount)
		assert.Equal(t, 20, resp.Data.Limit)
		assert.Equal(t, 0, resp.Data.Offset)
		require.Len(t, resp.Data.Notifications, 1)
		assert.Equal(t, notification.Title, resp.Data.Notifications[0].Title)
	})

	t.Run("Success_CountOnly", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("CountNotifications", mock.Anything, fix.userID).
			Return(42, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?count_only=true"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                          `json:"success"`
			Data    dto.NotificationCountResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.True(t, resp.Success)
		assert.Equal(t, 42, resp.Data.TotalCount)
	})

	t.Run("Success_Pagination", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("GetNotifications", mock.Anything, fix.userID, 10, 5).
			Return([]dto.Notification{}, 50, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?limit=10&offset=5"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.NotificationListResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.True(t, resp.Success)
		assert.Equal(t, 50, resp.Data.TotalCount)
		assert.Equal(t, 10, resp.Data.Limit)
		assert.Equal(t, 5, resp.Data.Offset)
		assert.Empty(t, resp.Data.Notifications)
	})

	t.Run("Success_EmptyNotifications", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("GetNotifications", mock.Anything, fix.userID, 20, 0).
			Return(nil, 0, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, ""))

		require.Equal(t, http.StatusOK, rr.Code)

		body := rr.Body.String()
		// Should be empty array, not null
		assert.Contains(t, body, `"notifications":[]`)
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, notificationsBaseURL, nil)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var resp struct {
			Success bool      `json:"success"`
			Error   dto.Error `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.False(t, resp.Success)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("Unauthorized_InvalidUUID", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, notificationsBaseURL, nil)
		require.NoError(t, err)
		req.Header.Set("X-User-Id", "not-a-valid-uuid")

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("BadRequest_InvalidLimit", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?limit=abc"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp struct {
			Success bool      `json:"success"`
			Error   dto.Error `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.False(t, resp.Success)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("BadRequest_LimitOutOfRange", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?limit=200"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("BadRequest_NegativeOffset", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?offset=-5"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("BadRequest_InvalidCountOnly", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?count_only=maybe"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InternalError_RepositoryFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("GetNotifications", mock.Anything, fix.userID, 20, 0).
			Return(nil, 0, errNotificationTest).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, ""))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp struct {
			Success bool      `json:"success"`
			Error   dto.Error `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.False(t, resp.Success)
		assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
	})

	t.Run("Success_BoundaryLimitMin", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("GetNotifications", mock.Anything, fix.userID, 1, 0).
			Return([]dto.Notification{}, 0, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?limit=1"))

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success_BoundaryLimitMax", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("GetNotifications", mock.Anything, fix.userID, 100, 0).
			Return([]dto.Notification{}, 0, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?limit=100"))

		require.Equal(t, http.StatusOK, rr.Code)
	})
}
