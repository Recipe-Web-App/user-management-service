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

// SocialRepository defines the interface for social data access.
type SocialRepository interface {
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.User, int, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.User, int, error)
	FollowUser(ctx context.Context, followerID, followeeID uuid.UUID) error
	UnfollowUser(ctx context.Context, followerID, followeeID uuid.UUID) error
	GetRecentRecipes(ctx context.Context, userID uuid.UUID, limit int) ([]dto.RecipeSummary, error)
	GetRecentFollows(ctx context.Context, userID uuid.UUID, limit int) ([]dto.UserSummary, error)
	GetRecentReviews(ctx context.Context, userID uuid.UUID, limit int) ([]dto.ReviewSummary, error)
	GetRecentFavorites(ctx context.Context, userID uuid.UUID, limit int) ([]dto.FavoriteSummary, error)
}

// SQLSocialRepository implements SocialRepository using a SQL database.
type SQLSocialRepository struct {
	db *sql.DB
}

// NewSocialRepository creates a new SQLSocialRepository.
func NewSocialRepository(db *sql.DB) *SQLSocialRepository {
	return &SQLSocialRepository{db: db}
}

// GetFollowing retrieves the list of users that the specified user follows with pagination.
func (r *SQLSocialRepository) GetFollowing(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, int, error) {
	// Get total count first
	totalCount, err := r.countFollowing(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	users, err := r.fetchFollowing(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

func (r *SQLSocialRepository) countFollowing(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM recipe_manager.user_follows
		WHERE follower_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count following: %w", err)
	}

	return count, nil
}

func (r *SQLSocialRepository) fetchFollowing(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, error) {
	query := `
		SELECT u.user_id, u.username, u.email, u.full_name, u.bio, u.is_active, u.created_at, u.updated_at
		FROM recipe_manager.user_follows uf
		JOIN recipe_manager.users u ON uf.followee_id = u.user_id
		WHERE uf.follower_id = $1
		ORDER BY uf.followed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch following: %w", err)
	}

	defer func() { _ = rows.Close() }()

	return scanUsers(rows)
}

func scanUsers(rows *sql.Rows) ([]dto.User, error) {
	var users []dto.User

	for rows.Next() {
		var (
			user                 dto.User
			email, fullName, bio sql.NullString
		)

		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan user: %w", err)
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

		users = append(users, user)
	}

	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating following results: %w", err)
	}

	return users, nil
}

// GetFollowers retrieves the list of users who follow the specified user with pagination.
func (r *SQLSocialRepository) GetFollowers(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, int, error) {
	// Get total count first
	totalCount, err := r.countFollowers(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	users, err := r.fetchFollowers(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

func (r *SQLSocialRepository) countFollowers(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM recipe_manager.user_follows
		WHERE followee_id = $1
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count followers: %w", err)
	}

	return count, nil
}

func (r *SQLSocialRepository) fetchFollowers(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.User, error) {
	query := `
		SELECT u.user_id, u.username, u.email, u.full_name, u.bio, u.is_active, u.created_at, u.updated_at
		FROM recipe_manager.user_follows uf
		JOIN recipe_manager.users u ON uf.follower_id = u.user_id
		WHERE uf.followee_id = $1
		ORDER BY uf.followed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch followers: %w", err)
	}

	defer func() { _ = rows.Close() }()

	return scanUsers(rows)
}

// FollowUser creates a follow relationship between follower and followee.
// Uses ON CONFLICT DO NOTHING for idempotency - duplicate follows are silently ignored.
// Also handles the case where a database trigger raises an error for existing follows.
func (r *SQLSocialRepository) FollowUser(ctx context.Context, followerID, followeeID uuid.UUID) error {
	query := `
		INSERT INTO recipe_manager.user_follows (follower_id, followee_id, followed_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (follower_id, followee_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, followerID, followeeID)
	if err != nil {
		// Handle PostgreSQL trigger that raises "already following" error
		// This is an idempotent operation - treat existing follows as success
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "P0001" &&
			strings.Contains(strings.ToLower(pgErr.Message), "already following") {
			return nil
		}

		return fmt.Errorf("failed to create follow relationship: %w", err)
	}

	return nil
}

// UnfollowUser removes a follow relationship between follower and followee.
// This operation is idempotent - deleting a non-existent relationship succeeds.
func (r *SQLSocialRepository) UnfollowUser(ctx context.Context, followerID, followeeID uuid.UUID) error {
	query := `
		DELETE FROM recipe_manager.user_follows
		WHERE follower_id = $1 AND followee_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, followerID, followeeID)
	if err != nil {
		return fmt.Errorf("failed to delete follow relationship: %w", err)
	}

	return nil
}

// GetRecentRecipes retrieves the most recent recipes created by a user.
func (r *SQLSocialRepository) GetRecentRecipes(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.RecipeSummary, error) {
	query := `
		SELECT recipe_id, title, created_at
		FROM recipe_manager.recipes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent recipes: %w", err)
	}

	defer func() { _ = rows.Close() }()

	return scanRecipeSummaries(rows)
}

func scanRecipeSummaries(rows *sql.Rows) ([]dto.RecipeSummary, error) {
	var recipes []dto.RecipeSummary

	for rows.Next() {
		var r dto.RecipeSummary

		err := rows.Scan(&r.RecipeID, &r.Title, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe summary: %w", err)
		}

		recipes = append(recipes, r)
	}

	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating recipe summaries: %w", err)
	}

	return recipes, nil
}

// GetRecentFollows retrieves the most recent users followed by a user.
func (r *SQLSocialRepository) GetRecentFollows(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.UserSummary, error) {
	query := `
		SELECT u.user_id, u.username, uf.followed_at
		FROM recipe_manager.user_follows uf
		JOIN recipe_manager.users u ON uf.followee_id = u.user_id
		WHERE uf.follower_id = $1 AND u.is_active = true
		ORDER BY uf.followed_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent follows: %w", err)
	}

	defer func() { _ = rows.Close() }()

	return scanUserSummaries(rows)
}

func scanUserSummaries(rows *sql.Rows) ([]dto.UserSummary, error) {
	var users []dto.UserSummary

	for rows.Next() {
		var u dto.UserSummary

		err := rows.Scan(&u.UserID, &u.Username, &u.FollowedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user summary: %w", err)
		}

		users = append(users, u)
	}

	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating user summaries: %w", err)
	}

	return users, nil
}

// GetRecentReviews retrieves the most recent reviews written by a user.
func (r *SQLSocialRepository) GetRecentReviews(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.ReviewSummary, error) {
	query := `
		SELECT review_id, recipe_id, rating, comment, created_at
		FROM recipe_manager.recipe_reviews
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent reviews: %w", err)
	}

	defer func() { _ = rows.Close() }()

	return scanReviewSummaries(rows)
}

func scanReviewSummaries(rows *sql.Rows) ([]dto.ReviewSummary, error) {
	var reviews []dto.ReviewSummary

	for rows.Next() {
		var r dto.ReviewSummary

		var comment sql.NullString

		err := rows.Scan(&r.ReviewID, &r.RecipeID, &r.Rating, &comment, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review summary: %w", err)
		}

		if comment.Valid {
			r.Comment = &comment.String
		}

		reviews = append(reviews, r)
	}

	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating review summaries: %w", err)
	}

	return reviews, nil
}

// GetRecentFavorites retrieves the most recent recipes favorited by a user.
func (r *SQLSocialRepository) GetRecentFavorites(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]dto.FavoriteSummary, error) {
	query := `
		SELECT rf.recipe_id, rec.title, rf.favorited_at
		FROM recipe_manager.recipe_favorites rf
		JOIN recipe_manager.recipes rec ON rf.recipe_id = rec.recipe_id
		WHERE rf.user_id = $1
		ORDER BY rf.favorited_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent favorites: %w", err)
	}

	defer func() { _ = rows.Close() }()

	return scanFavoriteSummaries(rows)
}

func scanFavoriteSummaries(rows *sql.Rows) ([]dto.FavoriteSummary, error) {
	var favorites []dto.FavoriteSummary

	for rows.Next() {
		var f dto.FavoriteSummary

		err := rows.Scan(&f.RecipeID, &f.Title, &f.FavoritedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite summary: %w", err)
		}

		favorites = append(favorites, f)
	}

	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating favorite summaries: %w", err)
	}

	return favorites, nil
}
