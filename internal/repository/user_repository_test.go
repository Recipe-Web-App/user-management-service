package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	selectUserQuery = `SELECT user_id, username, email, full_name, bio, is_active, ` +
		`created_at, updated_at FROM recipe_manager.users WHERE user_id = \$1`
	selectPrivacyQuery = `SELECT profile_visibility, contact_info_visibility ` +
		`FROM recipe_manager.user_privacy_preferences WHERE user_id = \$1`
)

func TestSQLUserRepositoryFindUserByID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewUserRepository(db)

		rows := sqlmock.NewRows([]string{
			"user_id", "username", "email", "full_name",
			"bio", "is_active", "created_at", "updated_at",
		}).AddRow(userID, "testuser", "email@example.com", "Test User", "Bio", true, now, now)

		mock.ExpectQuery(selectUserQuery).
			WithArgs(userID).
			WillReturnRows(rows)
		mock.ExpectClose()

		user, err := repo.FindUserByID(context.Background(), userID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID.String(), user.UserID)
		assert.Equal(t, "email@example.com", *user.Email)
	})

	t.Run("Not Found", func(t *testing.T) {
		t.Parallel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)

		defer func() {
			require.NoError(t, db.Close())
		}()

		repo := repository.NewUserRepository(db)

		mock.ExpectQuery(selectUserQuery).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectClose()

		user, err := repo.FindUserByID(context.Background(), userID)
		require.ErrorIs(t, err, repository.ErrUserNotFound)
		assert.Nil(t, user)
	})
}

type privacyTestCase struct {
	name               string
	mockSetup          func(sqlmock.Sqlmock)
	expectedVisibility string
	expectedEmailShow  bool
}

func TestSQLUserRepositoryFindPrivacyPreferencesByUserID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []privacyTestCase{
		{
			name: "Success - Public",
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"profile_visibility", "contact_info_visibility"}).
					AddRow("PUBLIC", "PUBLIC")
				m.ExpectQuery(selectPrivacyQuery).WithArgs(userID).WillReturnRows(rows)
			},
			expectedVisibility: "public",
			expectedEmailShow:  true,
		},
		{
			name: "Success - Friends Only Mapped",
			mockSetup: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"profile_visibility", "contact_info_visibility"}).
					AddRow("FRIENDS_ONLY", "PRIVATE")
				m.ExpectQuery(selectPrivacyQuery).WithArgs(userID).WillReturnRows(rows)
			},
			expectedVisibility: "followers_only",
			expectedEmailShow:  false,
		},
		{
			name: "Not Found - Defaults",
			mockSetup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(selectPrivacyQuery).WithArgs(userID).WillReturnError(sql.ErrNoRows)
			},
			expectedVisibility: "public",
			expectedEmailShow:  false,
		},
	}

	runPrivacyTest(t, tests, userID)
}

func runPrivacyTest(t *testing.T, tests []privacyTestCase, userID uuid.UUID) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			defer func() {
				require.NoError(t, db.Close())
			}()

			repo := repository.NewUserRepository(db)

			tt.mockSetup(mock)
			mock.ExpectClose()

			prefs, err := repo.FindPrivacyPreferencesByUserID(context.Background(), userID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVisibility, prefs.ProfileVisibility)
			assert.Equal(t, tt.expectedEmailShow, prefs.ShowEmail)
		})
	}
}
