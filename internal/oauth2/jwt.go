package oauth2

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DefaultTokenExpiry is the default expiry time for test tokens.
const DefaultTokenExpiry = 15 * time.Minute

// ValidateAccessToken validates a JWT access token using the provided secret.
// It returns the parsed claims if the token is valid, or an error if validation fails.
func ValidateAccessToken(tokenString, secret string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	if secret == "" {
		return nil, ErrMissingCredentials
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Validate the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSignMethod, token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		// Check for specific JWT errors
		if isTokenExpiredError(err) {
			return nil, ErrTokenExpired
		}

		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate token type is access_token
	if claims.Type != "" && claims.Type != "access_token" {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// isTokenExpiredError checks if the error is due to token expiration.
func isTokenExpiredError(err error) bool {
	// jwt/v5 returns wrapped errors, so we need to check the error type
	return jwt.ErrTokenExpired != nil && err != nil &&
		(errors.Is(err, jwt.ErrTokenExpired) || err.Error() == jwt.ErrTokenExpired.Error() ||
			containsString(err.Error(), "token is expired"))
}

// containsString is a simple substring check.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// CreateTestToken creates a JWT token for testing purposes.
// This should only be used in tests.
func CreateTestToken(claims *JWTClaims, secret string) (string, error) {
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(DefaultTokenExpiry))
	}

	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
	}

	if claims.Type == "" {
		claims.Type = "access_token"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signedToken, nil
}
