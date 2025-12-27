//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// NotificationHandler handles notification HTTP endpoints.
type NotificationHandler struct {
	notificationService service.NotificationService
	binder              *RequestBinder
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		binder:              NewRequestBinder(),
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
func (h *NotificationHandler) DeleteNotifications(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header
	userID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Parse and validate request body
	var req dto.NotificationDeleteRequest

	bindErr := h.binder.BindAndValidate(r, &req)
	if bindErr != nil {
		h.handleBindError(w, bindErr)

		return
	}

	// 3. Call service
	result, err := h.notificationService.DeleteNotifications(r.Context(), userID, req.NotificationIDs)
	if err != nil {
		InternalErrorResponse(w)

		return
	}

	// 4. Build response
	response := dto.NotificationDeleteResponse{
		DeletedNotificationIDs: result.DeletedIDs,
	}

	// Ensure empty slice not nil for JSON serialization
	if response.DeletedNotificationIDs == nil {
		response.DeletedNotificationIDs = []string{}
	}

	// 5. Determine status code and message based on result
	switch {
	case result.AllNotFound:
		NotFoundResponse(w, "Notifications")
	case result.IsPartial:
		response.Message = "Some notifications deleted successfully"
		SuccessResponse(w, http.StatusPartialContent, response)
	default:
		response.Message = "Notifications deleted successfully"
		SuccessResponse(w, http.StatusOK, response)
	}
}

// MarkNotificationRead handles PUT /notifications/{notification_id}/read.
func (h *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header
	userID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Extract notification_id from path
	notificationIDStr := chi.URLParam(r, "notification_id")
	if notificationIDStr == "" {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "notification_id is required")

		return
	}

	// 3. Validate notification_id is a valid UUID
	_, parseErr := uuid.Parse(notificationIDStr)
	if parseErr != nil {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "notification_id must be a valid UUID")

		return
	}

	// 4. Call service
	found, err := h.notificationService.MarkNotificationRead(r.Context(), userID, notificationIDStr)
	if err != nil {
		InternalErrorResponse(w)

		return
	}

	// 5. Return 404 if not found, 200 with success message if found
	if !found {
		NotFoundResponse(w, "Notification")

		return
	}

	SuccessResponse(w, http.StatusOK, dto.NotificationReadResponse{
		Message: "Notification marked as read successfully",
	})
}

// MarkAllNotificationsRead handles PUT /notifications/read-all.
func (h *NotificationHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header
	userID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Call service
	readIDs, err := h.notificationService.MarkAllNotificationsRead(r.Context(), userID)
	if err != nil {
		InternalErrorResponse(w)

		return
	}

	// 3. Ensure empty slice not nil for JSON serialization
	if readIDs == nil {
		readIDs = []string{}
	}

	// 4. Return success response
	SuccessResponse(w, http.StatusOK, dto.NotificationReadAllResponse{
		Message:             "All notifications marked as read successfully",
		ReadNotificationIDs: readIDs,
	})
}

// GetNotificationPreferences handles GET /notifications/preferences.
// GetNotificationPreferences handles GET /notifications/preferences.
func (h *NotificationHandler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	prefs, err := h.notificationService.GetNotificationPreferences(r.Context(), userID)
	if err != nil {
		InternalErrorResponse(w)
		slog.Error("failed to get notification preferences", "error", err)
		return
	}

	SuccessResponse(w, http.StatusOK, dto.UserPreferenceResponse{
		Preferences: *prefs,
	})
}

func (h *NotificationHandler) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header
	userID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Parse and validate request body
	var req dto.UpdateUserPreferenceRequest

	bindErr := h.binder.BindAndValidate(r, &req)
	if bindErr != nil {
		h.handleBindError(w, bindErr)

		return
	}

	// 3. Call service
	prefs, err := h.notificationService.UpdateNotificationPreferences(r.Context(), userID, &req)
	if err != nil {
		InternalErrorResponse(w)

		return
	}

	// 4. Return updated preferences
	SuccessResponse(w, http.StatusOK, dto.UserPreferenceResponse{
		Preferences: *prefs,
	})
}

func (h *NotificationHandler) handleBindError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmptyBody):
		ErrorResponse(w, http.StatusBadRequest, "EMPTY_BODY", "Request body is required")
	case errors.Is(err, ErrInvalidJSON), errors.Is(err, ErrInvalidFieldType):
		ErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
	case errors.Is(err, ErrValidationFailed):
		ValidationErrorResponse(w, err)
	default:
		slog.Error("failed to bind request body", "error", err)
		ErrorResponse(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	}
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
	ErrNotificationInvalidCountOnly = errors.New("countOnly must be a valid boolean")
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

	// Parse countOnly
	if countOnlyStr := r.URL.Query().Get("countOnly"); countOnlyStr != "" {
		countOnly, err := strconv.ParseBool(countOnlyStr)
		if err != nil {
			return nil, ErrNotificationInvalidCountOnly
		}

		params.countOnly = countOnly
	}

	return params, nil
}
