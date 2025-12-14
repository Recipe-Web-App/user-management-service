//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// SocialHandler handles social feature HTTP endpoints.
type SocialHandler struct{}

// NewSocialHandler creates a new social handler.
func NewSocialHandler() *SocialHandler {
	return &SocialHandler{}
}

// GetFollowing handles GET /users/{user_id}/following.
func (h *SocialHandler) GetFollowing(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	limit := 20
	offset := 0
	fullName := "Jane Smith"

	SuccessResponse(w, http.StatusOK, dto.GetFollowedUsersResponse{
		TotalCount: 1,
		FollowedUsers: []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "janesmith",
				FullName:  &fullName,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Limit:  &limit,
		Offset: &offset,
	})
}

// GetFollowers handles GET /users/{user_id}/followers.
func (h *SocialHandler) GetFollowers(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	limit := 20
	offset := 0
	fullName := "Bob Wilson"

	SuccessResponse(w, http.StatusOK, dto.GetFollowedUsersResponse{
		TotalCount: 1,
		FollowedUsers: []dto.User{
			{
				UserID:    uuid.New().String(),
				Username:  "bobwilson",
				FullName:  &fullName,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Limit:  &limit,
		Offset: &offset,
	})
}

// FollowUser handles POST /users/{user_id}/follow/{target_user_id}.
func (h *SocialHandler) FollowUser(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.FollowResponse{
		Message:     "Successfully followed user",
		IsFollowing: true,
	})
}

// UnfollowUser handles DELETE /users/{user_id}/follow/{target_user_id}.
func (h *SocialHandler) UnfollowUser(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.FollowResponse{
		Message:     "Successfully unfollowed user",
		IsFollowing: false,
	})
}

// GetUserActivity handles GET /users/{user_id}/activity.
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
