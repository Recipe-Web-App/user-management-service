package dto

// ============================================================================
// User Management Requests
// ============================================================================

// UserProfileUpdateRequest represents a request to update user profile.
type UserProfileUpdateRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50,username_pattern"`
	Email    *string `json:"email,omitempty"    validate:"omitempty,email"`
	FullName *string `json:"fullName,omitempty" validate:"omitempty,max=255"`
	Bio      *string `json:"bio,omitempty"      validate:"omitempty,max=1000"`
	IsActive *bool   `json:"-"` // Internal use only, not exposed in API
}

// UserAccountDeleteRequest represents a request to confirm account deletion.
type UserAccountDeleteRequest struct {
	ConfirmationToken string `json:"confirmationToken" validate:"required,min=1"`
}

// ============================================================================
// Notification Requests
// ============================================================================

// NotificationDeleteRequest represents a request to delete notifications.
type NotificationDeleteRequest struct {
	NotificationIDs []string `json:"notificationIds" validate:"required,min=1,max=100,dive,uuid"`
}

// UpdateNotificationPreferencesRequest represents a request to update notification preferences.
type UpdateNotificationPreferencesRequest struct {
	EmailNotifications   *bool `json:"emailNotifications,omitempty"`
	PushNotifications    *bool `json:"pushNotifications,omitempty"`
	FollowNotifications  *bool `json:"followNotifications,omitempty"`
	LikeNotifications    *bool `json:"likeNotifications,omitempty"`
	CommentNotifications *bool `json:"commentNotifications,omitempty"`
	RecipeNotifications  *bool `json:"recipeNotifications,omitempty"`
	SystemNotifications  *bool `json:"systemNotifications,omitempty"`
}

// UpdatePrivacyPreferencesRequest represents a request to update privacy preferences.
type UpdatePrivacyPreferencesRequest struct {
	ProfileVisibility *string `json:"profileVisibility,omitempty"`
	ShowEmail         *bool   `json:"showEmail,omitempty"`
	ShowFullName      *bool   `json:"showFullName,omitempty"`
	AllowFollows      *bool   `json:"allowFollows,omitempty"`
	AllowMessages     *bool   `json:"allowMessages,omitempty"`
}

// UpdateDisplayPreferencesRequest represents a request to update display preferences.
type UpdateDisplayPreferencesRequest struct {
	Theme    *string `json:"theme,omitempty"`
	Language *string `json:"language,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
}

// UpdateUserPreferenceRequest represents a request to update user preferences.
type UpdateUserPreferenceRequest struct {
	NotificationPreferences *UpdateNotificationPreferencesRequest `json:"notificationPreferences,omitempty"`
	PrivacyPreferences      *UpdatePrivacyPreferencesRequest      `json:"privacyPreferences,omitempty"`
	DisplayPreferences      *UpdateDisplayPreferencesRequest      `json:"displayPreferences,omitempty"`
}

// ============================================================================
// Metrics Requests
// ============================================================================

// CacheClearRequest represents a request to clear cache.
type CacheClearRequest struct {
	KeyPattern string `json:"keyPattern,omitempty"`
}
