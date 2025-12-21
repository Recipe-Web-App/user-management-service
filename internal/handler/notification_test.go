package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

var (
	errMockNotificationArgs = errors.New("mock: missing args")
	errTestDatabase         = errors.New("database error")
)

// MockNotificationService is a mock implementation of service.NotificationService.
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) GetNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	countOnly bool,
) (any, error) {
	args := m.Called(ctx, userID, limit, offset, countOnly)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockNotificationArgs
	}

	return args.Get(0), args.Error(1)
}

func (m *MockNotificationService) DeleteNotifications(
	ctx context.Context,
	userID uuid.UUID,
	notificationIDs []string,
) (*service.NotificationDeleteResult, error) {
	args := m.Called(ctx, userID, notificationIDs)
	if args.Get(0) == nil {
		return nil, fmt.Errorf("mock error: %w", args.Error(1))
	}

	result, _ := args.Get(0).(*service.NotificationDeleteResult)

	err := args.Error(1)
	if err != nil {
		return result, fmt.Errorf("mock error: %w", err)
	}

	return result, nil
}

func (m *MockNotificationService) MarkNotificationRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationID string,
) (bool, error) {
	args := m.Called(ctx, userID, notificationID)

	err := args.Error(1)
	if err != nil {
		return false, fmt.Errorf("mock error: %w", err)
	}

	return args.Bool(0), nil
}

type notificationHandlerTestCase struct {
	name           string
	userIDHeader   string
	queryParams    string
	mockRun        func(*MockNotificationService)
	expectedStatus int
	expectedBody   []string
	notExpected    []string
}

//nolint:funlen // Table-driven test with many test cases
func TestNotificationHandler_GetNotifications(t *testing.T) {
	t.Parallel()

	validUserID := uuid.New()
	notificationID := uuid.New()
	now := time.Now()

	sampleListResponse := &dto.NotificationListResponse{
		Notifications: []dto.Notification{
			{
				NotificationID:   notificationID.String(),
				UserID:           validUserID.String(),
				Title:            "New follower",
				Message:          "John started following you",
				NotificationType: "follow",
				IsRead:           false,
				IsDeleted:        false,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
		TotalCount: 10,
		Limit:      20,
		Offset:     0,
	}

	sampleCountResponse := &dto.NotificationCountResponse{
		TotalCount: 42,
	}

	tests := []notificationHandlerTestCase{
		{
			name:         "returns notifications with default params",
			userIDHeader: validUserID.String(),
			queryParams:  "",
			mockRun: func(m *MockNotificationService) {
				m.On("GetNotifications", mock.Anything, validUserID, 20, 0, false).
					Return(sampleListResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`, `"notifications"`, `"totalCount":10`, `"limit":20`},
		},
		{
			name:         "returns count only when count_only is true",
			userIDHeader: validUserID.String(),
			queryParams:  "?count_only=true",
			mockRun: func(m *MockNotificationService) {
				m.On("GetNotifications", mock.Anything, validUserID, 20, 0, true).
					Return(sampleCountResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`, `"totalCount":42`},
			notExpected:    []string{`"notifications"`, `"limit"`, `"offset"`},
		},
		{
			name:         "respects custom limit and offset",
			userIDHeader: validUserID.String(),
			queryParams:  "?limit=10&offset=5",
			mockRun: func(m *MockNotificationService) {
				m.On("GetNotifications", mock.Anything, validUserID, 10, 5, false).
					Return(sampleListResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`},
		},
		{
			name:           "returns 401 when X-User-Id header is missing",
			userIDHeader:   "",
			queryParams:    "",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{`"success":false`, `"UNAUTHORIZED"`, `"User authentication required"`},
		},
		{
			name:           "returns 401 when X-User-Id header is invalid UUID",
			userIDHeader:   "not-a-uuid",
			queryParams:    "",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{`"success":false`, `"UNAUTHORIZED"`, `"Invalid user ID in authentication header"`},
		},
		{
			name:           "returns 400 when limit is not an integer",
			userIDHeader:   validUserID.String(),
			queryParams:    "?limit=abc",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`, `"limit must be a valid integer"`},
		},
		{
			name:           "returns 400 when limit is below minimum",
			userIDHeader:   validUserID.String(),
			queryParams:    "?limit=0",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`, `"limit must be between 1 and 100"`},
		},
		{
			name:           "returns 400 when limit is above maximum",
			userIDHeader:   validUserID.String(),
			queryParams:    "?limit=101",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`, `"limit must be between 1 and 100"`},
		},
		{
			name:           "returns 400 when offset is not an integer",
			userIDHeader:   validUserID.String(),
			queryParams:    "?offset=abc",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`, `"offset must be a valid integer"`},
		},
		{
			name:           "returns 400 when offset is negative",
			userIDHeader:   validUserID.String(),
			queryParams:    "?offset=-1",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`, `"offset must be non-negative"`},
		},
		{
			name:           "returns 400 when count_only is not a boolean",
			userIDHeader:   validUserID.String(),
			queryParams:    "?count_only=maybe",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`, `"count_only must be a valid boolean"`},
		},
		{
			name:         "returns 500 when service returns error",
			userIDHeader: validUserID.String(),
			queryParams:  "",
			mockRun: func(m *MockNotificationService) {
				m.On("GetNotifications", mock.Anything, validUserID, 20, 0, false).
					Return(nil, errTestDatabase)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{`"success":false`, `"INTERNAL_ERROR"`},
		},
		{
			name:         "accepts limit at minimum boundary",
			userIDHeader: validUserID.String(),
			queryParams:  "?limit=1",
			mockRun: func(m *MockNotificationService) {
				m.On("GetNotifications", mock.Anything, validUserID, 1, 0, false).
					Return(sampleListResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`},
		},
		{
			name:         "accepts limit at maximum boundary",
			userIDHeader: validUserID.String(),
			queryParams:  "?limit=100",
			mockRun: func(m *MockNotificationService) {
				m.On("GetNotifications", mock.Anything, validUserID, 100, 0, false).
					Return(sampleListResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`},
		},
		{
			name:         "accepts offset at zero",
			userIDHeader: validUserID.String(),
			queryParams:  "?offset=0",
			mockRun: func(m *MockNotificationService) {
				m.On("GetNotifications", mock.Anything, validUserID, 20, 0, false).
					Return(sampleListResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockNotificationService)
			tt.mockRun(mockSvc)

			h := handler.NewNotificationHandler(mockSvc)

			r := chi.NewRouter()
			r.Get("/notifications", h.GetNotifications)

			req := httptest.NewRequest(http.MethodGet, "/notifications"+tt.queryParams, nil)
			if tt.userIDHeader != "" {
				req.Header.Set("X-User-Id", tt.userIDHeader)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			body := rr.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}

			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, body, notExpected)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

//nolint:funlen // Table-driven test with many test cases
func TestNotificationHandler_DeleteNotifications(t *testing.T) {
	t.Parallel()

	validUserID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()

	tests := []struct {
		name           string
		userIDHeader   string
		requestBody    string
		mockRun        func(*MockNotificationService)
		expectedStatus int
		expectedBody   []string
		notExpected    []string
	}{
		{
			name:         "returns 200 when all notifications deleted",
			userIDHeader: validUserID.String(),
			requestBody:  fmt.Sprintf(`{"notificationIds": ["%s", "%s"]}`, notificationID1.String(), notificationID2.String()),
			mockRun: func(m *MockNotificationService) {
				m.On("DeleteNotifications", mock.Anything, validUserID, mock.Anything).
					Return(&service.NotificationDeleteResult{
						DeletedIDs:   []string{notificationID1.String(), notificationID2.String()},
						RequestedIDs: []string{notificationID1.String(), notificationID2.String()},
						IsPartial:    false,
						AllNotFound:  false,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`, `"message":"Notifications deleted successfully"`},
		},
		{
			name:         "returns 206 when partial success",
			userIDHeader: validUserID.String(),
			requestBody:  fmt.Sprintf(`{"notificationIds": ["%s", "%s"]}`, notificationID1.String(), notificationID2.String()),
			mockRun: func(m *MockNotificationService) {
				m.On("DeleteNotifications", mock.Anything, validUserID, mock.Anything).
					Return(&service.NotificationDeleteResult{
						DeletedIDs:   []string{notificationID1.String()},
						RequestedIDs: []string{notificationID1.String(), notificationID2.String()},
						IsPartial:    true,
						AllNotFound:  false,
					}, nil)
			},
			expectedStatus: http.StatusPartialContent,
			expectedBody:   []string{`"success":true`, `"message":"Some notifications deleted successfully"`},
		},
		{
			name:         "returns 404 when no notifications found",
			userIDHeader: validUserID.String(),
			requestBody:  fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID1.String()),
			mockRun: func(m *MockNotificationService) {
				m.On("DeleteNotifications", mock.Anything, validUserID, mock.Anything).
					Return(&service.NotificationDeleteResult{
						DeletedIDs:   []string{},
						RequestedIDs: []string{notificationID1.String()},
						IsPartial:    false,
						AllNotFound:  true,
					}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{`"success":false`, `"NOT_FOUND"`},
		},
		{
			name:           "returns 401 when X-User-Id header is missing",
			userIDHeader:   "",
			requestBody:    fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID1.String()),
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{`"success":false`, `"UNAUTHORIZED"`, `"User authentication required"`},
		},
		{
			name:           "returns 401 when X-User-Id header is invalid UUID",
			userIDHeader:   "not-a-uuid",
			requestBody:    fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID1.String()),
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{`"success":false`, `"UNAUTHORIZED"`, `"Invalid user ID in authentication header"`},
		},
		{
			name:           "returns 400 when request body is empty",
			userIDHeader:   validUserID.String(),
			requestBody:    "",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"EMPTY_BODY"`},
		},
		{
			name:           "returns 400 when request body is invalid JSON",
			userIDHeader:   validUserID.String(),
			requestBody:    `{invalid json}`,
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"INVALID_JSON"`},
		},
		{
			name:           "returns 400 when notificationIds is empty array",
			userIDHeader:   validUserID.String(),
			requestBody:    `{"notificationIds": []}`,
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`},
		},
		{
			name:           "returns 400 when notificationIds contains invalid UUID",
			userIDHeader:   validUserID.String(),
			requestBody:    `{"notificationIds": ["not-a-uuid"]}`,
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`},
		},
		{
			name:         "returns 500 when service returns error",
			userIDHeader: validUserID.String(),
			requestBody:  fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID1.String()),
			mockRun: func(m *MockNotificationService) {
				m.On("DeleteNotifications", mock.Anything, validUserID, mock.Anything).
					Return(nil, errTestDatabase)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{`"success":false`, `"INTERNAL_ERROR"`},
		},
		{
			name:         "handles single notification deletion",
			userIDHeader: validUserID.String(),
			requestBody:  fmt.Sprintf(`{"notificationIds": ["%s"]}`, notificationID1.String()),
			mockRun: func(m *MockNotificationService) {
				m.On("DeleteNotifications", mock.Anything, validUserID, mock.Anything).
					Return(&service.NotificationDeleteResult{
						DeletedIDs:   []string{notificationID1.String()},
						RequestedIDs: []string{notificationID1.String()},
						IsPartial:    false,
						AllNotFound:  false,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`, `"message":"Notifications deleted successfully"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockNotificationService)
			tt.mockRun(mockSvc)

			h := handler.NewNotificationHandler(mockSvc)

			r := chi.NewRouter()
			r.Delete("/notifications", h.DeleteNotifications)

			var body *strings.Reader
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			} else {
				body = strings.NewReader("")
			}

			req := httptest.NewRequest(http.MethodDelete, "/notifications", body)
			req.Header.Set("Content-Type", "application/json")

			if tt.userIDHeader != "" {
				req.Header.Set("X-User-Id", tt.userIDHeader)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			respBody := rr.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, respBody, expected)
			}

			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, respBody, notExpected)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

//nolint:funlen // Table-driven test with many test cases
func TestNotificationHandler_MarkNotificationRead(t *testing.T) {
	t.Parallel()

	validUserID := uuid.New()
	notificationID := uuid.New()

	tests := []struct {
		name           string
		userIDHeader   string
		notificationID string
		mockRun        func(*MockNotificationService)
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "returns 200 when notification marked as read",
			userIDHeader:   validUserID.String(),
			notificationID: notificationID.String(),
			mockRun: func(m *MockNotificationService) {
				m.On("MarkNotificationRead", mock.Anything, validUserID, notificationID.String()).
					Return(true, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{`"success":true`, `"message":"Notification marked as read successfully"`},
		},
		{
			name:           "returns 404 when notification not found",
			userIDHeader:   validUserID.String(),
			notificationID: notificationID.String(),
			mockRun: func(m *MockNotificationService) {
				m.On("MarkNotificationRead", mock.Anything, validUserID, notificationID.String()).
					Return(false, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{`"success":false`, `"NOT_FOUND"`},
		},
		{
			name:           "returns 401 when X-User-Id header is missing",
			userIDHeader:   "",
			notificationID: notificationID.String(),
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{`"success":false`, `"UNAUTHORIZED"`, `"User authentication required"`},
		},
		{
			name:           "returns 401 when X-User-Id header is invalid UUID",
			userIDHeader:   "not-a-uuid",
			notificationID: notificationID.String(),
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{`"success":false`, `"UNAUTHORIZED"`, `"Invalid user ID in authentication header"`},
		},
		{
			name:           "returns 400 when notification_id is invalid UUID",
			userIDHeader:   validUserID.String(),
			notificationID: "not-a-uuid",
			mockRun:        func(_ *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{`"success":false`, `"VALIDATION_ERROR"`, `"notification_id must be a valid UUID"`},
		},
		{
			name:           "returns 500 when service returns error",
			userIDHeader:   validUserID.String(),
			notificationID: notificationID.String(),
			mockRun: func(m *MockNotificationService) {
				m.On("MarkNotificationRead", mock.Anything, validUserID, notificationID.String()).
					Return(false, errTestDatabase)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{`"success":false`, `"INTERNAL_ERROR"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockNotificationService)
			tt.mockRun(mockSvc)

			h := handler.NewNotificationHandler(mockSvc)

			r := chi.NewRouter()
			r.Put("/notifications/{notification_id}/read", h.MarkNotificationRead)

			reqPath := fmt.Sprintf("/notifications/%s/read", tt.notificationID)
			req := httptest.NewRequest(http.MethodPut, reqPath, nil)

			if tt.userIDHeader != "" {
				req.Header.Set("X-User-Id", tt.userIDHeader)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			body := rr.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
