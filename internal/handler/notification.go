//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// NotificationHandler handles notification HTTP endpoints.
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications handles GET /notifications.
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header (authentication required)
	userID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Parse query parameters
	params, err := h.parseNotificationParams(r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		return
	}

	// 3. Call service
	response, err := h.notificationService.GetNotifications(
		r.Context(),
		userID,
		params.limit,
		params.offset,
		params.countOnly,
	)
	if err != nil {
		InternalErrorResponse(w)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
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

func (h *NotificationHandler) extractAuthenticatedUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userIDStr := r.Header.Get("X-User-Id")
	if userIDStr == "" {
		UnauthorizedResponse(w, "User authentication required")

		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		UnauthorizedResponse(w, "Invalid user ID in authentication header")

		return uuid.Nil, false
	}

	return userID, true
}

// notificationParams holds parsed query parameters for GetNotifications.
type notificationParams struct {
	limit     int
	offset    int
	countOnly bool
}

// Notification parameter validation errors.
var (
	ErrNotificationInvalidLimit     = errors.New("limit must be a valid integer")
	ErrNotificationLimitOutOfRange  = errors.New("limit must be between 1 and 100")
	ErrNotificationInvalidOffset    = errors.New("offset must be a valid integer")
	ErrNotificationNegativeOffset   = errors.New("offset must be non-negative")
	ErrNotificationInvalidCountOnly = errors.New("count_only must be a valid boolean")
)

//nolint:dupl // Intentionally mirrors parseFollowingParams pattern for consistency
func (h *NotificationHandler) parseNotificationParams(r *http.Request) (*notificationParams, error) {
	params := &notificationParams{
		limit:     defaultLimit,
		offset:    0,
		countOnly: false,
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, ErrNotificationInvalidLimit
		}

		if limit < minLimit || limit > maxLimit {
			return nil, ErrNotificationLimitOutOfRange
		}

		params.limit = limit
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, ErrNotificationInvalidOffset
		}

		if offset < 0 {
			return nil, ErrNotificationNegativeOffset
		}

		params.offset = offset
	}

	// Parse count_only
	if countOnlyStr := r.URL.Query().Get("count_only"); countOnlyStr != "" {
		countOnly, err := strconv.ParseBool(countOnlyStr)
		if err != nil {
			return nil, ErrNotificationInvalidCountOnly
		}

		params.countOnly = countOnly
	}

	return params, nil
}
