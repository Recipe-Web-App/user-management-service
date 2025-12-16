package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// Placeholder constants for stub responses.
const (
	placeholderFullName = "John Doe"
	placeholderUsername = "johndoe"
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
func (h *UserHandler) RequestAccountDeletion(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.UserAccountDeleteRequestResponse{
		UserID:            uuid.New().String(),
		ConfirmationToken: uuid.New().String(),
		ExpiresAt:         time.Now().Add(24 * time.Hour), //nolint:mnd // placeholder expiry
	})
}

// ConfirmAccountDeletion handles DELETE /users/account.
func (h *UserHandler) ConfirmAccountDeletion(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.UserConfirmAccountDeleteResponse{
		UserID:        uuid.New().String(),
		DeactivatedAt: time.Now(),
	})
}

// SearchUsers handles GET /users/search.
func (h *UserHandler) SearchUsers(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	fullName := placeholderFullName

	SuccessResponse(w, http.StatusOK, dto.UserSearchResponse{
		Results: []dto.UserSearchResult{
			{
				UserID:    uuid.New().String(),
				Username:  placeholderUsername,
				FullName:  &fullName,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		TotalCount: 1,
		Limit:      20, //nolint:mnd // placeholder pagination
		Offset:     0,
	})
}

// GetUserByID handles GET /users/{user_id}.
func (h *UserHandler) GetUserByID(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	fullName := placeholderFullName

	SuccessResponse(w, http.StatusOK, dto.UserSearchResult{
		UserID:    uuid.New().String(),
		Username:  placeholderUsername,
		FullName:  &fullName,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	})
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
		InternalErrorResponse(w)
	}
}
