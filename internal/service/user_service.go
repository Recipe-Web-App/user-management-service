package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

// UserService defines business logic for user operations.
type UserService interface {
	GetUserProfile(ctx context.Context, requesterID, targetUserID uuid.UUID) (*dto.UserProfileResponse, error)
	UpdateUserProfile(
		ctx context.Context,
		userID uuid.UUID,
		update *dto.UserProfileUpdateRequest,
	) (*dto.UserProfileResponse, error)
}

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrProfilePrivate is returned when a profile is private and cannot be viewed.
var ErrProfilePrivate = errors.New("profile is private")

// ErrDuplicateUsername is returned when trying to use a username that already exists.
var ErrDuplicateUsername = errors.New("username already exists")

// UserServiceImpl implements UserService.
type UserServiceImpl struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
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
	if update.Username == nil && update.Email == nil && update.FullName == nil && update.Bio == nil {
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
