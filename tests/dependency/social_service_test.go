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
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	socialBaseURL = "/api/v1/user-management/users"
)

var (
	errUnexpectedUsersSliceType = errors.New("unexpected type for []dto.User")
	errDatabaseFailure          = errors.New("database error")
)

// MockSocialRepository is a mock implementation of repository.SocialRepository.
type MockSocialRepository struct {
	mock.Mock
}

func (m *MockSocialRepository) GetFollowing(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf("get following: %w", err)
	}

	if args.Get(0) == nil {
		return nil, args.Int(1), nil
	}

	users, ok := args.Get(0).([]dto.User)
	if !ok {
		return nil, 0, errUnexpectedUsersSliceType
	}

	return users, args.Int(1), nil
}

func (m *MockSocialRepository) GetFollowers(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf("get followers: %w", err)
	}

	if args.Get(0) == nil {
		return nil, args.Int(1), nil
	}

	users, ok := args.Get(0).([]dto.User)
	if !ok {
		return nil, 0, errUnexpectedUsersSliceType
	}

	return users, args.Int(1), nil
}

func (m *MockSocialRepository) FollowUser(
	ctx context.Context,
	followerID, followeeID uuid.UUID,
) error {
	args := m.Called(ctx, followerID, followeeID)

	err := args.Error(0)
	if err != nil {
		return fmt.Errorf("follow user: %w", err)
	}

	return nil
}

type socialTestFixture struct {
	handler        http.Handler
	mockUserRepo   *MockUserRepository
	mockSocialRepo *MockSocialRepository
	requesterID    uuid.UUID
}

func setupSocialTest(t *testing.T) *socialTestFixture {
	t.Helper()

	mockUserRepo := new(MockUserRepository)
	mockSocialRepo := new(MockSocialRepository)
	cfg := &config.Config{}

	container, err := app.NewContainer(app.ContainerConfig{
		Config:     cfg,
		UserRepo:   mockUserRepo,
		SocialRepo: mockSocialRepo,
	})
	require.NoError(t, err)

	srv := server.NewServerWithContainer(container)

	return &socialTestFixture{
		handler:        srv.Handler,
		mockUserRepo:   mockUserRepo,
		mockSocialRepo: mockSocialRepo,
		requesterID:    uuid.New(),
	}
}

func newGetFollowingRequest(t *testing.T, targetUserID, requesterID uuid.UUID, queryParams string) *http.Request {
	t.Helper()

	reqPath := fmt.Sprintf("%s/%s/following", socialBaseURL, targetUserID)
	if queryParams != "" {
		reqPath += "?" + queryParams
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set(headerUserID, requesterID.String())

	return req
}

func createTestUserForSocial(userID uuid.UUID) *dto.User {
	return &dto.User{
		UserID:    userID.String(),
		Username:  "socialuser",
		FullName:  func() *string { s := "Social User"; return &s }(),
		Email:     func() *string { s := "social@example.com"; return &s }(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

//nolint:funlen,maintidx // table-driven test with many test cases
func TestGetFollowing(t *testing.T) {
	t.Parallel()

	t.Run("Success_PublicProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}
		now := time.Now()

		followedUsers := []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "followed1",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				UserID:    uuid.New().String(),
				Username:  "followed2",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 20, 0).Return(followedUsers, 2, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, ""))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 2, resp.Data.TotalCount)
		assert.Len(t, resp.Data.FollowedUsers, 2)
	})

	t.Run("Success_OwnProfile_PrivateSettings", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		userID := fix.requesterID
		user := createTestUserForSocial(userID)
		privatePrivacy := &dto.PrivacyPreferences{ProfileVisibility: "private"}
		now := time.Now()

		followedUsers := []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "followed",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, userID).Return(privatePrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowing", mock.Anything, userID, 20, 0).Return(followedUsers, 1, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, userID, fix.requesterID, ""))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 1, resp.Data.TotalCount)
	})

	t.Run("Success_FollowersOnly_WhenFollowing", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).
			Return(followersOnlyPrivacy, nil).Once()
		fix.mockUserRepo.On("IsFollowing", mock.Anything, fix.requesterID, targetUserID).Return(true, nil).Once()
		fix.mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 20, 0).Return([]dto.User{}, 0, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, ""))

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success_CountOnly", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 20, 0).Return(nil, 42, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, "count_only=true"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 42, resp.Data.TotalCount)
		assert.Nil(t, resp.Data.Limit)
		assert.Nil(t, resp.Data.Offset)
	})

	t.Run("Success_WithPagination", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 50, 10).Return([]dto.User{}, 100, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, "limit=50&offset=10"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 100, resp.Data.TotalCount)
		require.NotNil(t, resp.Data.Limit)
		assert.Equal(t, 50, *resp.Data.Limit)
		require.NotNil(t, resp.Data.Offset)
		assert.Equal(t, 10, *resp.Data.Offset)
	})

	t.Run("Forbidden_PrivateProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		privatePrivacy := &dto.PrivacyPreferences{ProfileVisibility: "private"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(privatePrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "FORBIDDEN")
	})

	t.Run("Forbidden_FollowersOnly_NotFollowing", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).
			Return(followersOnlyPrivacy, nil).Once()
		fix.mockUserRepo.On("IsFollowing", mock.Anything, fix.requesterID, targetUserID).Return(false, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "FORBIDDEN")
	})

	t.Run("NotFound_UserDoesNotExist", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		nonExistentID := uuid.New()

		fix.mockUserRepo.On("FindUserByID", mock.Anything, nonExistentID).Return(nil, repository.ErrUserNotFound).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, nonExistentID, fix.requesterID, ""))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	})

	t.Run("NotFound_InactiveUser", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		inactiveUserID := uuid.New()
		inactiveUser := &dto.User{
			UserID:   inactiveUserID.String(),
			Username: "inactive",
			IsActive: false,
		}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, inactiveUserID).Return(inactiveUser, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, inactiveUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()

		reqPath := fmt.Sprintf("%s/%s/following", socialBaseURL, targetUserID)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
	})

	t.Run("ValidationError_InvalidUUID", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)

		reqPath := socialBaseURL + "/invalid-uuid/following"
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
		require.NoError(t, err)
		req.Header.Set(headerUserID, fix.requesterID.String())

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("ValidationError_InvalidLimit", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, "limit=abc"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("ValidationError_LimitOutOfRange", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, "limit=101"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("InternalError_RepositoryFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowing", mock.Anything, targetUserID, 20, 0).
			Return(nil, 0, errDatabaseFailure).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowingRequest(t, targetUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "INTERNAL_ERROR")
	})
}

func newGetFollowersRequest(t *testing.T, targetUserID, requesterID uuid.UUID, queryParams string) *http.Request {
	t.Helper()

	reqPath := fmt.Sprintf("%s/%s/followers", socialBaseURL, targetUserID)
	if queryParams != "" {
		reqPath += "?" + queryParams
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set(headerUserID, requesterID.String())

	return req
}

//nolint:funlen,maintidx,dupl // table-driven test with many test cases
func TestGetFollowers(t *testing.T) {
	t.Parallel()

	t.Run("Success_PublicProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}
		now := time.Now()

		followers := []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "follower1",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				UserID:    uuid.New().String(),
				Username:  "follower2",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 20, 0).Return(followers, 2, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, ""))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 2, resp.Data.TotalCount)
		assert.Len(t, resp.Data.FollowedUsers, 2)
	})

	t.Run("Success_OwnProfile_PrivateSettings", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		userID := fix.requesterID
		user := createTestUserForSocial(userID)
		now := time.Now()

		followers := []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "follower",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil).Once()
		fix.mockSocialRepo.On("GetFollowers", mock.Anything, userID, 20, 0).Return(followers, 1, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, userID, fix.requesterID, ""))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 1, resp.Data.TotalCount)
	})

	t.Run("Success_FollowersOnly_WhenFollowing", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).
			Return(followersOnlyPrivacy, nil).Once()
		fix.mockUserRepo.On("IsFollowing", mock.Anything, fix.requesterID, targetUserID).Return(true, nil).Once()
		fix.mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 20, 0).Return([]dto.User{}, 0, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, ""))

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Success_CountOnly", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 20, 0).Return(nil, 42, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, "count_only=true"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 42, resp.Data.TotalCount)
		assert.Nil(t, resp.Data.Limit)
		assert.Nil(t, resp.Data.Offset)
	})

	t.Run("Success_WithPagination", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 50, 10).Return([]dto.User{}, 100, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, "limit=50&offset=10"))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool                         `json:"success"`
			Data    dto.GetFollowedUsersResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.Equal(t, 100, resp.Data.TotalCount)
		require.NotNil(t, resp.Data.Limit)
		assert.Equal(t, 50, *resp.Data.Limit)
		require.NotNil(t, resp.Data.Offset)
		assert.Equal(t, 10, *resp.Data.Offset)
	})

	t.Run("Forbidden_PrivateProfile", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		privatePrivacy := &dto.PrivacyPreferences{ProfileVisibility: "private"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(privatePrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "FORBIDDEN")
	})

	t.Run("Forbidden_FollowersOnly_NotFollowing", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).
			Return(followersOnlyPrivacy, nil).Once()
		fix.mockUserRepo.On("IsFollowing", mock.Anything, fix.requesterID, targetUserID).Return(false, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "FORBIDDEN")
	})

	t.Run("NotFound_UserDoesNotExist", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		nonExistentID := uuid.New()

		fix.mockUserRepo.On("FindUserByID", mock.Anything, nonExistentID).Return(nil, repository.ErrUserNotFound).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, nonExistentID, fix.requesterID, ""))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	})

	t.Run("NotFound_InactiveUser", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		inactiveUserID := uuid.New()
		inactiveUser := &dto.User{
			UserID:   inactiveUserID.String(),
			Username: "inactive",
			IsActive: false,
		}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, inactiveUserID).Return(inactiveUser, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, inactiveUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()

		reqPath := fmt.Sprintf("%s/%s/followers", socialBaseURL, targetUserID)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
	})

	t.Run("ValidationError_InvalidUUID", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)

		reqPath := socialBaseURL + "/invalid-uuid/followers"
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqPath, nil)
		require.NoError(t, err)
		req.Header.Set(headerUserID, fix.requesterID.String())

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("ValidationError_InvalidLimit", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, "limit=abc"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("ValidationError_LimitOutOfRange", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, "limit=101"))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("InternalError_RepositoryFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("GetFollowers", mock.Anything, targetUserID, 20, 0).
			Return(nil, 0, errDatabaseFailure).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newGetFollowersRequest(t, targetUserID, fix.requesterID, ""))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "INTERNAL_ERROR")
	})
}

func newFollowUserRequest(
	t *testing.T,
	followerID, targetUserID, requesterID uuid.UUID,
	isAdmin bool,
) *http.Request {
	t.Helper()

	reqPath := fmt.Sprintf("%s/%s/follow/%s", socialBaseURL, followerID, targetUserID)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set(headerUserID, requesterID.String())

	if isAdmin {
		req.Header.Set("X-User-Role", "admin")
	}

	return req
}

//nolint:funlen // table-driven test with many test cases
func TestFollowUser(t *testing.T) {
	t.Parallel()

	t.Run("Success_FollowUser", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := fix.requesterID
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public", AllowFollows: true}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("FollowUser", mock.Anything, followerID, targetUserID).Return(nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, followerID, targetUserID, fix.requesterID, false))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool               `json:"success"`
			Data    dto.FollowResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.True(t, resp.Data.IsFollowing)
		assert.Contains(t, resp.Data.Message, "Successfully followed user")
	})

	t.Run("Success_AdminOverride", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		adminID := fix.requesterID
		followerID := uuid.New()
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public", AllowFollows: true}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("FollowUser", mock.Anything, followerID, targetUserID).Return(nil).Once()

		rr := httptest.NewRecorder()
		// Admin can follow on behalf of another user
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, followerID, targetUserID, adminID, true))

		require.Equal(t, http.StatusOK, rr.Code)

		var resp struct {
			Success bool               `json:"success"`
			Data    dto.FollowResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Success)
		assert.True(t, resp.Data.IsFollowing)
	})

	t.Run("Success_Idempotent_AlreadyFollowing", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := fix.requesterID
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public", AllowFollows: true}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		// ON CONFLICT DO NOTHING - returns success even if already following
		fix.mockSocialRepo.On("FollowUser", mock.Anything, followerID, targetUserID).Return(nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, followerID, targetUserID, fix.requesterID, false))

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("BadRequest_SelfFollow", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		userID := fix.requesterID

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, userID, userID, fix.requesterID, false))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
		assert.Contains(t, rr.Body.String(), "Cannot follow yourself")
	})

	t.Run("NotFound_TargetDoesNotExist", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := fix.requesterID
		nonExistentID := uuid.New()

		fix.mockUserRepo.On("FindUserByID", mock.Anything, nonExistentID).
			Return(nil, repository.ErrUserNotFound).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, followerID, nonExistentID, fix.requesterID, false))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	})

	t.Run("NotFound_TargetInactive", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := fix.requesterID
		inactiveUserID := uuid.New()
		inactiveUser := &dto.User{
			UserID:   inactiveUserID.String(),
			Username: "inactive",
			IsActive: false,
		}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, inactiveUserID).Return(inactiveUser, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, followerID, inactiveUserID, fix.requesterID, false))

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "USER_NOT_FOUND")
	})

	t.Run("Forbidden_UserDoesNotAllowFollows", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := fix.requesterID
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		noFollowsPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public", AllowFollows: false}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).
			Return(noFollowsPrivacy, nil).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, followerID, targetUserID, fix.requesterID, false))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "FORBIDDEN")
	})

	t.Run("Forbidden_UserIdMismatch_NonAdmin", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		differentUserID := uuid.New()
		targetUserID := uuid.New()

		rr := httptest.NewRecorder()
		// Non-admin trying to follow on behalf of another user
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, differentUserID, targetUserID, fix.requesterID, false))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "FORBIDDEN")
	})

	t.Run("Unauthorized_MissingHeader", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := uuid.New()
		targetUserID := uuid.New()

		reqPath := fmt.Sprintf("%s/%s/follow/%s", socialBaseURL, followerID, targetUserID)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
		require.NoError(t, err)
		// No X-User-Id header

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
	})

	t.Run("ValidationError_InvalidUserUUID", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		targetUserID := uuid.New()

		reqPath := fmt.Sprintf("%s/invalid-uuid/follow/%s", socialBaseURL, targetUserID)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
		require.NoError(t, err)
		req.Header.Set(headerUserID, fix.requesterID.String())

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("ValidationError_InvalidTargetUUID", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := fix.requesterID

		reqPath := fmt.Sprintf("%s/%s/follow/invalid-uuid", socialBaseURL, followerID)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqPath, nil)
		require.NoError(t, err)
		req.Header.Set(headerUserID, fix.requesterID.String())

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.Contains(t, rr.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("InternalError_RepositoryFailure", func(t *testing.T) {
		t.Parallel()

		fix := setupSocialTest(t)
		followerID := fix.requesterID
		targetUserID := uuid.New()
		targetUser := createTestUserForSocial(targetUserID)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public", AllowFollows: true}

		fix.mockUserRepo.On("FindUserByID", mock.Anything, targetUserID).Return(targetUser, nil).Once()
		fix.mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetUserID).Return(publicPrivacy, nil).Once()
		fix.mockSocialRepo.On("FollowUser", mock.Anything, followerID, targetUserID).
			Return(errDatabaseFailure).Once()

		rr := httptest.NewRecorder()
		fix.handler.ServeHTTP(rr, newFollowUserRequest(t, followerID, targetUserID, fix.requesterID, false))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "INTERNAL_ERROR")
	})
}
