package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

const (
	selectRecentRecipesQuery = `SELECT recipe_id, title, created_at FROM recipe_manager.recipes ` +
		`WHERE user_id = \$1 ORDER BY created_at DESC LIMIT \$2`
	selectRecentFollowsQuery = `SELECT u.user_id, u.username, uf.followed_at ` +
		`FROM recipe_manager.user_follows uf ` +
		`JOIN recipe_manager.users u ON uf.followee_id = u.user_id ` +
		`WHERE uf.follower_id = \$1 AND u.is_active = true ` +
		`ORDER BY uf.followed_at DESC LIMIT \$2`
	selectRecentReviewsQuery = `SELECT review_id, recipe_id, rating, comment, created_at ` +
		`FROM recipe_manager.recipe_reviews WHERE user_id = \$1 ` +
		`ORDER BY created_at DESC LIMIT \$2`
	selectRecentFavoritesQuery = `SELECT rf.recipe_id, rec.title, rf.favorited_at ` +
		`FROM recipe_manager.recipe_favorites rf ` +
		`JOIN recipe_manager.recipes rec ON rf.recipe_id = rec.recipe_id ` +
		`WHERE rf.user_id = \$1 ORDER BY rf.favorited_at DESC LIMIT \$2`
)

var errDBMock = errors.New("mock database error")

//nolint:dupl,funlen // Test structure is intentionally similar for consistency
func TestSocialRepositoryGetRecentRecipes(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Now()
	limit := 15

	t.Run("Success - returns recipes", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"recipe_id", "title", "created_at"}).
			AddRow(1, "Pasta Carbonara", now).
			AddRow(2, "Chicken Tikka", now.Add(-time.Hour))

		mock.ExpectQuery(selectRecentRecipesQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		recipes, err := repo.GetRecentRecipes(context.Background(), userID, limit)
		require.NoError(t, err)
		require.Len(t, recipes, 2)
		assert.Equal(t, 1, recipes[0].RecipeID)
		assert.Equal(t, "Pasta Carbonara", recipes[0].Title)
		assert.Equal(t, 2, recipes[1].RecipeID)
		assert.Equal(t, "Chicken Tikka", recipes[1].Title)
	})

	t.Run("Success - empty result", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"recipe_id", "title", "created_at"})

		mock.ExpectQuery(selectRecentRecipesQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		recipes, err := repo.GetRecentRecipes(context.Background(), userID, limit)
		require.NoError(t, err)
		assert.Empty(t, recipes)
	})

	t.Run("Error - database error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		mock.ExpectQuery(selectRecentRecipesQuery).
			WithArgs(userID, limit).
			WillReturnError(errDBMock)
		mock.ExpectClose()

		recipes, err := repo.GetRecentRecipes(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, recipes)
		assert.Contains(t, err.Error(), "failed to fetch recent recipes")
	})
}

//nolint:funlen // table-driven test with multiple sub-tests
func TestSocialRepositoryGetRecentFollows(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	followedUserID1 := uuid.New()
	followedUserID2 := uuid.New()
	now := time.Now()
	limit := 15

	t.Run("Success - returns followed users", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"user_id", "username", "followed_at"}).
			AddRow(followedUserID1, "chef1", now).
			AddRow(followedUserID2, "chef2", now.Add(-time.Hour))

		mock.ExpectQuery(selectRecentFollowsQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		follows, err := repo.GetRecentFollows(context.Background(), userID, limit)
		require.NoError(t, err)
		require.Len(t, follows, 2)
		assert.Equal(t, followedUserID1.String(), follows[0].UserID)
		assert.Equal(t, "chef1", follows[0].Username)
		assert.Equal(t, followedUserID2.String(), follows[1].UserID)
		assert.Equal(t, "chef2", follows[1].Username)
	})

	t.Run("Success - empty result", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"user_id", "username", "followed_at"})

		mock.ExpectQuery(selectRecentFollowsQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		follows, err := repo.GetRecentFollows(context.Background(), userID, limit)
		require.NoError(t, err)
		assert.Empty(t, follows)
	})

	t.Run("Error - database error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		mock.ExpectQuery(selectRecentFollowsQuery).
			WithArgs(userID, limit).
			WillReturnError(errDBMock)
		mock.ExpectClose()

		follows, err := repo.GetRecentFollows(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, follows)
		assert.Contains(t, err.Error(), "failed to fetch recent follows")
	})
}

//nolint:funlen // table-driven test with multiple sub-tests
func TestSocialRepositoryGetRecentReviews(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Now()
	limit := 15
	comment := "Great recipe!"

	t.Run("Success - returns reviews with comments", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"review_id", "recipe_id", "rating", "comment", "created_at"}).
			AddRow(1, 101, 5.0, comment, now).
			AddRow(2, 102, 4.0, nil, now.Add(-time.Hour))

		mock.ExpectQuery(selectRecentReviewsQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		reviews, err := repo.GetRecentReviews(context.Background(), userID, limit)
		require.NoError(t, err)
		require.Len(t, reviews, 2)
		assert.Equal(t, 1, reviews[0].ReviewID)
		assert.Equal(t, 101, reviews[0].RecipeID)
		assert.InDelta(t, 5.0, reviews[0].Rating, 0.001)
		assert.Equal(t, &comment, reviews[0].Comment)
		assert.Equal(t, 2, reviews[1].ReviewID)
		assert.Nil(t, reviews[1].Comment)
	})

	t.Run("Success - empty result", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"review_id", "recipe_id", "rating", "comment", "created_at"})

		mock.ExpectQuery(selectRecentReviewsQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		reviews, err := repo.GetRecentReviews(context.Background(), userID, limit)
		require.NoError(t, err)
		assert.Empty(t, reviews)
	})

	t.Run("Error - database error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		mock.ExpectQuery(selectRecentReviewsQuery).
			WithArgs(userID, limit).
			WillReturnError(errDBMock)
		mock.ExpectClose()

		reviews, err := repo.GetRecentReviews(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, reviews)
		assert.Contains(t, err.Error(), "failed to fetch recent reviews")
	})
}

//nolint:dupl,funlen // Test structure is intentionally similar for consistency
func TestSocialRepositoryGetRecentFavorites(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Now()
	limit := 15

	t.Run("Success - returns favorites", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"recipe_id", "title", "favorited_at"}).
			AddRow(1, "Apple Pie", now).
			AddRow(2, "Chocolate Cake", now.Add(-time.Hour))

		mock.ExpectQuery(selectRecentFavoritesQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		favorites, err := repo.GetRecentFavorites(context.Background(), userID, limit)
		require.NoError(t, err)
		require.Len(t, favorites, 2)
		assert.Equal(t, 1, favorites[0].RecipeID)
		assert.Equal(t, "Apple Pie", favorites[0].Title)
		assert.Equal(t, 2, favorites[1].RecipeID)
		assert.Equal(t, "Chocolate Cake", favorites[1].Title)
	})

	t.Run("Success - empty result", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		rows := sqlmock.NewRows([]string{"recipe_id", "title", "favorited_at"})

		mock.ExpectQuery(selectRecentFavoritesQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		favorites, err := repo.GetRecentFavorites(context.Background(), userID, limit)
		require.NoError(t, err)
		assert.Empty(t, favorites)
	})

	t.Run("Error - database error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		mock.ExpectQuery(selectRecentFavoritesQuery).
			WithArgs(userID, limit).
			WillReturnError(errDBMock)
		mock.ExpectClose()

		favorites, err := repo.GetRecentFavorites(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, favorites)
		assert.Contains(t, err.Error(), "failed to fetch recent favorites")
	})
}

// Row scan error tests.
//
//nolint:funlen // table-driven test with multiple sub-tests
func TestSocialRepositoryScanErrors(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	limit := 15

	t.Run("GetRecentRecipes - scan error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"recipe_id", "title"}).
			AddRow(1, "Pasta")

		mock.ExpectQuery(selectRecentRecipesQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		recipes, err := repo.GetRecentRecipes(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, recipes)
	})

	t.Run("GetRecentFollows - scan error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"user_id", "username"}).
			AddRow(uuid.New(), "chef1")

		mock.ExpectQuery(selectRecentFollowsQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		follows, err := repo.GetRecentFollows(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, follows)
	})

	t.Run("GetRecentReviews - scan error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"review_id", "recipe_id"}).
			AddRow(1, 101)

		mock.ExpectQuery(selectRecentReviewsQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		reviews, err := repo.GetRecentReviews(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, reviews)
	})

	t.Run("GetRecentFavorites - scan error", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewSocialRepository(db)

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"recipe_id", "title"}).
			AddRow(1, "Apple Pie")

		mock.ExpectQuery(selectRecentFavoritesQuery).
			WithArgs(userID, limit).
			WillReturnRows(rows)
		mock.ExpectClose()

		favorites, err := repo.GetRecentFavorites(context.Background(), userID, limit)
		require.Error(t, err)
		assert.Nil(t, favorites)
	})
}
