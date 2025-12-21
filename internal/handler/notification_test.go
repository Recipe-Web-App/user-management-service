package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
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
