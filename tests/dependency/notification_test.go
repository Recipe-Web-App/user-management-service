package dependency_test

import (
	"context"
	"encoding/json"
	"errors"
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

func (m *MockNotificationRepository) DeleteNotifications(
	ctx context.Context,
	userID uuid.UUID,
	notificationIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID, notificationIDs)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf("delete notifications: %w", err)
	}

	deletedIDs, _ := args.Get(0).([]uuid.UUID)

	return deletedIDs, nil
}

func (m *MockNotificationRepository) MarkNotificationRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, userID, notificationID)

	err := args.Error(1)
	if err != nil {
		return false, fmt.Errorf("mark notification read: %w", err)
	}

	return args.Bool(0), nil
}

func (m *MockNotificationRepository) MarkAllNotificationsRead(
	ctx context.Context,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf("mark all notifications read: %w", err)
	}

	readIDs, _ := args.Get(0).([]uuid.UUID)

	return readIDs, nil
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

		var resp dto.NotificationListResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, 10, resp.TotalCount)
		assert.Equal(t, 20, resp.Limit)
		assert.Equal(t, 0, resp.Offset)
		require.Len(t, resp.Notifications, 1)
		assert.Equal(t, notification.Title, resp.Notifications[0].Title)
	})

	t.Run("Success_CountOnly", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("CountNotifications", mock.Anything, fix.userID).
			Return(42, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?countOnly=true"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.NotificationCountResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, 42, resp.TotalCount)
	})

	t.Run("Success_Pagination", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("GetNotifications", mock.Anything, fix.userID, 10, 5).
			Return([]dto.Notification{}, 50, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?limit=10&offset=5"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.NotificationListResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, 50, resp.TotalCount)
		assert.Equal(t, 10, resp.Limit)
		assert.Equal(t, 5, resp.Offset)
		assert.Empty(t, resp.Notifications)
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

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "UNAUTHORIZED", resp.Code)
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

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
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
		fix.handler.ServeHTTP(rr, newNotificationRequest(t, fix.userID, "?countOnly=maybe"))

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

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "INTERNAL_ERROR", resp.Code)
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

func newDeleteNotificationRequest(t *testing.T, userID uuid.UUID, body string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		notificationsBaseURL,
		strings.NewReader(body),
	)
	require.NoError(t, err)
	req.Header.Set("X-User-Id", userID.String())
	req.Header.Set("Content-Type", "application/json")

	return req
}

//nolint:funlen // Integration test with many test cases
func TestDeleteNotifications(t *testing.T) {
	t.Parallel()

	t.Run("Success_AllDeleted", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID1 := uuid.New()
		notificationID2 := uuid.New()

		fix.mockRepo.On("DeleteNotifications", mock.Anything, fix.userID, mock.Anything).
			Return([]uuid.UUID{notificationID1, notificationID2}, nil).Once()

		reqBody := fmt.Sprintf(`{"notificationIds": ["%s", "%s"]}`, notificationID1.String(), notificationID2.String())
		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteNotificationRequest(t, fix.userID, reqBody))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.NotificationDeleteResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "Notifications deleted successfully", resp.Message)
		assert.Len(t, resp.DeletedNotificationIDs, 2)
	})

	t.Run("PartialSuccess_SomeDeleted", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID1 := uuid.New()
		notificationID2 := uuid.New()

		// Only first notification was deleted
		fix.mockRepo.On("DeleteNotifications", mock.Anything, fix.userID, mock.Anything).
			Return([]uuid.UUID{notificationID1}, nil).Once()

		reqBody := fmt.Sprintf(`{"notificationIds": ["%s", "%s"]}`, notificationID1.String(), notificationID2.String())
		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteNotificationRequest(t, fix.userID, reqBody))

		require.Equal(t, http.StatusPartialContent, rr.Code)

		var resp dto.NotificationDeleteResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "Some notifications deleted successfully", resp.Message)
		assert.Len(t, resp.DeletedNotificationIDs, 1)
	})

	t.Run("NotFound_NoneDeleted", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID := uuid.New()

		fix.mockRepo.On("DeleteNotifications", mock.Anything, fix.userID, mock.Anything).
			Return([]uuid.UUID{}, nil).Once()

		reqBody := fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID.String())
		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteNotificationRequest(t, fix.userID, reqBody))

		require.Equal(t, http.StatusNotFound, rr.Code)

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "NOT_FOUND", resp.Code)
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID := uuid.New()

		reqBody := fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID.String())
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodDelete,
			notificationsBaseURL,
			strings.NewReader(reqBody),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("BadRequest_EmptyBody", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteNotificationRequest(t, fix.userID, ""))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("BadRequest_InvalidJSON", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteNotificationRequest(t, fix.userID, "{invalid}"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("BadRequest_EmptyNotificationIds", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteNotificationRequest(t, fix.userID, `{"notificationIds": []}`))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InternalError_RepositoryFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID := uuid.New()

		fix.mockRepo.On("DeleteNotifications", mock.Anything, fix.userID, mock.Anything).
			Return(nil, errNotificationTest).Once()

		reqBody := fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID.String())
		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newDeleteNotificationRequest(t, fix.userID, reqBody))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "INTERNAL_ERROR", resp.Code)
	})
}

func newMarkNotificationReadRequest(t *testing.T, userID uuid.UUID, notificationID string) *http.Request {
	t.Helper()

	reqPath := fmt.Sprintf("%s/%s/read", notificationsBaseURL, notificationID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set("X-User-Id", userID.String())

	return req
}

//nolint:funlen // Integration test with many test cases
func TestMarkNotificationRead(t *testing.T) {
	t.Parallel()

	t.Run("Success_MarkedAsRead", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID := uuid.New()

		fix.mockRepo.On("MarkNotificationRead", mock.Anything, fix.userID, notificationID).
			Return(true, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newMarkNotificationReadRequest(t, fix.userID, notificationID.String()))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.NotificationReadResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "Notification marked as read successfully", resp.Message)
	})

	t.Run("NotFound_InvalidId", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID := uuid.New()

		fix.mockRepo.On("MarkNotificationRead", mock.Anything, fix.userID, notificationID).
			Return(false, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newMarkNotificationReadRequest(t, fix.userID, notificationID.String()))

		require.Equal(t, http.StatusNotFound, rr.Code)

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "NOT_FOUND", resp.Code)
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID := uuid.New()

		reqPath := fmt.Sprintf("%s/%s/read", notificationsBaseURL, notificationID.String())
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqPath, nil)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("BadRequest_InvalidUUID", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newMarkNotificationReadRequest(t, fix.userID, "not-a-valid-uuid"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("InternalError_RepositoryFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID := uuid.New()

		fix.mockRepo.On("MarkNotificationRead", mock.Anything, fix.userID, notificationID).
			Return(false, errNotificationTest).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newMarkNotificationReadRequest(t, fix.userID, notificationID.String()))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "INTERNAL_ERROR", resp.Code)
	})
}

func newMarkAllNotificationsReadRequest(t *testing.T, userID uuid.UUID) *http.Request {
	t.Helper()

	reqPath := notificationsBaseURL + "/read-all"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set("X-User-Id", userID.String())

	return req
}

//nolint:funlen // Table-driven test with multiple test cases
func TestMarkAllNotificationsRead(t *testing.T) {
	t.Parallel()

	t.Run("Success_AllMarkedAsRead", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)
		notificationID1 := uuid.New()
		notificationID2 := uuid.New()
		notificationID3 := uuid.New()

		fix.mockRepo.On("MarkAllNotificationsRead", mock.Anything, fix.userID).
			Return([]uuid.UUID{notificationID1, notificationID2, notificationID3}, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newMarkAllNotificationsReadRequest(t, fix.userID))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp dto.NotificationReadAllResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "All notifications marked as read successfully", resp.Message)
		assert.Len(t, resp.ReadNotificationIDs, 3)
		assert.Contains(t, resp.ReadNotificationIDs, notificationID1.String())
		assert.Contains(t, resp.ReadNotificationIDs, notificationID2.String())
		assert.Contains(t, resp.ReadNotificationIDs, notificationID3.String())
	})

	t.Run("Success_EmptyResult", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("MarkAllNotificationsRead", mock.Anything, fix.userID).
			Return([]uuid.UUID{}, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newMarkAllNotificationsReadRequest(t, fix.userID))

		require.Equal(t, http.StatusOK, rr.Code)

		body := rr.Body.String()
		assert.Contains(t, body, `"readNotificationIds":[]`)
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		reqPath := notificationsBaseURL + "/read-all"
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqPath, nil)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "UNAUTHORIZED", resp.Code)
	})

	t.Run("Unauthorized_InvalidUUID", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		reqPath := notificationsBaseURL + "/read-all"
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqPath, nil)
		require.NoError(t, err)
		req.Header.Set("X-User-Id", "not-a-valid-uuid")

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InternalError_RepositoryFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupNotificationTest(t)

		fix.mockRepo.On("MarkAllNotificationsRead", mock.Anything, fix.userID).
			Return(nil, errNotificationTest).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newMarkAllNotificationsReadRequest(t, fix.userID))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp dto.Error
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

		assert.Equal(t, "INTERNAL_ERROR", resp.Code)
	})
}
