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

// Pagination constants.
const (
	defaultLimit = 20
	maxLimit     = 100
	minLimit     = 1
)

// Search parameter validation errors.
var (
	ErrInvalidLimit     = errors.New("limit must be a valid integer")
	ErrLimitOutOfRange  = errors.New("limit must be between 1 and 100")
	ErrInvalidOffset    = errors.New("offset must be a valid integer")
	ErrNegativeOffset   = errors.New("offset must be non-negative")
	ErrInvalidCountOnly = errors.New("countOnly must be a valid boolean")
)

// UserHandler handles user-related HTTP endpoints.
type UserHandler struct {
	userService service.UserService
	binder      *RequestBinder
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		binder:      NewRequestBinder(),
	}
}

// GetUserProfile handles GET /users/{user_id}/profile.
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID from path
	userIDStr := chi.URLParam(r, "user_id")

	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	// 2. Identify Requester
	// In a real scenario, this comes from context set by Auth Middleware.
	// We check X-User-ID header or assume anonymous.
	var requesterID uuid.UUID

	requesterIDStr := r.Header.Get("X-User-Id")
	if requesterIDStr != "" {
		id, err := uuid.Parse(requesterIDStr)
		if err == nil {
			requesterID = id
		}
	}
	// If empty, requesterID is zero-value UUID (Anonymous)

	// 3. Call Service
	profile, err := h.userService.GetUserProfile(r.Context(), requesterID, targetUserID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}

		if errors.Is(err, service.ErrProfilePrivate) {
			ErrorResponse(w, http.StatusForbidden, "PROFILE_PRIVATE", "Profile is private")
			return
		}

		ErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve profile")

		return
	}

	SuccessResponse(w, http.StatusOK, profile)
}

// UpdateUserProfile handles PUT /users/profile.
func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	var req dto.UserProfileUpdateRequest

	bindErr := h.binder.BindAndValidate(r, &req)
	if bindErr != nil {
		h.handleBindError(w, bindErr)

		return
	}

	profile, err := h.userService.UpdateUserProfile(r.Context(), requesterID, &req)
	if err != nil {
		h.handleUpdateProfileError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, profile)
}

// RequestAccountDeletion handles POST /users/account/delete-request.
func (h *UserHandler) RequestAccountDeletion(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	response, err := h.userService.RequestAccountDeletion(r.Context(), requesterID)
	if err != nil {
		h.handleDeleteRequestError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// ConfirmAccountDeletion handles DELETE /users/account.
func (h *UserHandler) ConfirmAccountDeletion(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	var req dto.UserAccountDeleteRequest

	bindErr := h.binder.BindAndValidate(r, &req)
	if bindErr != nil {
		h.handleBindError(w, bindErr)

		return
	}

	response, err := h.userService.ConfirmAccountDeletion(r.Context(), requesterID, req.ConfirmationToken)
	if err != nil {
		h.handleConfirmDeletionError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// SearchUsers handles GET /users/search.
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	// 1. Require authentication
	_, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Parse query parameters
	params, err := h.parseSearchParams(r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		return
	}

	// 3. Call service
	response, err := h.userService.SearchUsers(r.Context(), params.query, params.limit, params.offset, params.countOnly)
	if err != nil {
		h.handleSearchError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// GetUserByID handles GET /users/{user_id}.
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID from path
	userIDStr := chi.URLParam(r, "user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid user ID format")

		return
	}

	// 2. Call Service
	result, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		h.handleGetUserByIDError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, result)
}

type searchParams struct {
	query     string
	limit     int
	offset    int
	countOnly bool
}

func (h *UserHandler) parseSearchParams(r *http.Request) (*searchParams, error) {
	params := &searchParams{
		query:     r.URL.Query().Get("query"),
		limit:     defaultLimit,
		offset:    0,
		countOnly: false,
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, ErrInvalidLimit
		}

		if limit < minLimit || limit > maxLimit {
			return nil, ErrLimitOutOfRange
		}

		params.limit = limit
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, ErrInvalidOffset
		}

		if offset < 0 {
			return nil, ErrNegativeOffset
		}

		params.offset = offset
	}

	// Parse countOnly
	if countOnlyStr := r.URL.Query().Get("countOnly"); countOnlyStr != "" {
		countOnly, err := strconv.ParseBool(countOnlyStr)
		if err != nil {
			return nil, ErrInvalidCountOnly
		}

		params.countOnly = countOnly
	}

	return params, nil
}

func (h *UserHandler) handleSearchError(w http.ResponseWriter, _ error) {
	// For now, any error from the service is an internal error
	// We can add more specific error handling as needed
	InternalErrorResponse(w)
}

func (h *UserHandler) handleGetUserByIDError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	default:
		InternalErrorResponse(w)
	}
}

func (h *UserHandler) extractAuthenticatedUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	requesterIDStr := r.Header.Get("X-User-Id")
	if requesterIDStr == "" {
		UnauthorizedResponse(w, "User authentication required")

		return uuid.Nil, false
	}

	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		UnauthorizedResponse(w, "Invalid user ID in authentication header")

		return uuid.Nil, false
	}

	return requesterID, true
}

func (h *UserHandler) handleBindError(w http.ResponseWriter, err error) {
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

func (h *UserHandler) handleUpdateProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrDuplicateUsername):
		ConflictResponse(w, "Username already taken")
	default:
		slog.Error("failed to update user profile", "error", err)
		InternalErrorResponse(w)
	}
}

func (h *UserHandler) handleDeleteRequestError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrCacheUnavailable):
		ServiceUnavailableResponse(w, "Service temporarily unavailable")
	default:
		slog.Error("failed to delete user request", "error", err)
		InternalErrorResponse(w)
	}
}

func (h *UserHandler) handleConfirmDeletionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidToken):
		ErrorResponse(w, http.StatusBadRequest, "INVALID_TOKEN", "Invalid or expired confirmation token")
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrCacheUnavailable):
		ServiceUnavailableResponse(w, "Service temporarily unavailable")
	default:
		slog.Error("failed to confirm user deletion", "error", err)
		InternalErrorResponse(w)
	}
}
