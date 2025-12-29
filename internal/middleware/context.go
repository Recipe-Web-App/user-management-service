package middleware

import (
	"context"
	"slices"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// AuthUserKey is the context key for the authenticated user.
const AuthUserKey contextKey = "auth_user"

// AuthenticatedUser represents the authenticated user extracted from a token.
type AuthenticatedUser struct {
	// UserID is the unique identifier of the user.
	UserID uuid.UUID

	// ClientID is the OAuth2 client ID that issued/owns the token.
	ClientID string

	// Scopes are the permissions granted to this token.
	Scopes []string

	// IsService indicates if this is a service-to-service token (client credentials flow)
	// rather than a user token. When true, UserID may be nil.
	IsService bool
}

// GetAuthenticatedUser retrieves the authenticated user from the request context.
// Returns nil and false if no user is present in the context.
func GetAuthenticatedUser(ctx context.Context) (*AuthenticatedUser, bool) {
	user, ok := ctx.Value(AuthUserKey).(*AuthenticatedUser)
	if !ok || user == nil {
		return nil, false
	}

	return user, true
}

// SetAuthenticatedUser stores the authenticated user in the context.
// Returns a new context with the user value.
func SetAuthenticatedUser(ctx context.Context, user *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, AuthUserKey, user)
}

// GetUserIDFromContext is a convenience function that extracts just the user ID
// from the context. Returns uuid.Nil and false if no authenticated user is present
// or if the user is a service account (IsService=true).
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	user, ok := GetAuthenticatedUser(ctx)
	if !ok {
		return uuid.Nil, false
	}

	// Service tokens don't have a user ID
	if user.IsService {
		return uuid.Nil, false
	}

	// Check if user ID is set (not nil UUID)
	if user.UserID == uuid.Nil {
		return uuid.Nil, false
	}

	return user.UserID, true
}

// HasScope checks if the authenticated user has the specified scope.
// Returns false if no user is present or the scope is not found.
func HasScope(ctx context.Context, scope string) bool {
	user, ok := GetAuthenticatedUser(ctx)
	if !ok {
		return false
	}

	return slices.Contains(user.Scopes, scope)
}
