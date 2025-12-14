package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// Placeholder constants for stub responses.
const (
	placeholderFullName = "John Doe"
	placeholderUsername = "johndoe"
)

// UserHandler handles user-related HTTP endpoints.
type UserHandler struct{}

// NewUserHandler creates a new user handler.
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// GetUserProfile handles GET /users/{user_id}/profile.
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	fullName := placeholderFullName
	bio := "Food enthusiast and home chef"

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
