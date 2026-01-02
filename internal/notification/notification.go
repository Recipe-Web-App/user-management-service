package notification

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

const (
	pathNewFollower  = "/notifications/new-follower"
	pathEmailChanged = "/notifications/email-changed"
)

// Client defines the interface for notification operations.
type Client interface {
	// NotifyNewFollower sends a notification when a user follows another user.
	// This is a fire-and-forget operation that logs errors but does not return them.
	NotifyNewFollower(ctx context.Context, recipientID, followerID uuid.UUID)

	// NotifyEmailChanged sends a security notification when a user changes their email.
	// This is a fire-and-forget operation that logs errors but does not return them.
	NotifyEmailChanged(ctx context.Context, recipientID uuid.UUID, oldEmail, newEmail string)
}

// NotificationClient implements Client using the notification service API.
type NotificationClient struct {
	client DownstreamClient
	logger *slog.Logger
}

// NewNotificationClient creates a new NotificationClient.
func NewNotificationClient(client DownstreamClient) *NotificationClient {
	return &NotificationClient{
		client: client,
		logger: slog.Default(),
	}
}

// NotifyNewFollower sends a notification when a user follows another user.
// This operation is fire-and-forget - errors are logged but not returned.
func (c *NotificationClient) NotifyNewFollower(ctx context.Context, recipientID, followerID uuid.UUID) {
	req := NewFollowerRequest{
		RecipientIDs: []string{recipientID.String()},
		FollowerID:   followerID.String(),
	}

	var resp BatchNotificationResponse

	err := c.client.Do(ctx, http.MethodPost, pathNewFollower, req, &resp)
	if err != nil {
		c.logger.Warn("failed to send new follower notification",
			"recipient_id", recipientID,
			"follower_id", followerID,
			"error", err,
		)

		return
	}

	c.logger.Debug("new follower notification sent",
		"recipient_id", recipientID,
		"follower_id", followerID,
		"queued_count", resp.QueuedCount,
	)
}

// NotifyEmailChanged sends a security notification when a user changes their email.
// This operation is fire-and-forget - errors are logged but not returned.
func (c *NotificationClient) NotifyEmailChanged(
	ctx context.Context,
	recipientID uuid.UUID,
	oldEmail, newEmail string,
) {
	req := EmailChangedRequest{
		RecipientIDs: []string{recipientID.String()},
		OldEmail:     oldEmail,
		NewEmail:     newEmail,
	}

	var resp BatchNotificationResponse

	err := c.client.Do(ctx, http.MethodPost, pathEmailChanged, req, &resp)
	if err != nil {
		c.logger.Warn("failed to send email changed notification",
			"recipient_id", recipientID,
			"error", err,
		)

		return
	}

	c.logger.Debug("email changed notification sent",
		"recipient_id", recipientID,
		"queued_count", resp.QueuedCount,
	)
}

// NoopClient is a no-op implementation for when notifications are disabled.
type NoopClient struct{}

// NotifyNewFollower is a no-op.
func (c *NoopClient) NotifyNewFollower(_ context.Context, _, _ uuid.UUID) {}

// NotifyEmailChanged is a no-op.
func (c *NoopClient) NotifyEmailChanged(_ context.Context, _ uuid.UUID, _, _ string) {}
