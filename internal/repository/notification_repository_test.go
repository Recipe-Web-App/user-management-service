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

//nolint:funlen // Table-driven test with many test cases
func TestSQLNotificationRepository_DeleteNotifications(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()
	notificationID3 := uuid.New()

	tests := []struct {
		name            string
		notificationIDs []uuid.UUID
		setupMock       func(mock sqlmock.Sqlmock)
		expectedIDs     []uuid.UUID
		expectError     bool
	}{
		{
			name:            "deletes all notifications successfully",
			notificationIDs: []uuid.UUID{notificationID1, notificationID2},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"notification_id"}).
					AddRow(notificationID1).
					AddRow(notificationID2)
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedIDs: []uuid.UUID{notificationID1, notificationID2},
			expectError: false,
		},
		{
			name:            "returns only found notifications on partial delete",
			notificationIDs: []uuid.UUID{notificationID1, notificationID2, notificationID3},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Only notificationID1 is found and deleted
				rows := sqlmock.NewRows([]string{"notification_id"}).
					AddRow(notificationID1)
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedIDs: []uuid.UUID{notificationID1},
			expectError: false,
		},
		{
			name:            "returns empty slice when no notifications found",
			notificationIDs: []uuid.UUID{notificationID1, notificationID2},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"notification_id"})
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedIDs: []uuid.UUID{},
			expectError: false,
		},
		{
			name:            "returns empty slice for empty input",
			notificationIDs: []uuid.UUID{},
			setupMock:       func(_ sqlmock.Sqlmock) {},
			expectedIDs:     []uuid.UUID{},
			expectError:     false,
		},
		{
			name:            "returns error on database failure",
			notificationIDs: []uuid.UUID{notificationID1},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID, sqlmock.AnyArg()).
					WillReturnError(errTestDB)
			},
			expectError: true,
		},
		{
			name:            "returns error on row scan failure",
			notificationIDs: []uuid.UUID{notificationID1},
			setupMock: func(mock sqlmock.Sqlmock) {
				// Return invalid data that can't be scanned as UUID
				rows := sqlmock.NewRows([]string{"notification_id"}).
					AddRow("not-a-uuid")
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID, sqlmock.AnyArg()).
					WillReturnRows(rows)
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
			deletedIDs, err := repo.DeleteNotifications(context.Background(), userID, tt.notificationIDs)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedIDs, deletedIDs)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

//nolint:funlen // Table-driven test with multiple test cases
func TestSQLNotificationRepository_MarkNotificationRead(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID := uuid.New()

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectedOk  bool
		expectError bool
	}{
		{
			name: "marks notification as read successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"notification_id"}).AddRow(notificationID)
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(notificationID, userID).
					WillReturnRows(rows)
			},
			expectedOk:  true,
			expectError: false,
		},
		{
			name: "returns false when notification not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(notificationID, userID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedOk:  false,
			expectError: false,
		},
		{
			name: "returns false when notification belongs to different user",
			setupMock: func(mock sqlmock.Sqlmock) {
				// No rows returned because user_id doesn't match
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(notificationID, userID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedOk:  false,
			expectError: false,
		},
		{
			name: "returns error on database failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(notificationID, userID).
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
			ok, err := repo.MarkNotificationRead(context.Background(), userID, notificationID)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedOk, ok)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

//nolint:funlen // Table-driven test with multiple test cases
func TestSQLNotificationRepository_MarkAllNotificationsRead(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()
	notificationID3 := uuid.New()

	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectedIDs []uuid.UUID
		expectError bool
	}{
		{
			name: "marks all unread notifications as read",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"notification_id"}).
					AddRow(notificationID1).
					AddRow(notificationID2).
					AddRow(notificationID3)
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedIDs: []uuid.UUID{notificationID1, notificationID2, notificationID3},
			expectError: false,
		},
		{
			name: "returns empty slice when no unread notifications",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"notification_id"})
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedIDs: []uuid.UUID{},
			expectError: false,
		},
		{
			name: "returns error on database failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID).
					WillReturnError(errTestDB)
			},
			expectError: true,
		},
		{
			name: "returns error on row scan failure",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"notification_id"}).
					AddRow("not-a-uuid")
				mock.ExpectQuery(`UPDATE recipe_manager.notifications`).
					WithArgs(userID).
					WillReturnRows(rows)
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
			readIDs, err := repo.MarkAllNotificationsRead(context.Background(), userID)

			if tt.expectError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedIDs, readIDs)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
