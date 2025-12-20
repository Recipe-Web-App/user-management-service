package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// SocialHandler handles social feature HTTP endpoints.
type SocialHandler struct {
	socialService service.SocialService
}

// NewSocialHandler creates a new social handler.
func NewSocialHandler(socialService service.SocialService) *SocialHandler {
	return &SocialHandler{
		socialService: socialService,
	}
}

// GetFollowing handles GET /users/{user_id}/following.
func (h *SocialHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header
	requesterID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Extract and validate target user ID from path
	userIDStr := chi.URLParam(r, "user_id")

	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid user ID format")

		return
	}

	// 3. Parse query parameters
	params, err := h.parseFollowingParams(r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		return
	}

	// 4. Call service
	response, err := h.socialService.GetFollowing(
		r.Context(),
		requesterID,
		targetUserID,
		params.limit,
		params.offset,
		params.countOnly,
	)
	if err != nil {
		h.handleGetFollowingError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// GetFollowers handles GET /users/{user_id}/followers.
func (h *SocialHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header
	requesterID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Extract and validate target user ID from path
	userIDStr := chi.URLParam(r, "user_id")

	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid user ID format")

		return
	}

	// 3. Parse query parameters
	params, err := h.parseFollowingParams(r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		return
	}

	// 4. Call service
	response, err := h.socialService.GetFollowers(
		r.Context(),
		requesterID,
		targetUserID,
		params.limit,
		params.offset,
		params.countOnly,
	)
	if err != nil {
		h.handleGetFollowersError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// FollowUser handles POST /users/{user_id}/follow/{target_user_id}.
func (h *SocialHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header (authenticated user)
	requesterID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Extract and validate user_id from path (the user performing the follow)
	userIDStr := chi.URLParam(r, "user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid user ID format")

		return
	}

	// 3. Authorization check: path user_id must match authenticated user OR user is admin
	isAdmin := h.isAdminUser(r)
	if userID != requesterID && !isAdmin {
		ForbiddenResponse(w, "Cannot perform follow action for another user")

		return
	}

	// 4. Extract and validate target_user_id from path
	targetUserIDStr := chi.URLParam(r, "target_user_id")

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid target user ID format")

		return
	}

	// 5. Call service (use path user_id as follower, not requester, for admin override)
	response, err := h.socialService.FollowUser(r.Context(), userID, targetUserID)
	if err != nil {
		h.handleFollowUserError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// UnfollowUser handles DELETE /users/{user_id}/follow/{target_user_id}.
// Stub implementation - to be implemented.
func (h *SocialHandler) UnfollowUser(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.FollowResponse{
		Message:     "Successfully unfollowed user",
		IsFollowing: false,
	})
}

// GetUserActivity handles GET /users/{user_id}/activity.
// Stub implementation - to be implemented.
//
//nolint:mnd // placeholder values for stub handler
func (h *SocialHandler) GetUserActivity(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()

	SuccessResponse(w, http.StatusOK, dto.UserActivityResponse{
		UserID: uuid.New().String(),
		RecentRecipes: []dto.RecipeSummary{
			{
				RecipeID:  1,
				Title:     "Spaghetti Carbonara",
				CreatedAt: now,
			},
		},
		RecentFollows: []dto.UserSummary{
			{
				UserID:     uuid.New().String(),
				Username:   "chefmike",
				FollowedAt: now,
			},
		},
		RecentReviews: []dto.ReviewSummary{
			{
				ReviewID:  1,
				RecipeID:  2,
				Rating:    4.5,
				CreatedAt: now,
			},
		},
		RecentFavorites: []dto.FavoriteSummary{
			{
				RecipeID:    3,
				Title:       "Chocolate Cake",
				FavoritedAt: now,
			},
		},
	})
}

// Private helper types and methods below.

type followingParams struct {
	limit     int
	offset    int
	countOnly bool
}

func (h *SocialHandler) parseFollowingParams(r *http.Request) (*followingParams, error) {
	params := &followingParams{
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

	// Parse count_only
	if countOnlyStr := r.URL.Query().Get("count_only"); countOnlyStr != "" {
		countOnly, err := strconv.ParseBool(countOnlyStr)
		if err != nil {
			return nil, ErrInvalidCountOnly
		}

		params.countOnly = countOnly
	}

	return params, nil
}

func (h *SocialHandler) handleGetFollowingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrAccessDenied):
		ForbiddenResponse(w, "Access to this user's following list is restricted")
	default:
		InternalErrorResponse(w)
	}
}

func (h *SocialHandler) handleGetFollowersError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrAccessDenied):
		ForbiddenResponse(w, "Access to this user's followers list is restricted")
	default:
		InternalErrorResponse(w)
	}
}

func (h *SocialHandler) extractAuthenticatedUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
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

func (h *SocialHandler) isAdminUser(r *http.Request) bool {
	role := r.Header.Get("X-User-Role")

	return role == "admin"
}

func (h *SocialHandler) handleFollowUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCannotFollowSelf):
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Cannot follow yourself")
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrFollowNotAllowed):
		ForbiddenResponse(w, "This user does not allow follows")
	default:
		InternalErrorResponse(w)
	}
}
