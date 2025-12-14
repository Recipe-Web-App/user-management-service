//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// NotificationHandler handles notification HTTP endpoints.
type NotificationHandler struct{}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// GetNotifications handles GET /notifications.
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()

	SuccessResponse(w, http.StatusOK, dto.NotificationListResponse{
		Notifications: []dto.Notification{
			{
				NotificationID:   uuid.New().String(),
				UserID:           uuid.New().String(),
				Title:            "New follower",
				Message:          "John Doe started following you",
				NotificationType: "follow",
				IsRead:           false,
				IsDeleted:        false,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
		TotalCount: 1,
		Limit:      20,
		Offset:     0,
	})
}

// DeleteNotifications handles DELETE /notifications.
func (h *NotificationHandler) DeleteNotifications(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.NotificationDeleteResponse{
		Message:                "Notifications deleted successfully",
		DeletedNotificationIDs: []string{uuid.New().String()},
	})
}

// MarkNotificationRead handles PUT /notifications/{notification_id}/read.
func (h *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.NotificationReadResponse{
		Message: "Notification marked as read successfully",
	})
}

// MarkAllNotificationsRead handles PUT /notifications/read-all.
func (h *NotificationHandler) MarkAllNotificationsRead(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.NotificationReadAllResponse{
		Message:             "All notifications marked as read successfully",
		ReadNotificationIDs: []string{uuid.New().String(), uuid.New().String()},
	})
}

// GetNotificationPreferences handles GET /notifications/preferences.
func (h *NotificationHandler) GetNotificationPreferences(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.UserPreferenceResponse{
		Preferences: dto.UserPreferences{
			NotificationPreferences: &dto.NotificationPreferences{
				EmailNotifications:   true,
				PushNotifications:    true,
				FollowNotifications:  true,
				LikeNotifications:    true,
				CommentNotifications: true,
				RecipeNotifications:  true,
				SystemNotifications:  true,
			},
			PrivacyPreferences: &dto.PrivacyPreferences{
				ProfileVisibility: "public",
				ShowEmail:         false,
				ShowFullName:      true,
				AllowFollows:      true,
				AllowMessages:     true,
			},
			DisplayPreferences: &dto.DisplayPreferences{
				Theme:    "auto",
				Language: "en",
				Timezone: "UTC",
			},
		},
	})
}

// UpdateNotificationPreferences handles PUT /notifications/preferences.
func (h *NotificationHandler) UpdateNotificationPreferences(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.UserPreferenceResponse{
		Preferences: dto.UserPreferences{
			NotificationPreferences: &dto.NotificationPreferences{
				EmailNotifications:   true,
				PushNotifications:    false,
				FollowNotifications:  true,
				LikeNotifications:    true,
				CommentNotifications: true,
				RecipeNotifications:  true,
				SystemNotifications:  true,
			},
			PrivacyPreferences: &dto.PrivacyPreferences{
				ProfileVisibility: "public",
				ShowEmail:         false,
				ShowFullName:      true,
				AllowFollows:      true,
				AllowMessages:     true,
			},
			DisplayPreferences: &dto.DisplayPreferences{
				Theme:    "dark",
				Language: "en",
				Timezone: "UTC",
			},
		},
	})
}
