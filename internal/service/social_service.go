package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

// SocialService defines business logic for social operations.
type SocialService interface {
	GetFollowing(
		ctx context.Context,
		requesterID, targetUserID uuid.UUID,
		limit, offset int,
		countOnly bool,
	) (*dto.GetFollowedUsersResponse, error)
	GetFollowers(
		ctx context.Context,
		requesterID, targetUserID uuid.UUID,
		limit, offset int,
		countOnly bool,
	) (*dto.GetFollowedUsersResponse, error)
}

// ErrAccessDenied is returned when access to a resource is denied due to privacy settings.
var ErrAccessDenied = errors.New("access denied")

// Profile visibility constants.
const (
	profileVisibilityPublic        = "public"
	profileVisibilityFollowersOnly = "followers_only"
	profileVisibilityPrivate       = "private"
)

// SocialServiceImpl implements SocialService.
type SocialServiceImpl struct {
	userRepo   repository.UserRepository
	socialRepo repository.SocialRepository
}

// NewSocialService creates a new SocialService.
func NewSocialService(
	userRepo repository.UserRepository,
	socialRepo repository.SocialRepository,
) *SocialServiceImpl {
	return &SocialServiceImpl{
		userRepo:   userRepo,
		socialRepo: socialRepo,
	}
}

// GetFollowing retrieves the list of users that the target user follows.
func (s *SocialServiceImpl) GetFollowing(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	limit, offset int,
	countOnly bool,
) (*dto.GetFollowedUsersResponse, error) {
	// 1. Verify target user exists
	user, err := s.userRepo.FindUserByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 2. Check if user is active
	if !user.IsActive {
		return nil, ErrUserNotFound
	}

	// 3. Check privacy settings
	canAccess, err := s.canAccessFollowingList(ctx, requesterID, targetUserID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, ErrAccessDenied
	}

	// 4. Get following list from repository
	users, totalCount, err := s.socialRepo.GetFollowing(ctx, targetUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get following list: %w", err)
	}

	// 5. Build response
	return s.buildFollowingResponse(users, totalCount, limit, offset, countOnly), nil
}

// GetFollowers retrieves the list of users who follow the target user.
func (s *SocialServiceImpl) GetFollowers(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	limit, offset int,
	countOnly bool,
) (*dto.GetFollowedUsersResponse, error) {
	// 1. Verify target user exists
	user, err := s.userRepo.FindUserByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 2. Check if user is active
	if !user.IsActive {
		return nil, ErrUserNotFound
	}

	// 3. Check privacy settings (same rules as following list)
	canAccess, err := s.canAccessFollowingList(ctx, requesterID, targetUserID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, ErrAccessDenied
	}

	// 4. Get followers list from repository
	users, totalCount, err := s.socialRepo.GetFollowers(ctx, targetUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers list: %w", err)
	}

	// 5. Build response
	return s.buildFollowingResponse(users, totalCount, limit, offset, countOnly), nil
}

func (s *SocialServiceImpl) canAccessFollowingList(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
) (bool, error) {
	// User can always view their own following list
	if requesterID == targetUserID {
		return true, nil
	}

	// Fetch privacy preferences
	privacy, err := s.userRepo.FindPrivacyPreferencesByUserID(ctx, targetUserID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch privacy preferences: %w", err)
	}

	switch privacy.ProfileVisibility {
	case profileVisibilityPublic:
		return true, nil
	case profileVisibilityFollowersOnly:
		// Check if requester follows the target user
		isFollowing, err := s.userRepo.IsFollowing(ctx, requesterID, targetUserID)
		if err != nil {
			return false, fmt.Errorf("failed to check following status: %w", err)
		}

		return isFollowing, nil
	case profileVisibilityPrivate:
		return false, nil
	default:
		return false, nil
	}
}

func (s *SocialServiceImpl) buildFollowingResponse(
	users []dto.User,
	totalCount, limit, offset int,
	countOnly bool,
) *dto.GetFollowedUsersResponse {
	response := &dto.GetFollowedUsersResponse{
		TotalCount: totalCount,
	}

	if countOnly {
		// When count_only is true, return only totalCount (other fields remain nil)
		return response
	}

	// Ensure users is not nil (return empty slice instead)
	if users == nil {
		users = []dto.User{}
	}

	response.FollowedUsers = users
	response.Limit = &limit
	response.Offset = &offset

	return response
}
