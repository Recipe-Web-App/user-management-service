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
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	errDB                = errors.New("db error")
	errMockArgs          = errors.New("mock: missing args")
	errStartType         = errors.New("invalid type assertion for UserProfileResponse")
	errDeleteRequestType = errors.New("invalid type assertion for UserAccountDeleteRequestResponse")
)

// MockUserService is a mock implementation of service.UserService.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserProfile(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
) (*dto.UserProfileResponse, error) {
	args := m.Called(ctx, requesterID, targetUserID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs // Fix nilnil
	}

	if val, ok := args.Get(0).(*dto.UserProfileResponse); ok {
		return val, nil
	}

	return nil, errStartType // Fix err113
}

func (m *MockUserService) UpdateUserProfile(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.UserProfileUpdateRequest,
) (*dto.UserProfileResponse, error) {
	args := m.Called(ctx, userID, update)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.UserProfileResponse); ok {
		return val, nil
	}

	return nil, errStartType
}

func (m *MockUserService) RequestAccountDeletion(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.UserAccountDeleteRequestResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.UserAccountDeleteRequestResponse); ok {
		return val, nil
	}

	return nil, errDeleteRequestType
}

type userHandlerTestCase struct {
	name           string
	targetIDPath   string
	requesterIDHdr string
	mockRun        func(*MockUserService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

func TestUserHandlerGetUserProfile(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	requesterID := uuid.New()

	baseProfile := &dto.UserProfileResponse{
		UserID:    targetID.String(),
		Username:  "targetuser",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := getHandlerTestCases(targetID, requesterID, baseProfile)

	runUserHandlerTest(t, tests)
}

func getHandlerTestCases(
	targetID, requesterID uuid.UUID,
	baseProfile *dto.UserProfileResponse,
) []userHandlerTestCase {
	return []userHandlerTestCase{
		{
			name:           "Success",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, requesterID, targetID).Return(baseProfile, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, targetID.String())
				assert.Contains(t, body, "targetuser")
			},
		},
		{
			name:         "User Not Found",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, uuid.Nil, targetID).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:         "Profile Private",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, uuid.Nil, targetID).Return(nil, service.ErrProfilePrivate)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:         "Internal Error",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, uuid.Nil, targetID).Return(nil, errDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:         "Invalid ID Format",
			targetIDPath: "invalid-uuid",
			mockRun: func(m *MockUserService) {
				// Service is not called because ID validation fails first.
			},
			expectedStatus: http.StatusBadRequest,
		},
	}
}

func runUserHandlerTest(t *testing.T, tests []userHandlerTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockUserService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewUserHandler(mockSvc)

			r := chi.NewRouter()
			r.Get("/users/{user_id}/profile", h.GetUserProfile)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.targetIDPath+"/profile", nil)
			if tt.requesterIDHdr != "" {
				req.Header.Set("X-User-Id", tt.requesterIDHdr)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.String())
			}
		})
	}
}

type updateProfileTestCase struct {
	name           string
	requesterIDHdr string
	requestBody    string
	contentType    string
	mockRun        func(*MockUserService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

func TestUserHandlerUpdateUserProfile(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	userID := uuid.New()
	now := time.Now()

	tests := []updateProfileTestCase{
		{
			name:           "Success - Update Username",
			requesterIDHdr: userID.String(),
			requestBody:    `{"username": "newusername"}`,
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("UpdateUserProfile", mock.Anything, userID, mock.Anything).Return(&dto.UserProfileResponse{
					UserID:    userID.String(),
					Username:  "newusername",
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "newusername")
				assert.Contains(t, body, `"success":true`)
			},
		},
		{
			name:           "Unauthorized - Missing X-User-Id",
			requesterIDHdr: "",
			requestBody:    `{"username": "newusername"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unauthorized - Invalid UUID in header",
			requesterIDHdr: "not-a-uuid",
			requestBody:    `{"username": "newusername"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Bad Request - Empty body",
			requesterIDHdr: userID.String(),
			requestBody:    "",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "EMPTY_BODY")
			},
		},
		{
			name:           "Bad Request - Invalid JSON",
			requesterIDHdr: userID.String(),
			requestBody:    `{"username": }`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "INVALID_JSON")
			},
		},
		{
			name:           "Bad Request - Validation Error (username too short)",
			requesterIDHdr: userID.String(),
			requestBody:    `{"username": "ab"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "username")
			},
		},
		{
			name:           "Bad Request - Validation Error (invalid username chars)",
			requesterIDHdr: userID.String(),
			requestBody:    `{"username": "user@name"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
			},
		},
		{
			name:           "Not Found",
			requesterIDHdr: userID.String(),
			requestBody:    `{"username": "newusername"}`,
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("UpdateUserProfile", mock.Anything, userID, mock.Anything).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Conflict - Duplicate Username",
			requesterIDHdr: userID.String(),
			requestBody:    `{"username": "existinguser"}`,
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("UpdateUserProfile", mock.Anything, userID, mock.Anything).Return(nil, service.ErrDuplicateUsername)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Internal Error",
			requesterIDHdr: userID.String(),
			requestBody:    `{"username": "newusername"}`,
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("UpdateUserProfile", mock.Anything, userID, mock.Anything).Return(nil, errDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockUserService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewUserHandler(mockSvc)

			r := chi.NewRouter()
			r.Put("/users/profile", h.UpdateUserProfile)

			req := httptest.NewRequest(http.MethodPut, "/users/profile", strings.NewReader(tt.requestBody))
			if tt.requesterIDHdr != "" {
				req.Header.Set("X-User-Id", tt.requesterIDHdr)
			}

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.String())
			}
		})
	}
}

type requestAccountDeletionTestCase struct {
	name           string
	requesterIDHdr string
	mockRun        func(*MockUserService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

func TestUserHandlerRequestAccountDeletion(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	userID := uuid.New()
	now := time.Now()

	tests := []requestAccountDeletionTestCase{
		{
			name:           "Success",
			requesterIDHdr: userID.String(),
			mockRun: func(m *MockUserService) {
				m.On("RequestAccountDeletion", mock.Anything, userID).Return(&dto.UserAccountDeleteRequestResponse{
					UserID:            userID.String(),
					ConfirmationToken: uuid.New().String(),
					ExpiresAt:         now.Add(24 * time.Hour),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, userID.String())
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"confirmationToken"`)
				assert.Contains(t, body, `"expiresAt"`)
			},
		},
		{
			name:           "Unauthorized - Missing X-User-Id",
			requesterIDHdr: "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unauthorized - Invalid UUID in header",
			requesterIDHdr: "not-a-uuid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Not Found - User does not exist",
			requesterIDHdr: userID.String(),
			mockRun: func(m *MockUserService) {
				m.On("RequestAccountDeletion", mock.Anything, userID).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
			},
		},
		{
			name:           "Service Unavailable - Cache unavailable",
			requesterIDHdr: userID.String(),
			mockRun: func(m *MockUserService) {
				m.On("RequestAccountDeletion", mock.Anything, userID).Return(nil, service.ErrCacheUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "SERVICE_UNAVAILABLE")
			},
		},
		{
			name:           "Internal Error",
			requesterIDHdr: userID.String(),
			mockRun: func(m *MockUserService) {
				m.On("RequestAccountDeletion", mock.Anything, userID).Return(nil, errDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockUserService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewUserHandler(mockSvc)

			r := chi.NewRouter()
			r.Post("/users/account/delete-request", h.RequestAccountDeletion)

			req := httptest.NewRequest(http.MethodPost, "/users/account/delete-request", nil)
			if tt.requesterIDHdr != "" {
				req.Header.Set("X-User-Id", tt.requesterIDHdr)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.String())
			}
		})
	}
}
