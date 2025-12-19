package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// ErrDuplicateUsername is returned when a username already exists.
var ErrDuplicateUsername = errors.New("username already exists")

// UserRepository defines the interface for user data access.
type UserRepository interface {
	FindUserByID(ctx context.Context, userID uuid.UUID) (*dto.User, error)
	FindPrivacyPreferencesByUserID(ctx context.Context, userID uuid.UUID) (*dto.PrivacyPreferences, error)
	IsFollowing(ctx context.Context, followerID, followedID uuid.UUID) (bool, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, update *dto.UserProfileUpdateRequest) (*dto.User, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]dto.UserSearchResult, int, error)
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
		WHERE follower_id = $1 AND followee_id = $2
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

// UpdateUser updates a user's profile and returns the updated user.
func (r *SQLUserRepository) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	update *dto.UserProfileUpdateRequest,
) (*dto.User, error) {
	setClauses, args, argIndex := buildUpdateClauses(update)
	args = append(args, userID)

	query := fmt.Sprintf(
		`UPDATE recipe_manager.users
		SET %s
		WHERE user_id = $%d
		RETURNING user_id, username, email, full_name, bio, is_active, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIndex)

	user, err := r.executeUpdateQuery(ctx, query, args)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func buildUpdateClauses(update *dto.UserProfileUpdateRequest) ([]string, []any, int) {
	setClauses := []string{"updated_at = NOW()"}
	args := []any{}
	argIndex := 1

	if update.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *update.Username)
		argIndex++
	}

	if update.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *update.Email)
		argIndex++
	}

	if update.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argIndex))
		args = append(args, *update.FullName)
		argIndex++
	}

	if update.Bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argIndex))
		args = append(args, *update.Bio)
		argIndex++
	}

	if update.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *update.IsActive)
		argIndex++
	}

	return setClauses, args, argIndex
}

func (r *SQLUserRepository) executeUpdateQuery(ctx context.Context, query string, args []any) (*dto.User, error) {
	var (
		user                 dto.User
		email, fullName, bio sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
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
		return nil, mapUpdateError(err)
	}

	assignNullableFields(&user, email, fullName, bio)

	return &user, nil
}

func mapUpdateError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicateUsername
	}

	return fmt.Errorf("failed to update user: %w", err)
}

func assignNullableFields(user *dto.User, email, fullName, bio sql.NullString) {
	if email.Valid {
		user.Email = &email.String
	}

	if fullName.Valid {
		user.FullName = &fullName.String
	}

	if bio.Valid {
		user.Bio = &bio.String
	}
}

// SearchUsers searches for active users by username or full name with pagination.
func (r *SQLUserRepository) SearchUsers(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]dto.UserSearchResult, int, error) {
	// Build search pattern for ILIKE
	searchPattern := "%" + query + "%"

	// Get total count first
	totalCount, err := r.countSearchResults(ctx, searchPattern)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	results, err := r.fetchSearchResults(ctx, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

func (r *SQLUserRepository) countSearchResults(ctx context.Context, searchPattern string) (int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM recipe_manager.users
		WHERE is_active = true
		  AND (username ILIKE $1 OR full_name ILIKE $1)
	`

	var count int

	err := r.db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return count, nil
}

func (r *SQLUserRepository) fetchSearchResults(
	ctx context.Context,
	searchPattern string,
	limit, offset int,
) ([]dto.UserSearchResult, error) {
	resultsQuery := `
		SELECT user_id, username, full_name, is_active, created_at, updated_at
		FROM recipe_manager.users
		WHERE is_active = true
		  AND (username ILIKE $1 OR full_name ILIKE $1)
		ORDER BY username ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, resultsQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	defer func() { _ = rows.Close() }()

	return scanSearchResults(rows)
}

func scanSearchResults(rows *sql.Rows) ([]dto.UserSearchResult, error) {
	var results []dto.UserSearchResult

	for rows.Next() {
		var (
			result   dto.UserSearchResult
			fullName sql.NullString
		)

		err := rows.Scan(
			&result.UserID,
			&result.Username,
			&fullName,
			&result.IsActive,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		if fullName.Valid {
			result.FullName = &fullName.String
		}

		results = append(results, result)
	}

	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, nil
}
