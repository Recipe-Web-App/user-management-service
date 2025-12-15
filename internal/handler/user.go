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
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
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
func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	fullName := placeholderFullName
	bio := "Updated bio"

	SuccessResponse(w, http.StatusOK, dto.UserProfileResponse{
		UserID:    uuid.New().String(),
		Username:  placeholderUsername,
		FullName:  &fullName,
		Bio:       &bio,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	})
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
