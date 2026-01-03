package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/oauth2"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	xUserIDHeader       = "X-User-Id"
)

// AuthConfig holds the configuration for the Auth middleware.
type AuthConfig struct {
	// OAuth2Enabled indicates whether OAuth2 validation is enabled.
	// If false, the middleware extracts user ID from X-User-Id header.
	OAuth2Enabled bool

	// IntrospectionEnabled indicates whether to use remote token introspection.
	// Only used when OAuth2Enabled is true.
	// If false, local JWT validation is used.
	IntrospectionEnabled bool

	// JWTSecret is the secret used for local JWT validation.
	// Required when OAuth2Enabled is true and IntrospectionEnabled is false.
	JWTSecret string

	// OAuth2Client is the client used for token introspection.
	// Required when OAuth2Enabled is true and IntrospectionEnabled is true.
	OAuth2Client oauth2.Client
}

// Auth creates the authentication middleware with the specified configuration.
// This middleware extracts and validates authentication tokens and sets the
// authenticated user in the request context.
//
// Validation modes:
//   - OAuth2Enabled=false: Extract user ID from X-User-Id header
//   - OAuth2Enabled=true, IntrospectionEnabled=false: Validate JWT locally
//   - OAuth2Enabled=true, IntrospectionEnabled=true: Validate via introspection endpoint
func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				authUser *AuthenticatedUser
				err      error
			)

			switch {
			case !cfg.OAuth2Enabled:
				// Header-based authentication (for local development/testing)
				authUser, err = extractFromHeader(r)
			case cfg.IntrospectionEnabled:
				// Remote token introspection
				authUser, err = validateWithIntrospection(r, cfg.OAuth2Client)
			default:
				// Local JWT validation
				authUser, err = validateJWT(r, cfg.JWTSecret)
			}

			if err != nil {
				slog.Debug("authentication failed",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
				)
				unauthorizedResponse(w, "Authentication required")

				return
			}

			// Set the authenticated user in the request context
			ctx := SetAuthenticatedUser(r.Context(), authUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractFromHeader extracts the user ID from the X-User-Id header.
// This mode is used when OAuth2 is disabled (local development/testing).
func extractFromHeader(r *http.Request) (*AuthenticatedUser, error) {
	userIDStr := r.Header.Get(xUserIDHeader)
	if userIDStr == "" {
		return nil, oauth2.ErrMissingToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, oauth2.ErrInvalidToken
	}

	return &AuthenticatedUser{
		UserID:    userID,
		ClientID:  "local",
		Scopes:    nil,
		IsService: false,
	}, nil
}

// validateJWT validates the Bearer token using local JWT validation.
func validateJWT(r *http.Request, jwtSecret string) (*AuthenticatedUser, error) {
	tokenString, err := extractBearerToken(r)
	if err != nil {
		return nil, err
	}

	claims, err := oauth2.ValidateAccessToken(tokenString, jwtSecret)
	if err != nil {
		return nil, err //nolint:wrapcheck // oauth2 errors are already wrapped
	}

	return buildAuthUserFromJWT(claims)
}

// validateWithIntrospection validates the Bearer token via the introspection endpoint.
func validateWithIntrospection(r *http.Request, client oauth2.Client) (*AuthenticatedUser, error) {
	tokenString, err := extractBearerToken(r)
	if err != nil {
		return nil, err
	}

	resp, err := client.IntrospectToken(r.Context(), tokenString)
	if err != nil {
		return nil, err //nolint:wrapcheck // oauth2 errors are already wrapped
	}

	return buildAuthUserFromIntrospection(resp)
}

// extractBearerToken extracts the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get(authorizationHeader)
	if authHeader == "" {
		return "", oauth2.ErrMissingToken
	}

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", oauth2.ErrInvalidTokenFormat
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", oauth2.ErrMissingToken
	}

	return token, nil
}

// buildAuthUserFromJWT creates an AuthenticatedUser from JWT claims.
func buildAuthUserFromJWT(claims *oauth2.JWTClaims) (*AuthenticatedUser, error) {
	userID, err := claims.GetUserUUID()
	if err != nil {
		// Check if this is a service token (no user ID)
		if claims.ClientID != "" {
			return &AuthenticatedUser{
				UserID:    uuid.Nil,
				ClientID:  claims.ClientID,
				Scopes:    claims.Scopes,
				IsService: true,
			}, nil
		}

		return nil, err //nolint:wrapcheck // oauth2 errors are already wrapped
	}

	return &AuthenticatedUser{
		UserID:    userID,
		ClientID:  claims.ClientID,
		Scopes:    claims.Scopes,
		IsService: false,
	}, nil
}

// buildAuthUserFromIntrospection creates an AuthenticatedUser from an introspection response.
func buildAuthUserFromIntrospection(resp *oauth2.IntrospectResponse) (*AuthenticatedUser, error) {
	userID, err := resp.GetUserID()
	if err != nil {
		// Check if this is a service token (no user ID)
		if resp.ClientID != "" {
			return &AuthenticatedUser{
				UserID:    uuid.Nil,
				ClientID:  resp.ClientID,
				Scopes:    resp.GetScopes(),
				IsService: true,
			}, nil
		}

		return nil, err //nolint:wrapcheck // oauth2 errors are already wrapped
	}

	return &AuthenticatedUser{
		UserID:    userID,
		ClientID:  resp.ClientID,
		Scopes:    resp.GetScopes(),
		IsService: false,
	}, nil
}

// unauthorizedResponse sends a 401 Unauthorized response.
func unauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"UNAUTHORIZED","message":"` + message + `"}`))
}
