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
	errDB                  = errors.New("db error")
	errMockArgs            = errors.New("mock: missing args")
	errStartType           = errors.New("invalid type assertion for UserProfileResponse")
	errDeleteRequestType   = errors.New("invalid type assertion for UserAccountDeleteRequestResponse")
	errConfirmDeletionType = errors.New("invalid type assertion for UserConfirmAccountDeleteResponse")
	errSearchResponseType  = errors.New("invalid type assertion for UserSearchResponse")
	errSearchResultType    = errors.New("invalid type assertion for UserSearchResult")
	errUserStatsType       = errors.New("invalid type assertion for UserStatsResponse")
	internalErrorStr       = "Internal Error"
	userIDHeaderStr        = "X-User-Id"
	userNotFoundStr        = "Not Found - User does not exist"
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

func (m *MockUserService) ConfirmAccountDeletion(
	ctx context.Context,
	userID uuid.UUID,
	token string,
) (*dto.UserConfirmAccountDeleteResponse, error) {
	args := m.Called(ctx, userID, token)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.UserConfirmAccountDeleteResponse); ok {
		return val, nil
	}

	return nil, errConfirmDeletionType
}

func (m *MockUserService) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
	countOnly bool,
) (*dto.UserSearchResponse, error) {
	args := m.Called(ctx, query, limit, offset, countOnly)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.UserSearchResponse); ok {
		return val, nil
	}

	return nil, errSearchResponseType
}

func (m *MockUserService) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.UserSearchResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.UserSearchResult); ok {
		return val, nil
	}

	return nil, errSearchResultType
}

func (m *MockUserService) GetUserStats(ctx context.Context) (*dto.UserStatsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockArgs
	}

	if val, ok := args.Get(0).(*dto.UserStatsResponse); ok {
		return val, nil
	}

	return nil, errUserStatsType
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
			name:         internalErrorStr,
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
				req.Header.Set(userIDHeaderStr, tt.requesterIDHdr)
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
			name:           internalErrorStr,
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
				req.Header.Set(userIDHeaderStr, tt.requesterIDHdr)
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
			name:           userNotFoundStr,
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
			name:           internalErrorStr,
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
				req.Header.Set(userIDHeaderStr, tt.requesterIDHdr)
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

type confirmAccountDeletionTestCase struct {
	name           string
	requesterIDHdr string
	requestBody    string
	contentType    string
	mockRun        func(*MockUserService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

func TestUserHandlerConfirmAccountDeletion(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	userID := uuid.New()
	token := uuid.New().String()
	now := time.Now()

	tests := []confirmAccountDeletionTestCase{
		{
			name:           "Success",
			requesterIDHdr: userID.String(),
			requestBody:    fmt.Sprintf(`{"confirmationToken": "%s"}`, token),
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("ConfirmAccountDeletion", mock.Anything, userID, token).Return(&dto.UserConfirmAccountDeleteResponse{
					UserID:        userID.String(),
					DeactivatedAt: now,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, userID.String())
				assert.Contains(t, body, `"deactivatedAt"`)
			},
		},
		{
			name:           "Unauthorized - Missing X-User-Id",
			requesterIDHdr: "",
			requestBody:    fmt.Sprintf(`{"confirmationToken": "%s"}`, token),
			contentType:    "application/json",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unauthorized - Invalid UUID in header",
			requesterIDHdr: "not-a-uuid",
			requestBody:    fmt.Sprintf(`{"confirmationToken": "%s"}`, token),
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
			name:           "Bad Request - Missing token",
			requesterIDHdr: userID.String(),
			requestBody:    `{}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
			},
		},
		{
			name:           "Bad Request - Invalid token",
			requesterIDHdr: userID.String(),
			requestBody:    `{"confirmationToken": "wrong-token"}`,
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("ConfirmAccountDeletion", mock.Anything, userID, "wrong-token").Return(nil, service.ErrInvalidToken)
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "INVALID_TOKEN")
			},
		},
		{
			name:           userNotFoundStr,
			requesterIDHdr: userID.String(),
			requestBody:    fmt.Sprintf(`{"confirmationToken": "%s"}`, token),
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("ConfirmAccountDeletion", mock.Anything, userID, token).Return(nil, service.ErrUserNotFound)
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
			requestBody:    fmt.Sprintf(`{"confirmationToken": "%s"}`, token),
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("ConfirmAccountDeletion", mock.Anything, userID, token).Return(nil, service.ErrCacheUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "SERVICE_UNAVAILABLE")
			},
		},
		{
			name:           internalErrorStr,
			requesterIDHdr: userID.String(),
			requestBody:    fmt.Sprintf(`{"confirmationToken": "%s"}`, token),
			contentType:    "application/json",
			mockRun: func(m *MockUserService) {
				m.On("ConfirmAccountDeletion", mock.Anything, userID, token).Return(nil, errDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests { //nolint:dupl // table-driven test runner
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockUserService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewUserHandler(mockSvc)

			r := chi.NewRouter()
			r.Delete("/users/account", h.ConfirmAccountDeletion)

			req := httptest.NewRequest(http.MethodDelete, "/users/account", strings.NewReader(tt.requestBody))
			if tt.requesterIDHdr != "" {
				req.Header.Set(userIDHeaderStr, tt.requesterIDHdr)
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

type searchUsersTestCase struct {
	name           string
	requesterIDHdr string
	queryParams    string
	mockRun        func(*MockUserService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

func TestUserHandlerSearchUsers(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	userID := uuid.New()
	now := time.Now()
	fullName := "Test User"

	tests := []searchUsersTestCase{
		{
			name:           "Success - Returns search results",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test&limit=10&offset=0",
			mockRun: func(m *MockUserService) {
				m.On("SearchUsers", mock.Anything, "test", 10, 0, false).Return(&dto.UserSearchResponse{
					Results: []dto.UserSearchResult{
						{
							UserID:    uuid.New().String(),
							Username:  "testuser",
							FullName:  &fullName,
							IsActive:  true,
							CreatedAt: now,
							UpdatedAt: now,
						},
					},
					TotalCount: 1,
					Limit:      10,
					Offset:     0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "testuser")
				assert.Contains(t, body, `"totalCount":1`)
			},
		},
		{
			name:           "Success - countOnly returns only count",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test&countOnly=true",
			mockRun: func(m *MockUserService) {
				m.On("SearchUsers", mock.Anything, "test", 20, 0, true).Return(&dto.UserSearchResponse{
					Results:    []dto.UserSearchResult{},
					TotalCount: 5,
					Limit:      20,
					Offset:     0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"totalCount":5`)
				assert.Contains(t, body, `"results":[]`)
			},
		},
		{
			name:           "Success - Empty query returns all users",
			requesterIDHdr: userID.String(),
			queryParams:    "",
			mockRun: func(m *MockUserService) {
				m.On("SearchUsers", mock.Anything, "", 20, 0, false).Return(&dto.UserSearchResponse{
					Results:    []dto.UserSearchResult{},
					TotalCount: 0,
					Limit:      20,
					Offset:     0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"totalCount":0`)
			},
		},
		{
			name:           "Success - No results found (empty array)",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=nonexistent",
			mockRun: func(m *MockUserService) {
				m.On("SearchUsers", mock.Anything, "nonexistent", 20, 0, false).Return(&dto.UserSearchResponse{
					Results:    []dto.UserSearchResult{},
					TotalCount: 0,
					Limit:      20,
					Offset:     0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"results":[]`)
				assert.Contains(t, body, `"totalCount":0`)
			},
		},
		{
			name:           "Unauthorized - Missing X-User-Id",
			requesterIDHdr: "",
			queryParams:    "?query=test",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unauthorized - Invalid UUID in header",
			requesterIDHdr: "not-a-uuid",
			queryParams:    "?query=test",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Bad Request - Invalid limit (zero)",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test&limit=0",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit")
			},
		},
		{
			name:           "Bad Request - Invalid limit (over max)",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test&limit=101",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit")
			},
		},
		{
			name:           "Bad Request - Invalid limit (not a number)",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test&limit=abc",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit")
			},
		},
		{
			name:           "Bad Request - Negative offset",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test&offset=-1",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "offset")
			},
		},
		{
			name:           "Bad Request - Invalid countOnly",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test&countOnly=notabool",
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "countOnly")
			},
		},
		{
			name:           "Internal Error - Database failure",
			requesterIDHdr: userID.String(),
			queryParams:    "?query=test",
			mockRun: func(m *MockUserService) {
				m.On("SearchUsers", mock.Anything, "test", 20, 0, false).Return(nil, errDB)
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
			r.Get("/users/search", h.SearchUsers)

			req := httptest.NewRequest(http.MethodGet, "/users/search"+tt.queryParams, nil)
			if tt.requesterIDHdr != "" {
				req.Header.Set(userIDHeaderStr, tt.requesterIDHdr)
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

type getUserByIDTestCase struct {
	name           string
	targetIDPath   string
	mockRun        func(*MockUserService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

func TestUserHandlerGetUserByID(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	targetID := uuid.New()
	now := time.Now()
	fullName := "Test User"

	baseResult := &dto.UserSearchResult{
		UserID:    targetID.String(),
		Username:  "testuser",
		FullName:  &fullName,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []getUserByIDTestCase{
		{
			name:         "Success - Public profile",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, targetID).Return(baseResult, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"userId"`)
				assert.Contains(t, body, `"username"`)
			},
		},
		{
			name:         userNotFoundStr,
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, targetID).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
			},
		},
		{
			name:         "Not Found - Private profile returns 404",
			targetIDPath: uuid.New().String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, mock.Anything).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
			},
		},
		{
			name:           "Validation Error - Invalid UUID format",
			targetIDPath:   "invalid-uuid",
			mockRun:        func(_ *MockUserService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
			},
		},
		{
			name:         "Internal Error - Database failure",
			targetIDPath: targetID.String(),
			mockRun: func(m *MockUserService) {
				m.On("GetUserByID", mock.Anything, targetID).Return(nil, errDB)
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
			r.Get("/users/{user_id}", h.GetUserByID)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.targetIDPath, nil)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.String())
			}
		})
	}
}
