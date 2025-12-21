package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// NotificationRepository defines the interface for notification data access.
type NotificationRepository interface {
	GetNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.Notification, int, error)
	CountNotifications(ctx context.Context, userID uuid.UUID) (int, error)
}

// SQLNotificationRepository implements NotificationRepository using a SQL database.
type SQLNotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new SQLNotificationRepository.
func NewNotificationRepository(db *sql.DB) *SQLNotificationRepository {
	return &SQLNotificationRepository{db: db}
}

// GetNotifications retrieves notifications for a user with pagination and returns total count.
func (r *SQLNotificationRepository) GetNotifications(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.Notification, int, error) {
	// First, get the total count
	totalCount, err := r.CountNotifications(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	// Then, get the paginated notifications
	query := `
		SELECT notification_id, user_id, title, message, notification_type,
		       is_read, is_deleted, created_at, updated_at
		FROM recipe_manager.notifications
		WHERE user_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query notifications: %w", err)
	}

	defer func() { _ = rows.Close() }()

	notifications := make([]dto.Notification, 0)

	for rows.Next() {
		var n dto.Notification

		err := rows.Scan(
			&n.NotificationID,
			&n.UserID,
			&n.Title,
			&n.Message,
			&n.NotificationType,
			&n.IsRead,
			&n.IsDeleted,
			&n.CreatedAt,
			&n.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}

		notifications = append(notifications, n)
	}

	err = rows.Err()
	if err != nil {
		return nil, 0, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, totalCount, nil
}

// CountNotifications returns the count of non-deleted notifications for a user.
func (r *SQLNotificationRepository) CountNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM recipe_manager.notifications
		WHERE user_id = $1 AND is_deleted = false
	`

	var count int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count notifications: %w", err)
	}

	return count, nil
}
