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
	DeleteNotifications(
		ctx context.Context,
		userID uuid.UUID,
		notificationIDs []string,
	) (*NotificationDeleteResult, error)
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
	repo repository.NotificationRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(repo repository.NotificationRepository) *NotificationServiceImpl {
	return &NotificationServiceImpl{
		repo: repo,
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
