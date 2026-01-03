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
// Metrics Requests
// ============================================================================

// CacheClearRequest represents a request to clear cache.
type CacheClearRequest struct {
	KeyPattern string `json:"keyPattern,omitempty"`
}
