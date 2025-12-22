package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

var errTestRepo = errors.New("database error")

// MockNotificationRepository is a mock for NotificationRepository.
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) GetNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.Notification, int, error) {
	args := m.Called(ctx, userID, limit, offset)

	notifications, _ := args.Get(0).([]dto.Notification)

	return notifications, args.Int(1), args.Error(2)
}

func (m *MockNotificationRepository) CountNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)

	return args.Int(0), args.Error(1)
}

func (m *MockNotificationRepository) DeleteNotifications(
	ctx context.Context,
	userID uuid.UUID,
	notificationIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID, notificationIDs)

	deletedIDs, _ := args.Get(0).([]uuid.UUID)

	err := args.Error(1)
	if err != nil {
		return deletedIDs, fmt.Errorf("mock error: %w", err)
	}

	return deletedIDs, nil
}

func (m *MockNotificationRepository) MarkNotificationRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationID uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, userID, notificationID)

	err := args.Error(1)
	if err != nil {
		return false, fmt.Errorf("mock error: %w", err)
	}

	return args.Bool(0), nil
}

func (m *MockNotificationRepository) MarkAllNotificationsRead(
	ctx context.Context,
	userID uuid.UUID,
) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)

	readIDs, _ := args.Get(0).([]uuid.UUID)

	err := args.Error(1)
	if err != nil {
		return nil, fmt.Errorf("mock error: %w", err)
	}

	return readIDs, nil
}

//nolint:funlen // Table-driven test with many test cases
func TestNotificationService_GetNotifications(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID := uuid.New()
	now := time.Now()

	sampleNotification := dto.Notification{
		NotificationID:   notificationID.String(),
		UserID:           userID.String(),
		Title:            "Test Title",
		Message:          "Test Message",
		NotificationType: "follow",
		IsRead:           false,
		IsDeleted:        false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		countOnly     bool
		setupMock     func(*MockNotificationRepository)
		expectedType  string
		expectedCount int
		expectedLen   int
		expectError   bool
		checkNilSlice bool
	}{
		{
			name:      "returns list response when countOnly is false",
			limit:     20,
			offset:    0,
			countOnly: false,
			setupMock: func(m *MockNotificationRepository) {
				m.On("GetNotifications", mock.Anything, userID, 20, 0).
					Return([]dto.Notification{sampleNotification}, 10, nil)
			},
			expectedType:  "list",
			expectedCount: 10,
			expectedLen:   1,
			expectError:   false,
		},
		{
			name:      "returns count response when countOnly is true",
			limit:     20,
			offset:    0,
			countOnly: true,
			setupMock: func(m *MockNotificationRepository) {
				m.On("CountNotifications", mock.Anything, userID).
					Return(42, nil)
			},
			expectedType:  "count",
			expectedCount: 42,
			expectError:   false,
		},
		{
			name:      "returns empty slice not nil when no notifications",
			limit:     20,
			offset:    0,
			countOnly: false,
			setupMock: func(m *MockNotificationRepository) {
				m.On("GetNotifications", mock.Anything, userID, 20, 0).
					Return(nil, 0, nil)
			},
			expectedType:  "list",
			expectedCount: 0,
			expectedLen:   0,
			expectError:   false,
			checkNilSlice: true,
		},
		{
			name:      "respects pagination parameters",
			limit:     10,
			offset:    5,
			countOnly: false,
			setupMock: func(m *MockNotificationRepository) {
				m.On("GetNotifications", mock.Anything, userID, 10, 5).
					Return([]dto.Notification{sampleNotification}, 50, nil)
			},
			expectedType:  "list",
			expectedCount: 50,
			expectedLen:   1,
			expectError:   false,
		},
		{
			name:      "returns error when GetNotifications fails",
			limit:     20,
			offset:    0,
			countOnly: false,
			setupMock: func(m *MockNotificationRepository) {
				m.On("GetNotifications", mock.Anything, userID, 20, 0).
					Return(nil, 0, errTestRepo)
			},
			expectError: true,
		},
		{
			name:      "returns error when CountNotifications fails",
			limit:     20,
			offset:    0,
			countOnly: true,
			setupMock: func(m *MockNotificationRepository) {
				m.On("CountNotifications", mock.Anything, userID).
					Return(0, errTestRepo)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockNotificationRepository)
			tt.setupMock(mockRepo)

			svc := service.NewNotificationService(mockRepo)
			result, err := svc.GetNotifications(context.Background(), userID, tt.limit, tt.offset, tt.countOnly)

			if tt.expectError {
				require.Error(t, err)
				mockRepo.AssertExpectations(t)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			switch tt.expectedType {
			case "list":
				listResp, ok := result.(*dto.NotificationListResponse)
				require.True(t, ok, "expected NotificationListResponse type")
				assert.Equal(t, tt.expectedCount, listResp.TotalCount)
				assert.Len(t, listResp.Notifications, tt.expectedLen)
				assert.Equal(t, tt.limit, listResp.Limit)
				assert.Equal(t, tt.offset, listResp.Offset)

				if tt.checkNilSlice {
					assert.NotNil(t, listResp.Notifications, "notifications should be empty slice, not nil")
				}
			case "count":
				countResp, ok := result.(*dto.NotificationCountResponse)
				require.True(t, ok, "expected NotificationCountResponse type")
				assert.Equal(t, tt.expectedCount, countResp.TotalCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

//nolint:funlen // Table-driven test with many test cases
func TestNotificationService_DeleteNotifications(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()
	notificationID3 := uuid.New()

	tests := []struct {
		name             string
		notificationIDs  []string
		setupMock        func(*MockNotificationRepository)
		expectedDeleted  int
		expectedPartial  bool
		expectedNotFound bool
		expectError      bool
	}{
		{
			name:            "returns full success when all IDs deleted",
			notificationIDs: []string{notificationID1.String(), notificationID2.String()},
			setupMock: func(m *MockNotificationRepository) {
				m.On("DeleteNotifications", mock.Anything, userID, mock.Anything).
					Return([]uuid.UUID{notificationID1, notificationID2}, nil)
			},
			expectedDeleted:  2,
			expectedPartial:  false,
			expectedNotFound: false,
			expectError:      false,
		},
		{
			name:            "returns partial success when some IDs not found",
			notificationIDs: []string{notificationID1.String(), notificationID2.String(), notificationID3.String()},
			setupMock: func(m *MockNotificationRepository) {
				m.On("DeleteNotifications", mock.Anything, userID, mock.Anything).
					Return([]uuid.UUID{notificationID1}, nil)
			},
			expectedDeleted:  1,
			expectedPartial:  true,
			expectedNotFound: false,
			expectError:      false,
		},
		{
			name:            "returns all not found when no IDs deleted",
			notificationIDs: []string{notificationID1.String(), notificationID2.String()},
			setupMock: func(m *MockNotificationRepository) {
				m.On("DeleteNotifications", mock.Anything, userID, mock.Anything).
					Return([]uuid.UUID{}, nil)
			},
			expectedDeleted:  0,
			expectedPartial:  false,
			expectedNotFound: true,
			expectError:      false,
		},
		{
			name:             "handles empty input",
			notificationIDs:  []string{},
			setupMock:        func(_ *MockNotificationRepository) {},
			expectedDeleted:  0,
			expectedPartial:  false,
			expectedNotFound: true,
			expectError:      false,
		},
		{
			name:             "handles all invalid UUIDs",
			notificationIDs:  []string{"invalid-uuid-1", "invalid-uuid-2"},
			setupMock:        func(_ *MockNotificationRepository) {},
			expectedDeleted:  0,
			expectedPartial:  false,
			expectedNotFound: true,
			expectError:      false,
		},
		{
			name:            "returns error on repository failure",
			notificationIDs: []string{notificationID1.String()},
			setupMock: func(m *MockNotificationRepository) {
				m.On("DeleteNotifications", mock.Anything, userID, mock.Anything).
					Return(nil, errTestRepo)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockNotificationRepository)
			tt.setupMock(mockRepo)

			svc := service.NewNotificationService(mockRepo)
			result, err := svc.DeleteNotifications(context.Background(), userID, tt.notificationIDs)

			if tt.expectError {
				require.Error(t, err)
				mockRepo.AssertExpectations(t)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.DeletedIDs, tt.expectedDeleted)
			assert.Equal(t, tt.expectedPartial, result.IsPartial)
			assert.Equal(t, tt.expectedNotFound, result.AllNotFound)
			assert.Equal(t, tt.notificationIDs, result.RequestedIDs)
			mockRepo.AssertExpectations(t)
		})
	}
}

//nolint:funlen // Table-driven test with multiple test cases
func TestNotificationService_MarkNotificationRead(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID := uuid.New()

	tests := []struct {
		name        string
		notifIDStr  string
		setupMock   func(*MockNotificationRepository)
		expectedOk  bool
		expectError bool
	}{
		{
			name:       "marks notification as read successfully",
			notifIDStr: notificationID.String(),
			setupMock: func(m *MockNotificationRepository) {
				m.On("MarkNotificationRead", mock.Anything, userID, notificationID).
					Return(true, nil)
			},
			expectedOk:  true,
			expectError: false,
		},
		{
			name:       "returns false when notification not found",
			notifIDStr: notificationID.String(),
			setupMock: func(m *MockNotificationRepository) {
				m.On("MarkNotificationRead", mock.Anything, userID, notificationID).
					Return(false, nil)
			},
			expectedOk:  false,
			expectError: false,
		},
		{
			name:        "returns false for invalid UUID",
			notifIDStr:  "not-a-valid-uuid",
			setupMock:   func(_ *MockNotificationRepository) {},
			expectedOk:  false,
			expectError: false,
		},
		{
			name:       "returns error on repository failure",
			notifIDStr: notificationID.String(),
			setupMock: func(m *MockNotificationRepository) {
				m.On("MarkNotificationRead", mock.Anything, userID, notificationID).
					Return(false, errTestRepo)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockNotificationRepository)
			tt.setupMock(mockRepo)

			svc := service.NewNotificationService(mockRepo)
			ok, err := svc.MarkNotificationRead(context.Background(), userID, tt.notifIDStr)

			if tt.expectError {
				require.Error(t, err)
				mockRepo.AssertExpectations(t)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedOk, ok)
			mockRepo.AssertExpectations(t)
		})
	}
}

//nolint:funlen // Table-driven test with multiple test cases
func TestNotificationService_MarkAllNotificationsRead(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	notificationID1 := uuid.New()
	notificationID2 := uuid.New()
	notificationID3 := uuid.New()

	tests := []struct {
		name        string
		setupMock   func(*MockNotificationRepository)
		expectedLen int
		expectError bool
	}{
		{
			name: "marks all notifications as read successfully",
			setupMock: func(m *MockNotificationRepository) {
				m.On("MarkAllNotificationsRead", mock.Anything, userID).
					Return([]uuid.UUID{notificationID1, notificationID2, notificationID3}, nil)
			},
			expectedLen: 3,
			expectError: false,
		},
		{
			name: "returns empty slice when no unread notifications",
			setupMock: func(m *MockNotificationRepository) {
				m.On("MarkAllNotificationsRead", mock.Anything, userID).
					Return([]uuid.UUID{}, nil)
			},
			expectedLen: 0,
			expectError: false,
		},
		{
			name: "converts UUIDs to strings correctly",
			setupMock: func(m *MockNotificationRepository) {
				m.On("MarkAllNotificationsRead", mock.Anything, userID).
					Return([]uuid.UUID{notificationID1}, nil)
			},
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "returns error on repository failure",
			setupMock: func(m *MockNotificationRepository) {
				m.On("MarkAllNotificationsRead", mock.Anything, userID).
					Return(nil, errTestRepo)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockNotificationRepository)
			tt.setupMock(mockRepo)

			svc := service.NewNotificationService(mockRepo)
			readIDs, err := svc.MarkAllNotificationsRead(context.Background(), userID)

			if tt.expectError {
				require.Error(t, err)
				mockRepo.AssertExpectations(t)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, readIDs)
			assert.Len(t, readIDs, tt.expectedLen)

			// Verify UUIDs are valid strings
			for _, id := range readIDs {
				_, err := uuid.Parse(id)
				require.NoError(t, err, "returned ID should be valid UUID string")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
