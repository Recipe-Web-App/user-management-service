package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// UserRepository defines the interface for user data access.
type UserRepository interface {
	FindUserByID(ctx context.Context, userID uuid.UUID) (*dto.User, error)
	FindPrivacyPreferencesByUserID(ctx context.Context, userID uuid.UUID) (*dto.PrivacyPreferences, error)
	IsFollowing(ctx context.Context, followerID, followedID uuid.UUID) (bool, error)
}

// SQLUserRepository implements UserRepository using a SQL database.
type SQLUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new SQLUserRepository.
func NewUserRepository(db *sql.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

// FindUserByID retrieves a user by their ID.
func (r *SQLUserRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*dto.User, error) {
	query := `
		SELECT user_id, username, email, full_name, bio, is_active, created_at, updated_at
		FROM recipe_manager.users
		WHERE user_id = $1
	`

	var (
		user                 dto.User
		email, fullName, bio sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&email,
		&fullName,
		&bio,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if email.Valid {
		user.Email = &email.String
	}

	if fullName.Valid {
		user.FullName = &fullName.String
	}

	if bio.Valid {
		user.Bio = &bio.String
	}

	return &user, nil
}

// FindPrivacyPreferencesByUserID retrieves privacy preferences for a user.
func (r *SQLUserRepository) FindPrivacyPreferencesByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.PrivacyPreferences, error) {
	// Note: The PrivacyPreferences DTO expects booleans (ShowEmail, etc.) but the DB schema
	// uses enums (contact_info_visibility, etc.). We map best-effort here.
	// We map contact_info_visibility='PUBLIC' -> ShowEmail=true.
	// We map FRIENDS_ONLY to DTO's followers_only (implied equivalent in specific context).
	query := `
		SELECT profile_visibility, contact_info_visibility
		FROM recipe_manager.user_privacy_preferences
		WHERE user_id = $1
	`

	// Default values matching typical "Public" profile expectation if row missing
	// (Schema defaults are PUBLIC/PRIVATE for contact)
	prefs := &dto.PrivacyPreferences{
		ProfileVisibility: "public",
		ShowEmail:         false,
		ShowFullName:      true, // No DB column, default true
		AllowFollows:      true, // No DB column, default true
		AllowMessages:     true, // No DB column, default true
	}

	var profileVisibility, contactVisibility string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profileVisibility,
		&contactVisibility,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no preferences found, return defaults
			return prefs, nil
		}

		return nil, fmt.Errorf("failed to query privacy preferences: %w", err)
	}

	// Map DB Enum to DTO String
	// DB: PUBLIC, FRIENDS_ONLY, PRIVATE
	// DTO: public, followers_only, private
	switch profileVisibility {
	case "FRIENDS_ONLY":
		prefs.ProfileVisibility = "followers_only"
	case "PRIVATE":
		prefs.ProfileVisibility = "private"
	default:
		prefs.ProfileVisibility = "public"
	}

	// Map Contact Visibility to ShowEmail
	if contactVisibility == "PUBLIC" {
		prefs.ShowEmail = true
	} else {
		prefs.ShowEmail = false
	}

	return prefs, nil
}

// IsFollowing checks if followerID follows followedID.
func (r *SQLUserRepository) IsFollowing(ctx context.Context, followerID, followedID uuid.UUID) (bool, error) {
	query := `
		SELECT 1 FROM recipe_manager.user_follows
		WHERE follower_id = $1 AND followed_id = $2
	`

	var exists int

	err := r.db.QueryRowContext(ctx, query, followerID, followedID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("failed to check following status: %w", err)
	}

	return true, nil
}
