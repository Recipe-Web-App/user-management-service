package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
)

// NotificationService defines business logic for notification operations.
type NotificationService interface {
	GetNotifications(
		ctx context.Context,
		userID uuid.UUID,
		limit, offset int,
		countOnly bool,
	) (any, error)
	GetNotificationPreferences(
		ctx context.Context,
		userID uuid.UUID,
	) (*dto.UserPreferences, error)
	DeleteNotifications(
		ctx context.Context,
		userID uuid.UUID,
		notificationIDs []string,
	) (*NotificationDeleteResult, error)
	MarkNotificationRead(
		ctx context.Context,
		userID uuid.UUID,
		notificationID string,
	) (bool, error)
	MarkAllNotificationsRead(
		ctx context.Context,
		userID uuid.UUID,
	) ([]string, error)
	UpdateNotificationPreferences(
		ctx context.Context,
		userID uuid.UUID,
		req *dto.UpdateUserPreferenceRequest,
	) (*dto.UserPreferences, error)
}

// NotificationDeleteResult contains the result of a batch delete operation.
type NotificationDeleteResult struct {
	DeletedIDs   []string // IDs that were successfully deleted
	RequestedIDs []string // IDs that were requested
	IsPartial    bool     // True if some IDs were not deleted (for 206 response)
	AllNotFound  bool     // True if no IDs were deleted (for 404 response)
}

// NotificationServiceImpl implements NotificationService.
type NotificationServiceImpl struct {
	repo     repository.NotificationRepository
	userRepo repository.UserRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(
	repo repository.NotificationRepository,
	userRepo repository.UserRepository,
) *NotificationServiceImpl {
	return &NotificationServiceImpl{
		repo:     repo,
		userRepo: userRepo,
	}
}

// GetNotifications retrieves notifications for a user.
// If countOnly is true, returns NotificationCountResponse.
// Otherwise, returns NotificationListResponse.
func (s *NotificationServiceImpl) GetNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	countOnly bool,
) (any, error) {
	if countOnly {
		count, err := s.repo.CountNotifications(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("count notifications: %w", err)
		}

		return &dto.NotificationCountResponse{
			TotalCount: count,
		}, nil
	}

	notifications, totalCount, err := s.repo.GetNotifications(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get notifications: %w", err)
	}

	// Ensure we return an empty slice, not nil
	if notifications == nil {
		notifications = []dto.Notification{}
	}

	return &dto.NotificationListResponse{
		Notifications: notifications,
		TotalCount:    totalCount,
		Limit:         limit,
		Offset:        offset,
	}, nil
}

// DeleteNotifications soft-deletes notifications for a user.
func (s *NotificationServiceImpl) DeleteNotifications(
	ctx context.Context,
	userID uuid.UUID,
	notificationIDStrings []string,
) (*NotificationDeleteResult, error) {
	// Convert string IDs to UUIDs, skipping invalid ones
	notificationIDs := make([]uuid.UUID, 0, len(notificationIDStrings))

	for _, idStr := range notificationIDStrings {
		id, err := uuid.Parse(idStr)
		if err != nil {
			// Skip invalid UUIDs - validation should catch these upstream
			continue
		}

		notificationIDs = append(notificationIDs, id)
	}

	// Handle case where no valid UUIDs were provided
	if len(notificationIDs) == 0 {
		return &NotificationDeleteResult{
			DeletedIDs:   []string{},
			RequestedIDs: notificationIDStrings,
			IsPartial:    false,
			AllNotFound:  true,
		}, nil
	}

	// Call repository
	deletedUUIDs, err := s.repo.DeleteNotifications(ctx, userID, notificationIDs)
	if err != nil {
		return nil, fmt.Errorf("delete notifications: %w", err)
	}

	// Convert UUIDs back to strings
	deletedIDs := make([]string, len(deletedUUIDs))
	for i, id := range deletedUUIDs {
		deletedIDs[i] = id.String()
	}

	// Determine result type
	return &NotificationDeleteResult{
		DeletedIDs:   deletedIDs,
		RequestedIDs: notificationIDStrings,
		IsPartial:    len(deletedIDs) > 0 && len(deletedIDs) < len(notificationIDStrings),
		AllNotFound:  len(deletedIDs) == 0,
	}, nil
}

// MarkNotificationRead marks a notification as read.
// Returns true if the notification was found and updated, false if not found.
// Returns false without error for invalid UUID (treated as not found).
func (s *NotificationServiceImpl) MarkNotificationRead(
	ctx context.Context,
	userID uuid.UUID,
	notificationIDStr string,
) (bool, error) {
	notificationID, parseErr := uuid.Parse(notificationIDStr)
	if parseErr != nil {
		// Invalid UUID is treated as "not found" rather than an error
		return false, nil //nolint:nilerr // Invalid UUID is intentionally treated as not found
	}

	found, err := s.repo.MarkNotificationRead(ctx, userID, notificationID)
	if err != nil {
		return false, fmt.Errorf("mark notification read: %w", err)
	}

	return found, nil
}

// MarkAllNotificationsRead marks all unread notifications as read for a user.
// Returns the IDs of notifications that were marked as read.
func (s *NotificationServiceImpl) MarkAllNotificationsRead(
	ctx context.Context,
	userID uuid.UUID,
) ([]string, error) {
	readUUIDs, err := s.repo.MarkAllNotificationsRead(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("mark all notifications read: %w", err)
	}

	// Convert UUIDs to strings
	readIDs := make([]string, len(readUUIDs))
	for i, id := range readUUIDs {
		readIDs[i] = id.String()
	}

	return readIDs, nil
}

// GetNotificationPreferences retrieves all notification-related preferences for a user.
func (s *NotificationServiceImpl) GetNotificationPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.UserPreferences, error) {
	// 1. Get Notification Preferences
	notifPrefs, err := s.userRepo.FindNotificationPreferencesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get notification preferences: %w", err)
	}

	// 2. Get Privacy Preferences
	privacyPrefs, err := s.userRepo.FindPrivacyPreferencesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get privacy preferences: %w", err)
	}

	// 3. Get Display Preferences
	displayPrefs, err := s.userRepo.FindDisplayPreferencesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get display preferences: %w", err)
	}

	return &dto.UserPreferences{
		NotificationPreferences: notifPrefs,
		PrivacyPreferences:      privacyPrefs,
		DisplayPreferences:      displayPrefs,
	}, nil
}

// UpdateNotificationPreferences updates user preferences (notification, privacy, display).
func (s *NotificationServiceImpl) UpdateNotificationPreferences(
	ctx context.Context,
	userID uuid.UUID,
	req *dto.UpdateUserPreferenceRequest,
) (*dto.UserPreferences, error) {
	if req.NotificationPreferences != nil {
		err := s.userRepo.UpdateNotificationPreferences(ctx, userID, req.NotificationPreferences)
		if err != nil {
			return nil, fmt.Errorf("update notification preferences: %w", err)
		}
	}

	if req.PrivacyPreferences != nil {
		err := s.userRepo.UpdatePrivacyPreferences(ctx, userID, req.PrivacyPreferences)
		if err != nil {
			return nil, fmt.Errorf("update privacy preferences: %w", err)
		}
	}

	if req.DisplayPreferences != nil {
		err := s.userRepo.UpdateDisplayPreferences(ctx, userID, req.DisplayPreferences)
		if err != nil {
			return nil, fmt.Errorf("update display preferences: %w", err)
		}
	}

	// Return the updated state
	return s.GetNotificationPreferences(ctx, userID)
}
