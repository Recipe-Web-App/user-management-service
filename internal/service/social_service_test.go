package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

const mockSocialErrorFmt = "mock error: %w"

var (
	errMockSocialArgs    = errors.New("mock: missing args")
	errMockSocialUser    = errors.New("invalid type assertion for User")
	errMockSocialPrivacy = errors.New("invalid type assertion for PrivacyPreferences")
	errRepoSocial        = errors.New("repository error")
)

// MockUserRepoForSocial is a mock implementation of repository.UserRepository.
type MockUserRepoForSocial struct {
	mock.Mock
}

func (m *MockUserRepoForSocial) FindUserByID(ctx context.Context, userID uuid.UUID) (*dto.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockSocialErrorFmt, err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.User); ok {
		return val, nil
	}

	return nil, errMockSocialUser
}

func (m *MockUserRepoForSocial) FindPrivacyPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.PrivacyPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockSocialErrorFmt, err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.PrivacyPreferences); ok {
		return val, nil
	}

	return nil, errMockSocialPrivacy
}

func (m *MockUserRepoForSocial) IsFollowing(
	ctx context.Context,
	followerID, followedID uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, followerID, followedID)

	err := args.Error(1)
	if err != nil {
		return args.Bool(0), fmt.Errorf(mockSocialErrorFmt, err)
	}

	return args.Bool(0), nil
}

func (m *MockUserRepoForSocial) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.UserProfileUpdateRequest,
) (*dto.User, error) {
	args := m.Called(ctx, userID, update)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf(mockSocialErrorFmt, err)
		}

		return nil, errMockSocialArgs
	}

	if val, ok := args.Get(0).(*dto.User); ok {
		return val, nil
	}

	return nil, errMockSocialUser
}

func (m *MockUserRepoForSocial) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]dto.UserSearchResult, int, error) {
	args := m.Called(ctx, query, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf(mockSocialErrorFmt, err)
	}

	results, _ := args.Get(0).([]dto.UserSearchResult)

	return results, args.Int(1), nil
}

// MockSocialRepo is a mock implementation of repository.SocialRepository.
type MockSocialRepo struct {
	mock.Mock
}

func (m *MockSocialRepo) GetFollowing(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf(mockSocialErrorFmt, err)
	}

	users, _ := args.Get(0).([]dto.User)

	return users, args.Int(1), nil
}

func (m *MockSocialRepo) GetFollowers(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	err := args.Error(2)
	if err != nil {
		return nil, 0, fmt.Errorf(mockSocialErrorFmt, err)
	}

	users, _ := args.Get(0).([]dto.User)

	return users, args.Int(1), nil
}

func createTestUser(userID uuid.UUID, isActive bool) *dto.User {
	now := time.Now()
	fullName := "Test User"
	email := "test@example.com"

	return &dto.User{
		UserID:    userID.String(),
		Username:  "testuser",
		Email:     &email,
		FullName:  &fullName,
		IsActive:  isActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func createFollowedUsers(count int) []dto.User {
	users := make([]dto.User, count)
	now := time.Now()

	for i := range count {
		fullName := fmt.Sprintf("User %d", i+1)
		users[i] = dto.User{
			UserID:    uuid.New().String(),
			Username:  fmt.Sprintf("user%d", i+1),
			FullName:  &fullName,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return users
}

//nolint:funlen,maintidx // table-driven test with many test cases
func TestSocialServiceGetFollowing(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	requesterID := uuid.New()

	t.Run("Success - public profile returns following list", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}
		followedUsers := createFollowedUsers(2)

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(publicPrivacy, nil).Once()
		mockSocialRepo.On("GetFollowing", mock.Anything, targetID, 20, 0).Return(followedUsers, 2, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 2, resp.TotalCount)
		assert.Len(t, resp.FollowedUsers, 2)
		assert.NotNil(t, resp.Limit)
		assert.NotNil(t, resp.Offset)
		assert.Equal(t, 20, *resp.Limit)
		assert.Equal(t, 0, *resp.Offset)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertExpectations(t)
	})

	t.Run("Success - viewing own following list with private profile", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		ownUser := createTestUser(requesterID, true)
		privatePrivacy := &dto.PrivacyPreferences{ProfileVisibility: "private"}
		followedUsers := createFollowedUsers(1)

		// User viewing their own list - no privacy check needed beyond initial fetch
		mockUserRepo.On("FindUserByID", mock.Anything, requesterID).Return(ownUser, nil).Once()
		mockSocialRepo.On("GetFollowing", mock.Anything, requesterID, 20, 0).Return(followedUsers, 1, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, requesterID, 20, 0, false)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 1, resp.TotalCount)

		// Privacy preferences should not be fetched when viewing own list
		mockUserRepo.AssertNotCalled(t, "FindPrivacyPreferencesByUserID", mock.Anything, mock.Anything)
		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertExpectations(t)

		_ = privatePrivacy // Silence unused variable warning
	})

	t.Run("Success - followers_only profile when requester follows target", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}
		followedUsers := createFollowedUsers(3)

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(followersOnlyPrivacy, nil).Once()
		mockUserRepo.On("IsFollowing", mock.Anything, requesterID, targetID).Return(true, nil).Once()
		mockSocialRepo.On("GetFollowing", mock.Anything, targetID, 20, 0).Return(followedUsers, 3, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 3, resp.TotalCount)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertExpectations(t)
	})

	t.Run("Success - count_only mode", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(publicPrivacy, nil).Once()
		mockSocialRepo.On("GetFollowing", mock.Anything, targetID, 20, 0).Return([]dto.User{}, 42, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, true)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 42, resp.TotalCount)
		assert.Nil(t, resp.FollowedUsers)
		assert.Nil(t, resp.Limit)
		assert.Nil(t, resp.Offset)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertExpectations(t)
	})

	t.Run("Success - empty following list", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(publicPrivacy, nil).Once()
		mockSocialRepo.On("GetFollowing", mock.Anything, targetID, 20, 0).Return([]dto.User{}, 0, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 0, resp.TotalCount)
		assert.Empty(t, resp.FollowedUsers)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertExpectations(t)
	})

	t.Run("Error - user not found", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(nil, repository.ErrUserNotFound).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		require.ErrorIs(t, err, service.ErrUserNotFound)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertNotCalled(t, "GetFollowing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error - user inactive", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		inactiveUser := createTestUser(targetID, false)

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(inactiveUser, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		require.ErrorIs(t, err, service.ErrUserNotFound)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertNotCalled(t, "GetFollowing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error - access denied for private profile", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		privatePrivacy := &dto.PrivacyPreferences{ProfileVisibility: "private"}

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(privatePrivacy, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		require.ErrorIs(t, err, service.ErrAccessDenied)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertNotCalled(t, "GetFollowing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error - access denied for followers_only when not following", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(followersOnlyPrivacy, nil).Once()
		mockUserRepo.On("IsFollowing", mock.Anything, requesterID, targetID).Return(false, nil).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		require.ErrorIs(t, err, service.ErrAccessDenied)

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertNotCalled(t, "GetFollowing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error - repository error on GetFollowing", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		publicPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "public"}

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(publicPrivacy, nil).Once()
		mockSocialRepo.On("GetFollowing", mock.Anything, targetID, 20, 0).Return(nil, 0, errRepoSocial).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get following list")

		mockUserRepo.AssertExpectations(t)
		mockSocialRepo.AssertExpectations(t)
	})

	t.Run("Error - repository error on FindUserByID", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(nil, errRepoSocial).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to fetch user")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Error - repository error on FindPrivacyPreferencesByUserID", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(nil, errRepoSocial).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to fetch privacy preferences")

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Error - repository error on IsFollowing", func(t *testing.T) {
		t.Parallel()

		mockUserRepo := new(MockUserRepoForSocial)
		mockSocialRepo := new(MockSocialRepo)

		targetUser := createTestUser(targetID, true)
		followersOnlyPrivacy := &dto.PrivacyPreferences{ProfileVisibility: "followers_only"}

		mockUserRepo.On("FindUserByID", mock.Anything, targetID).Return(targetUser, nil).Once()
		mockUserRepo.On("FindPrivacyPreferencesByUserID", mock.Anything, targetID).Return(followersOnlyPrivacy, nil).Once()
		mockUserRepo.On("IsFollowing", mock.Anything, requesterID, targetID).Return(false, errRepoSocial).Once()

		svc := service.NewSocialService(mockUserRepo, mockSocialRepo)
		resp, err := svc.GetFollowing(context.Background(), requesterID, targetID, 20, 0, false)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to check following status")

		mockUserRepo.AssertExpectations(t)
	})
}
