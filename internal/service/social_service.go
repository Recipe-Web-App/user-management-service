package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/notification"
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
	FollowUser(
		ctx context.Context,
		followerID, targetUserID uuid.UUID,
	) (*dto.FollowResponse, error)
	UnfollowUser(
		ctx context.Context,
		followerID, targetUserID uuid.UUID,
	) (*dto.FollowResponse, error)
	GetUserActivity(
		ctx context.Context,
		requesterID *uuid.UUID,
		targetUserID uuid.UUID,
		perTypeLimit int,
	) (*dto.UserActivityResponse, error)
}

// ErrAccessDenied is returned when access to a resource is denied due to privacy settings.
var ErrAccessDenied = errors.New("access denied")

// ErrCannotFollowSelf is returned when a user tries to follow themselves.
var ErrCannotFollowSelf = errors.New("cannot follow yourself")

// ErrFollowNotAllowed is returned when target user has disabled follows.
var ErrFollowNotAllowed = errors.New("user does not allow follows")

// ErrCannotUnfollowSelf is returned when a user tries to unfollow themselves.
var ErrCannotUnfollowSelf = errors.New("cannot unfollow yourself")

// Profile visibility constants.
const (
	profileVisibilityPublic        = "public"
	profileVisibilityFollowersOnly = "followers_only"
	profileVisibilityPrivate       = "private"
)

// SocialServiceImpl implements SocialService.
type SocialServiceImpl struct {
	userRepo           repository.UserRepository
	socialRepo         repository.SocialRepository
	notificationClient notification.Client
}

// NewSocialService creates a new SocialService.
func NewSocialService(
	userRepo repository.UserRepository,
	socialRepo repository.SocialRepository,
	notificationClient notification.Client,
) *SocialServiceImpl {
	return &SocialServiceImpl{
		userRepo:           userRepo,
		socialRepo:         socialRepo,
		notificationClient: notificationClient,
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

// FollowUser creates a follow relationship from follower to target user.
func (s *SocialServiceImpl) FollowUser(
	ctx context.Context,
	followerID, targetUserID uuid.UUID,
) (*dto.FollowResponse, error) {
	// 1. Check self-follow
	if followerID == targetUserID {
		return nil, ErrCannotFollowSelf
	}

	// 2. Verify target user exists and is active
	targetUser, err := s.userRepo.FindUserByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch target user: %w", err)
	}

	if !targetUser.IsActive {
		return nil, ErrUserNotFound
	}

	// 3. Check privacy settings - if AllowFollows is false, return forbidden
	privacy, err := s.userRepo.FindPrivacyPreferencesByUserID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch privacy preferences: %w", err)
	}

	if !privacy.AllowFollows {
		return nil, ErrFollowNotAllowed
	}

	// 4. Create follow relationship (idempotent - duplicate follows are OK)
	err = s.socialRepo.FollowUser(ctx, followerID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to follow user: %w", err)
	}

	// 5. Send notification (fire-and-forget)
	// Use context.Background() to decouple from request context so notification
	// continues even if the request is cancelled.
	if s.notificationClient != nil {
		go s.notificationClient.NotifyNewFollower(context.Background(), targetUserID, followerID) //nolint:contextcheck
	}

	// 6. Return success response
	return &dto.FollowResponse{
		Message:     "Successfully followed user",
		IsFollowing: true,
	}, nil
}

// UnfollowUser removes a follow relationship from follower to target user.
func (s *SocialServiceImpl) UnfollowUser(
	ctx context.Context,
	followerID, targetUserID uuid.UUID,
) (*dto.FollowResponse, error) {
	// 1. Check self-unfollow
	if followerID == targetUserID {
		return nil, ErrCannotUnfollowSelf
	}

	// 2. Verify target user exists and is active
	targetUser, err := s.userRepo.FindUserByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch target user: %w", err)
	}

	if !targetUser.IsActive {
		return nil, ErrUserNotFound
	}

	// 3. Delete follow relationship (idempotent - success even if not following)
	err = s.socialRepo.UnfollowUser(ctx, followerID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to unfollow user: %w", err)
	}

	// 4. Return success response
	return &dto.FollowResponse{
		Message:     "Successfully unfollowed user",
		IsFollowing: false,
	}, nil
}

// GetUserActivity retrieves a user's recent activity with privacy checks.
//
//nolint:cyclop,funlen // Complexity and length are from sequential error checks which are clear and necessary
func (s *SocialServiceImpl) GetUserActivity(
	ctx context.Context,
	requesterID *uuid.UUID,
	targetUserID uuid.UUID,
	perTypeLimit int,
) (*dto.UserActivityResponse, error) {
	// 1. Verify target user exists and is active
	user, err := s.userRepo.FindUserByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if !user.IsActive {
		return nil, ErrUserNotFound
	}

	// 2. Check privacy settings
	canAccess, err := s.canAccessUserActivity(ctx, requesterID, targetUserID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, ErrAccessDenied
	}

	// 3. Fetch all activity data
	recipes, err := s.socialRepo.GetRecentRecipes(ctx, targetUserID, perTypeLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent recipes: %w", err)
	}

	follows, err := s.socialRepo.GetRecentFollows(ctx, targetUserID, perTypeLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent follows: %w", err)
	}

	reviews, err := s.socialRepo.GetRecentReviews(ctx, targetUserID, perTypeLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent reviews: %w", err)
	}

	favorites, err := s.socialRepo.GetRecentFavorites(ctx, targetUserID, perTypeLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent favorites: %w", err)
	}

	// 4. Ensure slices are not nil (return empty arrays in JSON)
	if recipes == nil {
		recipes = []dto.RecipeSummary{}
	}

	if follows == nil {
		follows = []dto.UserSummary{}
	}

	if reviews == nil {
		reviews = []dto.ReviewSummary{}
	}

	if favorites == nil {
		favorites = []dto.FavoriteSummary{}
	}

	return &dto.UserActivityResponse{
		UserID:          targetUserID.String(),
		RecentRecipes:   recipes,
		RecentFollows:   follows,
		RecentReviews:   reviews,
		RecentFavorites: favorites,
	}, nil
}

// canAccessUserActivity checks if requester can view target's activity.
func (s *SocialServiceImpl) canAccessUserActivity(
	ctx context.Context,
	requesterID *uuid.UUID,
	targetUserID uuid.UUID,
) (bool, error) {
	// User can always view their own activity
	if requesterID != nil && *requesterID == targetUserID {
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
		// Anonymous users cannot access followers_only profiles
		if requesterID == nil {
			return false, nil
		}
		// Check if requester follows the target user
		isFollowing, err := s.userRepo.IsFollowing(ctx, *requesterID, targetUserID)
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
		// When countOnly is true, return only totalCount (other fields remain nil)
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
