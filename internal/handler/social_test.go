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
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	errMockSocialArgs        = errors.New("mock: missing args")
	errFollowedUsersRespType = errors.New("invalid type assertion for GetFollowedUsersResponse")
	errFollowRespType        = errors.New("invalid type assertion for FollowResponse")
	errUserActivityRespType  = errors.New("invalid type assertion for UserActivityResponse")
	errUnexpectedService     = errors.New("unexpected service error")
)

// MockSocialService is a mock implementation of service.SocialService.
type MockSocialService struct {
	mock.Mock
}

func (m *MockSocialService) GetFollowing(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	limit, offset int,
	countOnly bool,
) (*dto.GetFollowedUsersResponse, error) {
	args := m.Called(ctx, requesterID, targetUserID, limit, offset, countOnly)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.GetFollowedUsersResponse); ok {
		return val, nil
	}

	return nil, errFollowedUsersRespType
}

func (m *MockSocialService) GetFollowers(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	limit, offset int,
	countOnly bool,
) (*dto.GetFollowedUsersResponse, error) {
	args := m.Called(ctx, requesterID, targetUserID, limit, offset, countOnly)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.GetFollowedUsersResponse); ok {
		return val, nil
	}

	return nil, errFollowedUsersRespType
}

func (m *MockSocialService) FollowUser(
	ctx context.Context,
	followerID, targetUserID uuid.UUID,
) (*dto.FollowResponse, error) {
	args := m.Called(ctx, followerID, targetUserID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.FollowResponse); ok {
		return val, nil
	}

	return nil, errFollowRespType
}

func (m *MockSocialService) UnfollowUser(
	ctx context.Context,
	followerID, targetUserID uuid.UUID,
) (*dto.FollowResponse, error) {
	args := m.Called(ctx, followerID, targetUserID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.FollowResponse); ok {
		return val, nil
	}

	return nil, errFollowRespType
}

func (m *MockSocialService) GetUserActivity(
	ctx context.Context,
	requesterID *uuid.UUID,
	targetUserID uuid.UUID,
	perTypeLimit int,
) (*dto.UserActivityResponse, error) {
	args := m.Called(ctx, requesterID, targetUserID, perTypeLimit)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.UserActivityResponse); ok {
		return val, nil
	}

	return nil, errUserActivityRespType
}

type socialHandlerTestCase struct {
	name           string
	targetIDPath   string
	requesterIDHdr string
	queryParams    string
	mockRun        func(*MockSocialService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

//nolint:funlen,maintidx // table-driven test with many test cases
func TestSocialHandlerGetFollowing(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	requesterID := uuid.New()

	now := time.Now()
	limit := 20
	offset := 0
	fullName := "Jane Smith"

	baseResponse := &dto.GetFollowedUsersResponse{
		TotalCount: 1,
		FollowedUsers: []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "janesmith",
				FullName:  &fullName,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Limit:  &limit,
		Offset: &offset,
	}

	countOnlyResponse := &dto.GetFollowedUsersResponse{
		TotalCount: 42,
	}

	emptyResponse := &dto.GetFollowedUsersResponse{
		TotalCount:    0,
		FollowedUsers: []dto.User{},
		Limit:         &limit,
		Offset:        &offset,
	}

	tests := []socialHandlerTestCase{
		{
			name:           "Success - returns following list with pagination",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, targetID, 20, 0, false).Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"totalCount":1`)
				assert.Contains(t, body, `"janesmith"`)
			},
		},
		{
			name:           "Success - count_only returns only totalCount",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "count_only=true",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, targetID, 20, 0, true).Return(countOnlyResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"totalCount":42`)
			},
		},
		{
			name:           "Success - empty following list",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, targetID, 20, 0, false).Return(emptyResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"totalCount":0`)
				// Empty followedUsers is omitted due to omitempty tag
				assert.NotContains(t, body, `"followedUsers"`)
			},
		},
		{
			name:           "Success - custom pagination",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=50&offset=10",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, targetID, 50, 10, false).Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
			},
		},
		{
			name:           "Success - viewing own following list",
			targetIDPath:   requesterID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, requesterID, 20, 0, false).Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
			},
		},
		{
			name:           "Unauthorized - missing X-User-Id header",
			targetIDPath:   targetID.String(),
			requesterIDHdr: "",
			queryParams:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
				assert.Contains(t, body, "User authentication required")
			},
		},
		{
			name:           "Unauthorized - invalid X-User-Id header",
			targetIDPath:   targetID.String(),
			requesterIDHdr: "invalid-uuid",
			queryParams:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
				assert.Contains(t, body, "Invalid user ID in authentication header")
			},
		},
		{
			name:           "Validation Error - invalid UUID format",
			targetIDPath:   "invalid-uuid",
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Invalid user ID format")
			},
		},
		{
			name:           "Validation Error - invalid limit",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=abc",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit must be a valid integer")
			},
		},
		{
			name:           "Validation Error - limit out of range (too low)",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=0",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit must be between 1 and 100")
			},
		},
		{
			name:           "Validation Error - limit out of range (too high)",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=101",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit must be between 1 and 100")
			},
		},
		{
			name:           "Validation Error - invalid offset",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "offset=abc",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "offset must be a valid integer")
			},
		},
		{
			name:           "Validation Error - negative offset",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "offset=-1",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "offset must be non-negative")
			},
		},
		{
			name:           "Validation Error - invalid count_only",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "count_only=maybe",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "count_only must be a valid boolean")
			},
		},
		{
			name:           "Not Found - user does not exist",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, targetID, 20, 0, false).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
				assert.Contains(t, body, "User not found")
			},
		},
		{
			name:           "Forbidden - private profile",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, targetID, 20, 0, false).Return(nil, service.ErrAccessDenied)
			},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "FORBIDDEN")
				assert.Contains(t, body, "Access to this user's following list is restricted")
			},
		},
		{
			name:           "Internal Error - service error",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowing", mock.Anything, requesterID, targetID, 20, 0, false).Return(nil, errUnexpectedService)
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "INTERNAL_ERROR")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockSocialService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewSocialHandler(mockSvc)

			r := chi.NewRouter()
			r.Get("/users/{user_id}/following", h.GetFollowing)

			url := "/users/" + tt.targetIDPath + "/following"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
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

//nolint:funlen,maintidx,dupl // table-driven test with many test cases
func TestSocialHandlerGetFollowers(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	requesterID := uuid.New()

	now := time.Now()
	limit := 20
	offset := 0
	fullName := "Jane Smith"

	baseResponse := &dto.GetFollowedUsersResponse{
		TotalCount: 1,
		FollowedUsers: []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "janesmith",
				FullName:  &fullName,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Limit:  &limit,
		Offset: &offset,
	}

	countOnlyResponse := &dto.GetFollowedUsersResponse{
		TotalCount: 42,
	}

	emptyResponse := &dto.GetFollowedUsersResponse{
		TotalCount:    0,
		FollowedUsers: []dto.User{},
		Limit:         &limit,
		Offset:        &offset,
	}

	tests := []socialHandlerTestCase{
		{
			name:           "Success - returns followers list with pagination",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, targetID, 20, 0, false).Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"totalCount":1`)
				assert.Contains(t, body, `"janesmith"`)
			},
		},
		{
			name:           "Success - count_only returns only totalCount",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "count_only=true",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, targetID, 20, 0, true).Return(countOnlyResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"totalCount":42`)
			},
		},
		{
			name:           "Success - empty followers list",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, targetID, 20, 0, false).Return(emptyResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"totalCount":0`)
				// Empty followedUsers is omitted due to omitempty tag
				assert.NotContains(t, body, `"followedUsers"`)
			},
		},
		{
			name:           "Success - custom pagination",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=50&offset=10",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, targetID, 50, 10, false).Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
			},
		},
		{
			name:           "Success - viewing own followers list",
			targetIDPath:   requesterID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, requesterID, 20, 0, false).Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
			},
		},
		{
			name:           "Unauthorized - missing X-User-Id header",
			targetIDPath:   targetID.String(),
			requesterIDHdr: "",
			queryParams:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
				assert.Contains(t, body, "User authentication required")
			},
		},
		{
			name:           "Unauthorized - invalid X-User-Id header",
			targetIDPath:   targetID.String(),
			requesterIDHdr: "invalid-uuid",
			queryParams:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
				assert.Contains(t, body, "Invalid user ID in authentication header")
			},
		},
		{
			name:           "Validation Error - invalid UUID format",
			targetIDPath:   "invalid-uuid",
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Invalid user ID format")
			},
		},
		{
			name:           "Validation Error - invalid limit",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=abc",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit must be a valid integer")
			},
		},
		{
			name:           "Validation Error - limit out of range (too low)",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=0",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit must be between 1 and 100")
			},
		},
		{
			name:           "Validation Error - limit out of range (too high)",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "limit=101",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "limit must be between 1 and 100")
			},
		},
		{
			name:           "Validation Error - invalid offset",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "offset=abc",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "offset must be a valid integer")
			},
		},
		{
			name:           "Validation Error - negative offset",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "offset=-1",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "offset must be non-negative")
			},
		},
		{
			name:           "Validation Error - invalid count_only",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "count_only=maybe",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "count_only must be a valid boolean")
			},
		},
		{
			name:           "Not Found - user does not exist",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, targetID, 20, 0, false).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
				assert.Contains(t, body, "User not found")
			},
		},
		{
			name:           "Forbidden - private profile",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, targetID, 20, 0, false).Return(nil, service.ErrAccessDenied)
			},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "FORBIDDEN")
				assert.Contains(t, body, "Access to this user's followers list is restricted")
			},
		},
		{
			name:           "Internal Error - service error",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetFollowers", mock.Anything, requesterID, targetID, 20, 0, false).Return(nil, errUnexpectedService)
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "INTERNAL_ERROR")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockSocialService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewSocialHandler(mockSvc)

			r := chi.NewRouter()
			r.Get("/users/{user_id}/followers", h.GetFollowers)

			url := "/users/" + tt.targetIDPath + "/followers"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
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

type followUserTestCase struct {
	name           string
	userIDPath     string
	targetIDPath   string
	requesterIDHdr string
	userRoleHdr    string
	mockRun        func(*MockSocialService)
	expectedStatus int
	validateBody   func(*testing.T, string)
}

//nolint:funlen,dupl // table-driven test with many test cases, mirrors UnfollowUser pattern
func TestSocialHandlerFollowUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	targetID := uuid.New()
	differentUserID := uuid.New()

	successResponse := &dto.FollowResponse{
		Message:     "Successfully followed user",
		IsFollowing: true,
	}

	tests := []followUserTestCase{
		{
			name:           "Success - follow user (user_id matches requester)",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("FollowUser", mock.Anything, userID, targetID).Return(successResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"isFollowing":true`)
				assert.Contains(t, body, `"Successfully followed user"`)
			},
		},
		{
			name:           "Success - admin follows on behalf of another user",
			userIDPath:     differentUserID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "admin",
			mockRun: func(m *MockSocialService) {
				m.On("FollowUser", mock.Anything, differentUserID, targetID).Return(successResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"isFollowing":true`)
			},
		},
		{
			name:           "Unauthorized - missing X-User-Id header",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: "",
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
			},
		},
		{
			name:           "Unauthorized - invalid X-User-Id header",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: "invalid-uuid",
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
			},
		},
		{
			name:           "Validation Error - invalid user_id format",
			userIDPath:     "invalid-uuid",
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Invalid user ID format")
			},
		},
		{
			name:           "Validation Error - invalid target_user_id format",
			userIDPath:     userID.String(),
			targetIDPath:   "invalid-uuid",
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Invalid target user ID format")
			},
		},
		{
			name:           "Forbidden - user_id does not match authenticated user (non-admin)",
			userIDPath:     differentUserID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "FORBIDDEN")
				assert.Contains(t, body, "Cannot perform follow action for another user")
			},
		},
		{
			name:           "Bad Request - cannot follow self",
			userIDPath:     userID.String(),
			targetIDPath:   userID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("FollowUser", mock.Anything, userID, userID).Return(nil, service.ErrCannotFollowSelf)
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Cannot follow yourself")
			},
		},
		{
			name:           "Not Found - target user does not exist",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("FollowUser", mock.Anything, userID, targetID).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
			},
		},
		{
			name:           "Forbidden - target user does not allow follows",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("FollowUser", mock.Anything, userID, targetID).Return(nil, service.ErrFollowNotAllowed)
			},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "FORBIDDEN")
				assert.Contains(t, body, "does not allow follows")
			},
		},
		{
			name:           "Internal Error - service error",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("FollowUser", mock.Anything, userID, targetID).Return(nil, errUnexpectedService)
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "INTERNAL_ERROR")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockSocialService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewSocialHandler(mockSvc)

			r := chi.NewRouter()
			r.Post("/users/{user_id}/follow/{target_user_id}", h.FollowUser)

			url := "/users/" + tt.userIDPath + "/follow/" + tt.targetIDPath

			req := httptest.NewRequest(http.MethodPost, url, nil)
			if tt.requesterIDHdr != "" {
				req.Header.Set("X-User-Id", tt.requesterIDHdr)
			}

			if tt.userRoleHdr != "" {
				req.Header.Set("X-User-Role", tt.userRoleHdr)
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

//nolint:funlen,dupl // table-driven test with many test cases, mirrors FollowUser pattern
func TestSocialHandlerUnfollowUser(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	targetID := uuid.New()
	differentUserID := uuid.New()

	successResponse := &dto.FollowResponse{
		Message:     "Successfully unfollowed user",
		IsFollowing: false,
	}

	tests := []followUserTestCase{
		{
			name:           "Success - unfollow user (user_id matches requester)",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("UnfollowUser", mock.Anything, userID, targetID).Return(successResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"isFollowing":false`)
				assert.Contains(t, body, `"Successfully unfollowed user"`)
			},
		},
		{
			name:           "Success - admin unfollows on behalf of another user",
			userIDPath:     differentUserID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "admin",
			mockRun: func(m *MockSocialService) {
				m.On("UnfollowUser", mock.Anything, differentUserID, targetID).Return(successResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"isFollowing":false`)
			},
		},
		{
			name:           "Unauthorized - missing X-User-Id header",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: "",
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
			},
		},
		{
			name:           "Unauthorized - invalid X-User-Id header",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: "invalid-uuid",
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnauthorized,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "UNAUTHORIZED")
			},
		},
		{
			name:           "Validation Error - invalid user_id format",
			userIDPath:     "invalid-uuid",
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Invalid user ID format")
			},
		},
		{
			name:           "Validation Error - invalid target_user_id format",
			userIDPath:     userID.String(),
			targetIDPath:   "invalid-uuid",
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Invalid target user ID format")
			},
		},
		{
			name:           "Forbidden - user_id does not match authenticated user (non-admin)",
			userIDPath:     differentUserID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "FORBIDDEN")
				assert.Contains(t, body, "Cannot perform unfollow action for another user")
			},
		},
		{
			name:           "Bad Request - cannot unfollow self",
			userIDPath:     userID.String(),
			targetIDPath:   userID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("UnfollowUser", mock.Anything, userID, userID).Return(nil, service.ErrCannotUnfollowSelf)
			},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Cannot unfollow yourself")
			},
		},
		{
			name:           "Not Found - target user does not exist",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("UnfollowUser", mock.Anything, userID, targetID).Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
			},
		},
		{
			name:           "Internal Error - service error",
			userIDPath:     userID.String(),
			targetIDPath:   targetID.String(),
			requesterIDHdr: userID.String(),
			userRoleHdr:    "",
			mockRun: func(m *MockSocialService) {
				m.On("UnfollowUser", mock.Anything, userID, targetID).Return(nil, errUnexpectedService)
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "INTERNAL_ERROR")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockSocialService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewSocialHandler(mockSvc)

			r := chi.NewRouter()
			r.Delete("/users/{user_id}/follow/{target_user_id}", h.UnfollowUser)

			url := "/users/" + tt.userIDPath + "/follow/" + tt.targetIDPath

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			if tt.requesterIDHdr != "" {
				req.Header.Set("X-User-Id", tt.requesterIDHdr)
			}

			if tt.userRoleHdr != "" {
				req.Header.Set("X-User-Role", tt.userRoleHdr)
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

//nolint:funlen,maintidx,dupl // table-driven test with many test cases
func TestSocialHandlerGetUserActivity(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	requesterID := uuid.New()

	now := time.Now()

	baseResponse := &dto.UserActivityResponse{
		UserID: targetID.String(),
		RecentRecipes: []dto.RecipeSummary{
			{RecipeID: 1, Title: "Test Recipe", CreatedAt: now},
		},
		RecentFollows: []dto.UserSummary{
			{UserID: uuid.New().String(), Username: "testuser", FollowedAt: now},
		},
		RecentReviews:   []dto.ReviewSummary{},
		RecentFavorites: []dto.FavoriteSummary{},
	}

	emptyResponse := &dto.UserActivityResponse{
		UserID:          targetID.String(),
		RecentRecipes:   []dto.RecipeSummary{},
		RecentFollows:   []dto.UserSummary{},
		RecentReviews:   []dto.ReviewSummary{},
		RecentFavorites: []dto.FavoriteSummary{},
	}

	tests := []socialHandlerTestCase{
		{
			name:           "Success - authenticated user views public profile",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 15).
					Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"userId"`)
				assert.Contains(t, body, `"recentRecipes"`)
				assert.Contains(t, body, `"Test Recipe"`)
			},
		},
		{
			name:           "Success - anonymous user views public profile",
			targetIDPath:   targetID.String(),
			requesterIDHdr: "",
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, (*uuid.UUID)(nil), targetID, 15).
					Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"recentRecipes"`)
			},
		},
		{
			name:           "Success - custom per_type_limit",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "per_type_limit=50",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 50).
					Return(baseResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
			},
		},
		{
			name:           "Success - minimum per_type_limit",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "per_type_limit=1",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 1).
					Return(emptyResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - maximum per_type_limit",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "per_type_limit=100",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 100).
					Return(emptyResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - empty activity",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 15).
					Return(emptyResponse, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, `"success":true`)
				assert.Contains(t, body, `"recentRecipes":[]`)
			},
		},
		{
			name:           "Validation Error - invalid UUID format",
			targetIDPath:   "invalid-uuid",
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusUnprocessableEntity,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "Invalid user ID format")
			},
		},
		{
			name:           "Validation Error - per_type_limit not integer",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "per_type_limit=abc",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "per_type_limit must be a valid integer")
			},
		},
		{
			name:           "Validation Error - per_type_limit too low",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "per_type_limit=0",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "per_type_limit must be between 1 and 100")
			},
		},
		{
			name:           "Validation Error - per_type_limit too high",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "per_type_limit=101",
			mockRun:        func(_ *MockSocialService) {},
			expectedStatus: http.StatusBadRequest,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "VALIDATION_ERROR")
				assert.Contains(t, body, "per_type_limit must be between 1 and 100")
			},
		},
		{
			name:           "Not Found - user does not exist",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 15).
					Return(nil, service.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "USER_NOT_FOUND")
				assert.Contains(t, body, "User not found")
			},
		},
		{
			name:           "Forbidden - private profile",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 15).
					Return(nil, service.ErrAccessDenied)
			},
			expectedStatus: http.StatusForbidden,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "FORBIDDEN")
				assert.Contains(t, body, "Access to this user's activity is restricted")
			},
		},
		{
			name:           "Internal Error - service error",
			targetIDPath:   targetID.String(),
			requesterIDHdr: requesterID.String(),
			queryParams:    "",
			mockRun: func(m *MockSocialService) {
				m.On("GetUserActivity", mock.Anything, mock.AnythingOfType("*uuid.UUID"), targetID, 15).
					Return(nil, errUnexpectedService)
			},
			expectedStatus: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body string) {
				t.Helper()
				assert.Contains(t, body, "INTERNAL_ERROR")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockSocialService)
			if tt.mockRun != nil {
				tt.mockRun(mockSvc)
			}

			h := handler.NewSocialHandler(mockSvc)

			r := chi.NewRouter()
			r.Get("/users/{user_id}/activity", h.GetUserActivity)

			url := "/users/" + tt.targetIDPath + "/activity"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
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
