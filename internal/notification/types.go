// Package notification provides client functionality for the notification service.
package notification

import "errors"

// NewFollowerRequest represents the payload for POST /notifications/new-follower.
//
//nolint:tagliatelle // API spec requires snake_case
type NewFollowerRequest struct {
	RecipientIDs []string `json:"recipient_ids"`
	FollowerID   string   `json:"follower_id"`
}

// EmailChangedRequest represents the payload for POST /notifications/email-changed.
//
//nolint:tagliatelle // API spec requires snake_case
type EmailChangedRequest struct {
	RecipientIDs []string `json:"recipient_ids"`
	OldEmail     string   `json:"old_email"`
	NewEmail     string   `json:"new_email"`
}

// BatchNotificationResponse represents the response from notification endpoints.
//
//nolint:tagliatelle // API spec requires snake_case
type BatchNotificationResponse struct {
	Notifications []NotificationRef `json:"notifications"`
	QueuedCount   int               `json:"queued_count"`
	Message       string            `json:"message"`
}

// NotificationRef represents a created notification reference.
//
//nolint:tagliatelle // API spec requires snake_case
type NotificationRef struct {
	NotificationID string `json:"notification_id"`
	RecipientID    string `json:"recipient_id"`
}

// ErrorResponse represents an error from the notification service.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Sentinel errors for notification operations.
var (
	ErrNotificationDisabled   = errors.New("notification service is disabled")
	ErrNotificationFailed     = errors.New("notification request failed")
	ErrNotificationBadRequest = errors.New("invalid notification request")
	ErrNotificationForbidden  = errors.New("notification forbidden")
	ErrNotificationRateLimit  = errors.New("notification rate limit exceeded")
)
