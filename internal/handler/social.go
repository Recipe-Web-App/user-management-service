package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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
//
//nolint:dupl // Intentionally mirrors FollowUser pattern for consistency
func (h *SocialHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate requester ID from header (authenticated user)
	requesterID, ok := h.extractAuthenticatedUserID(w, r)
	if !ok {
		return
	}

	// 2. Extract and validate user_id from path (the user performing the unfollow)
	userIDStr := chi.URLParam(r, "user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid user ID format")

		return
	}

	// 3. Authorization check: path user_id must match authenticated user OR user is admin
	isAdmin := h.isAdminUser(r)
	if userID != requesterID && !isAdmin {
		ForbiddenResponse(w, "Cannot perform unfollow action for another user")

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
	response, err := h.socialService.UnfollowUser(r.Context(), userID, targetUserID)
	if err != nil {
		h.handleUnfollowUserError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// Activity parameter constants.
const (
	defaultPerTypeLimit = 15
	minPerTypeLimit     = 1
	maxPerTypeLimit     = 100
)

// Activity parameter validation errors.
var (
	ErrInvalidPerTypeLimit    = errors.New("per_type_limit must be a valid integer")
	ErrPerTypeLimitOutOfRange = errors.New("per_type_limit must be between 1 and 100")
)

// GetUserActivity handles GET /users/{user_id}/activity.
func (h *SocialHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	// 1. Extract optional requester ID from header (anonymous access allowed)
	requesterID := h.extractOptionalUserID(r)

	// 2. Extract and validate target user ID from path
	userIDStr := chi.URLParam(r, "user_id")

	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid user ID format")

		return
	}

	// 3. Parse query parameters
	perTypeLimit, err := h.parseActivityParams(r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())

		return
	}

	// 4. Call service
	response, err := h.socialService.GetUserActivity(
		r.Context(),
		requesterID,
		targetUserID,
		perTypeLimit,
	)
	if err != nil {
		h.handleGetUserActivityError(w, err)

		return
	}

	SuccessResponse(w, http.StatusOK, response)
}

// extractOptionalUserID extracts user ID from header (nil if not provided or invalid).
func (h *SocialHandler) extractOptionalUserID(r *http.Request) *uuid.UUID {
	requesterIDStr := r.Header.Get("X-User-Id")
	if requesterIDStr == "" {
		return nil
	}

	requesterID, err := uuid.Parse(requesterIDStr)
	if err != nil {
		return nil
	}

	return &requesterID
}

func (h *SocialHandler) parseActivityParams(r *http.Request) (int, error) {
	perTypeLimit := defaultPerTypeLimit

	if limitStr := r.URL.Query().Get("per_type_limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, ErrInvalidPerTypeLimit
		}

		if limit < minPerTypeLimit || limit > maxPerTypeLimit {
			return 0, ErrPerTypeLimitOutOfRange
		}

		perTypeLimit = limit
	}

	return perTypeLimit, nil
}

func (h *SocialHandler) handleGetUserActivityError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrAccessDenied):
		ForbiddenResponse(w, "Access to this user's activity is restricted")
	default:
		InternalErrorResponse(w)
	}
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

func (h *SocialHandler) handleUnfollowUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCannotUnfollowSelf):
		ErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "Cannot unfollow yourself")
	case errors.Is(err, service.ErrUserNotFound):
		ErrorResponse(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
	default:
		InternalErrorResponse(w)
	}
}
