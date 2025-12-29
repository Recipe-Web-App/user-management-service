// Package oauth2 provides OAuth2 client functionality for authentication.
package oauth2

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenResponse represents the OAuth2 token endpoint response.
//
//nolint:tagliatelle // OAuth2 spec (RFC 6749) requires snake_case JSON field names
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// IntrospectResponse represents the OAuth2 introspection response (RFC 7662).
//
//nolint:tagliatelle // OAuth2 spec (RFC 7662) requires snake_case JSON field names
type IntrospectResponse struct {
	Active   bool   `json:"active"`
	Sub      string `json:"sub,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Scope    string `json:"scope,omitempty"`
	Exp      int64  `json:"exp,omitempty"`
	Iat      int64  `json:"iat,omitempty"`
	Type     string `json:"type,omitempty"`
}

// GetScopes parses the space-delimited scope string into a slice.
func (r *IntrospectResponse) GetScopes() []string {
	if r.Scope == "" {
		return nil
	}

	return strings.Split(r.Scope, " ")
}

// GetUserID parses the user_id claim as a UUID.
func (r *IntrospectResponse) GetUserID() (uuid.UUID, error) {
	if r.UserID == "" {
		if r.Sub != "" {
			id, err := uuid.Parse(r.Sub)
			if err != nil {
				return uuid.Nil, fmt.Errorf("parsing sub claim: %w", err)
			}

			return id, nil
		}

		return uuid.Nil, ErrNoUserID
	}

	id, err := uuid.Parse(r.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing user_id claim: %w", err)
	}

	return id, nil
}

// JWTClaims represents the claims in a JWT access token.
//
//nolint:tagliatelle // OAuth2/JWT spec requires snake_case JSON field names
type JWTClaims struct {
	jwt.RegisteredClaims

	UserID   string   `json:"user_id,omitempty"`
	ClientID string   `json:"client_id,omitempty"`
	Scopes   []string `json:"scopes,omitempty"`
	Type     string   `json:"type,omitempty"`
}

// GetUserUUID parses the user_id claim as a UUID.
func (c *JWTClaims) GetUserUUID() (uuid.UUID, error) {
	if c.UserID == "" {
		if c.Subject != "" {
			id, err := uuid.Parse(c.Subject)
			if err != nil {
				return uuid.Nil, fmt.Errorf("parsing subject claim: %w", err)
			}

			return id, nil
		}

		return uuid.Nil, ErrNoUserID
	}

	id, err := uuid.Parse(c.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing user_id claim: %w", err)
	}

	return id, nil
}

// OAuth2Error represents an error response from the OAuth2 server.
//
//nolint:tagliatelle // OAuth2 spec (RFC 6749) requires snake_case JSON field names
type OAuth2Error struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// OAuth2 error codes.
const (
	ErrorInvalidRequest       = "invalid_request"
	ErrorInvalidClient        = "invalid_client"
	ErrorInvalidGrant         = "invalid_grant"
	ErrorUnauthorizedClient   = "unauthorized_client"
	ErrorUnsupportedGrantType = "unsupported_grant_type"
	ErrorInvalidScope         = "invalid_scope"
	ErrorAccessDenied         = "access_denied"
	ErrorServerError          = "server_error"
)

// Sentinel errors for OAuth2 operations.
var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrTokenInactive        = errors.New("token is not active")
	ErrIntrospectionFailed  = errors.New("token introspection failed")
	ErrTokenFetchFailed     = errors.New("failed to fetch token")
	ErrMissingCredentials   = errors.New("missing client credentials")
	ErrNoUserID             = errors.New("no user ID in token claims")
	ErrInvalidTokenType     = errors.New("invalid token type")
	ErrMissingToken         = errors.New("missing authorization token")
	ErrInvalidTokenFormat   = errors.New("invalid token format")
	ErrCreatingRequest      = errors.New("failed to create HTTP request")
	ErrDecodingResponse     = errors.New("failed to decode response")
	ErrRequestFailed        = errors.New("request failed")
	ErrOAuth2ServerError    = errors.New("OAuth2 server error")
	ErrUnexpectedSignMethod = errors.New("unexpected signing method")
)
