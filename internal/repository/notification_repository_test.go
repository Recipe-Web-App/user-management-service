package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

var errTestDB = errors.New("database error")

//nolint:funlen // Table-driven test with many test cases
func TestSQLNotificationRepository_GetNotifications(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID := uuid.New()
	now := time.Now()

	tests := []struct {
		name              string
		limit             int
		offset            int
		setupMock         func(mock sqlmock.Sqlmock)
		expectedCount     int
		expectedLen       int
		expectError       bool
		expectedFirstID   string
		expectedFirstRead bool
	}{
		{
			name:   "returns notifications with count",
			limit:  20,
			offset: 0,
			setupMock: func(mock sqlmock.Sqlmock) {
				// Count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnRows(countRows)

				// Main query
				rows := sqlmock.NewRows([]string{
					"notification_id", "user_id", "title", "message", "notification_type",
					"is_read", "is_deleted", "created_at", "updated_at",
				}).AddRow(
					notificationID.String(), userID.String(), "Test Title", "Test Message", "follow",
					false, false, now, now,
				)
				mock.ExpectQuery(`SELECT notification_id, user_id, title, message`).
					WithArgs(userID, 20, 0).
					WillReturnRows(rows)
			},
			expectedCount:     5,
			expectedLen:       1,
			expectError:       false,
			expectedFirstID:   notificationID.String(),
			expectedFirstRead: false,
		},
		{
			name:   "returns empty slice when no notifications",
			limit:  20,
			offset: 0,
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"notification_id", "user_id", "title", "message", "notification_type",
					"is_read", "is_deleted", "created_at", "updated_at",
				})
				mock.ExpectQuery(`SELECT notification_id, user_id, title, message`).
					WithArgs(userID, 20, 0).
					WillReturnRows(rows)
			},
			expectedCount: 0,
			expectedLen:   0,
			expectError:   false,
		},
		{
			name:   "respects pagination",
			limit:  10,
			offset: 5,
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(20)
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{
					"notification_id", "user_id", "title", "message", "notification_type",
					"is_read", "is_deleted", "created_at", "updated_at",
				}).AddRow(
					notificationID.String(), userID.String(), "Title", "Message", "like",
					true, false, now, now,
				)
				mock.ExpectQuery(`SELECT notification_id, user_id, title, message`).
					WithArgs(userID, 10, 5).
					WillReturnRows(rows)
			},
			expectedCount:     20,
			expectedLen:       1,
			expectError:       false,
			expectedFirstID:   notificationID.String(),
			expectedFirstRead: true,
		},
		{
			name:   "returns error on count query failure",
			limit:  20,
			offset: 0,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnError(errTestDB)
			},
			expectError: true,
		},
		{
			name:   "returns error on main query failure",
			limit:  20,
			offset: 0,
			setupMock: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnRows(countRows)

				mock.ExpectQuery(`SELECT notification_id, user_id, title, message`).
					WithArgs(userID, 20, 0).
					WillReturnError(errTestDB)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			defer func() { _ = db.Close() }()

			tt.setupMock(mock)

			repo := repository.NewNotificationRepository(db)
			notifications, count, err := repo.GetNotifications(context.Background(), userID, tt.limit, tt.offset)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
			assert.Len(t, notifications, tt.expectedLen)

			if tt.expectedLen > 0 {
				assert.Equal(t, tt.expectedFirstID, notifications[0].NotificationID)
				assert.Equal(t, tt.expectedFirstRead, notifications[0].IsRead)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

//nolint:funlen // Table-driven test with multiple test cases
func TestSQLNotificationRepository_CountNotifications(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		expectedCount int
		expectError   bool
	}{
		{
			name: "returns count successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(42)
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedCount: 42,
			expectError:   false,
		},
		{
			name: "returns zero when no notifications",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "returns error on database failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(userID).
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			defer func() { _ = db.Close() }()

			tt.setupMock(mock)

			repo := repository.NewNotificationRepository(db)
			count, err := repo.CountNotifications(context.Background(), userID)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
