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

// UpdateUserPreferenceRequest represents a request to update user preferences.
type UpdateUserPreferenceRequest struct {
	NotificationPreferences *NotificationPreferences `json:"notificationPreferences,omitempty"`
	PrivacyPreferences      *PrivacyPreferences      `json:"privacyPreferences,omitempty"`
	DisplayPreferences      *DisplayPreferences      `json:"displayPreferences,omitempty"`
}

// ============================================================================
// Metrics Requests
// ============================================================================

// CacheClearRequest represents a request to clear cache.
type CacheClearRequest struct {
	KeyPattern string `json:"keyPattern,omitempty"`
}
