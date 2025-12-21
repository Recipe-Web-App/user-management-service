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
