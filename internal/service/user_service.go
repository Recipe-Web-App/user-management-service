package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/redis"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

// DeleteTokenTTL is the duration for which delete confirmation tokens are valid.
const DeleteTokenTTL = 24 * time.Hour

// UserService defines business logic for user operations.
type UserService interface {
	GetUserProfile(ctx context.Context, requesterID, targetUserID uuid.UUID) (*dto.UserProfileResponse, error)
	UpdateUserProfile(
		ctx context.Context,
		userID uuid.UUID,
		update *dto.UserProfileUpdateRequest,
	) (*dto.UserProfileResponse, error)
	RequestAccountDeletion(ctx context.Context, userID uuid.UUID) (*dto.UserAccountDeleteRequestResponse, error)
	ConfirmAccountDeletion(
		ctx context.Context,
		userID uuid.UUID,
		token string,
	) (*dto.UserConfirmAccountDeleteResponse, error)
	SearchUsers(
		ctx context.Context,
		query string,
		limit, offset int,
		countOnly bool,
	) (*dto.UserSearchResponse, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserSearchResult, error)
}

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrProfilePrivate is returned when a profile is private and cannot be viewed.
var ErrProfilePrivate = errors.New("profile is private")

// ErrDuplicateUsername is returned when trying to use a username that already exists.
var ErrDuplicateUsername = errors.New("username already exists")

// ErrCacheUnavailable is returned when the cache (Redis) is not available.
var ErrCacheUnavailable = errors.New("cache unavailable")

// ErrInvalidToken is returned when a confirmation token is invalid or expired.
var ErrInvalidToken = errors.New("invalid or expired token")

// UserServiceImpl implements UserService.
type UserServiceImpl struct {
	repo       repository.UserRepository
	tokenStore repository.TokenStore
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository, tokenStore repository.TokenStore) *UserServiceImpl {
	return &UserServiceImpl{
		repo:       repo,
		tokenStore: tokenStore,
	}
}

// GetUserProfile retrieves a user profile respecting privacy settings.
func (s *UserServiceImpl) GetUserProfile(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
) (*dto.UserProfileResponse, error) {
	// 1. Fetch user
	user, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 2. Fetch privacy preferences
	privacy, err := s.repo.FindPrivacyPreferencesByUserID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch privacy preferences: %w", err)
	}

	// 3. Apply Privacy Logic
	canViewProfile, err := s.canViewProfile(ctx, requesterID, targetUserID, privacy)
	if err != nil {
		return nil, err
	}

	if !canViewProfile {
		return nil, ErrProfilePrivate
	}

	// 4. Construct Response
	return s.buildProfileResponse(user, privacy, requesterID == targetUserID), nil
}

// GetUserByID retrieves a public user profile by ID.
// Private and followers_only profiles are not accessible (returns ErrUserNotFound).
func (s *UserServiceImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserSearchResult, error) {
	// 1. Fetch user
	user, err := s.repo.FindUserByID(ctx, userID)
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

	// 3. Fetch privacy preferences
	privacy, err := s.repo.FindPrivacyPreferencesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch privacy preferences: %w", err)
	}

	// 4. Apply privacy rule - only public profiles are accessible
	if privacy.ProfileVisibility != "public" {
		return nil, ErrUserNotFound
	}

	// 5. Build response (UserSearchResult schema - limited fields)
	return &dto.UserSearchResult{
		UserID:    user.UserID,
		Username:  user.Username,
		FullName:  user.FullName,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *UserServiceImpl) canViewProfile(
	ctx context.Context,
	requesterID, targetUserID uuid.UUID,
	privacy *dto.PrivacyPreferences,
) (bool, error) {
	if requesterID == targetUserID {
		return true, nil
	}

	switch privacy.ProfileVisibility {
	case "public":
		return true, nil
	case "followers_only":
		isFollowing, err := s.repo.IsFollowing(ctx, requesterID, targetUserID)
		if err != nil {
			return false, fmt.Errorf("failed to check following status: %w", err)
		}

		return isFollowing, nil
	case "private":
		return false, nil
	default:
		return false, nil
	}
}

func (s *UserServiceImpl) buildProfileResponse(
	user *dto.User,
	privacy *dto.PrivacyPreferences,
	isSelf bool,
) *dto.UserProfileResponse {
	response := &dto.UserProfileResponse{
		UserID:    user.UserID,
		Username:  user.Username,
		Bio:       user.Bio,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Full Name
	if isSelf || privacy.ShowFullName {
		response.FullName = user.FullName
	}

	// Email
	if isSelf || privacy.ShowEmail {
		response.Email = user.Email
	}

	return response
}

// UpdateUserProfile updates a user's profile and returns the updated profile.
func (s *UserServiceImpl) UpdateUserProfile(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.UserProfileUpdateRequest,
) (*dto.UserProfileResponse, error) {
	// 1. Verify user exists before attempting update
	existingUser, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to verify user exists: %w", err)
	}

	// 2. Check if there are any fields to update
	noFieldsToUpdate := update.Username == nil && update.Email == nil &&
		update.FullName == nil && update.Bio == nil && update.IsActive == nil
	if noFieldsToUpdate {
		// No changes requested, return current profile
		return &dto.UserProfileResponse{
			UserID:    existingUser.UserID,
			Username:  existingUser.Username,
			Email:     existingUser.Email,
			FullName:  existingUser.FullName,
			Bio:       existingUser.Bio,
			IsActive:  existingUser.IsActive,
			CreatedAt: existingUser.CreatedAt,
			UpdatedAt: existingUser.UpdatedAt,
		}, nil
	}

	// 3. Perform the update
	updatedUser, err := s.repo.UpdateUser(ctx, userID, update)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		if errors.Is(err, repository.ErrDuplicateUsername) {
			return nil, ErrDuplicateUsername
		}

		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	// 4. Build response
	return &dto.UserProfileResponse{
		UserID:    updatedUser.UserID,
		Username:  updatedUser.Username,
		Email:     updatedUser.Email,
		FullName:  updatedUser.FullName,
		Bio:       updatedUser.Bio,
		IsActive:  updatedUser.IsActive,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

// RequestAccountDeletion creates a deletion request and returns a confirmation token.
func (s *UserServiceImpl) RequestAccountDeletion(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.UserAccountDeleteRequestResponse, error) {
	// 1. Check if token store is available
	if s.tokenStore == nil {
		return nil, ErrCacheUnavailable
	}

	// 2. Verify user exists
	_, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to verify user exists: %w", err)
	}

	// 3. Generate confirmation token
	token := uuid.New().String()

	// 4. Store token in cache with TTL (replaces any existing token)
	err = s.tokenStore.StoreDeleteToken(ctx, userID, token, DeleteTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCacheUnavailable, err)
	}

	// 5. Calculate expiration time
	expiresAt := time.Now().Add(DeleteTokenTTL)

	return &dto.UserAccountDeleteRequestResponse{
		UserID:            userID.String(),
		ConfirmationToken: token,
		ExpiresAt:         expiresAt,
	}, nil
}

// ConfirmAccountDeletion validates the token and deactivates the user account.
func (s *UserServiceImpl) ConfirmAccountDeletion(
	ctx context.Context,
	userID uuid.UUID,
	token string,
) (*dto.UserConfirmAccountDeleteResponse, error) {
	// 1. Check if token store is available
	if s.tokenStore == nil {
		return nil, ErrCacheUnavailable
	}

	// 2. Retrieve stored token from cache
	storedToken, err := s.tokenStore.GetDeleteToken(ctx, userID)
	if err != nil {
		if errors.Is(err, redis.ErrTokenNotFound) {
			return nil, ErrInvalidToken
		}

		return nil, fmt.Errorf("%w: %w", ErrCacheUnavailable, err)
	}

	// 3. Validate token matches
	if storedToken != token {
		return nil, ErrInvalidToken
	}

	// 4. Deactivate user (set is_active = false)
	isActive := false

	_, err = s.repo.UpdateUser(ctx, userID, &dto.UserProfileUpdateRequest{
		IsActive: &isActive,
	})
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}

	// 5. Delete token from cache (best-effort cleanup)
	_ = s.tokenStore.DeleteDeleteToken(ctx, userID)

	// 6. Return response
	return &dto.UserConfirmAccountDeleteResponse{
		UserID:        userID.String(),
		DeactivatedAt: time.Now(),
	}, nil
}

// SearchUsers searches for users by username or full name with pagination.
func (s *UserServiceImpl) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
	countOnly bool,
) (*dto.UserSearchResponse, error) {
	// Get results from repository
	results, totalCount, err := s.repo.SearchUsers(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// If countOnly, return only the count with empty results
	if countOnly {
		return &dto.UserSearchResponse{
			Results:    []dto.UserSearchResult{},
			TotalCount: totalCount,
			Limit:      limit,
			Offset:     offset,
		}, nil
	}

	// Ensure results is not nil (return empty slice instead)
	if results == nil {
		results = []dto.UserSearchResult{}
	}

	return &dto.UserSearchResponse{
		Results:    results,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
