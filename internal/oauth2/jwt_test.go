package oauth2_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/oauth2"
)

const testSecret = "test-secret-key-minimum-32-chars!"

//nolint:funlen // table-driven test
func TestValidateAccessToken(t *testing.T) {
	t.Parallel()

	t.Run("validates valid token successfully", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New().String()
		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "auth-service",
			},
			UserID:   userID,
			ClientID: "test-client",
			Scopes:   []string{"read", "write"},
			Type:     "access_token",
		}

		tokenString, err := oauth2.CreateTestToken(claims, testSecret)
		require.NoError(t, err)

		result, err := oauth2.ValidateAccessToken(tokenString, testSecret)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, "test-client", result.ClientID)
		assert.Equal(t, []string{"read", "write"}, result.Scopes)
		assert.Equal(t, "access_token", result.Type)
	})

	t.Run("returns error for expired token", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
			Type: "access_token",
		}

		tokenString, err := oauth2.CreateTestToken(claims, testSecret)
		require.NoError(t, err)

		result, err := oauth2.ValidateAccessToken(tokenString, testSecret)
		require.ErrorIs(t, err, oauth2.ErrTokenExpired)
		assert.Nil(t, result)
	})

	t.Run("returns error for invalid signature", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			},
			Type: "access_token",
		}

		// Create token with one secret
		tokenString, err := oauth2.CreateTestToken(claims, testSecret)
		require.NoError(t, err)

		// Validate with different secret
		result, err := oauth2.ValidateAccessToken(tokenString, "different-secret-key-32-chars!!")
		require.Error(t, err)
		require.ErrorIs(t, err, oauth2.ErrInvalidToken)
		assert.Nil(t, result)
	})

	t.Run("returns error for empty token", func(t *testing.T) {
		t.Parallel()

		result, err := oauth2.ValidateAccessToken("", testSecret)
		require.ErrorIs(t, err, oauth2.ErrMissingToken)
		assert.Nil(t, result)
	})

	t.Run("returns error for empty secret", func(t *testing.T) {
		t.Parallel()

		result, err := oauth2.ValidateAccessToken("some.token.here", "")
		require.ErrorIs(t, err, oauth2.ErrMissingCredentials)
		assert.Nil(t, result)
	})

	t.Run("returns error for malformed token", func(t *testing.T) {
		t.Parallel()

		result, err := oauth2.ValidateAccessToken("not-a-valid-jwt", testSecret)
		require.Error(t, err)
		require.ErrorIs(t, err, oauth2.ErrInvalidToken)
		assert.Nil(t, result)
	})

	t.Run("returns error for wrong token type", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			},
			Type: "refresh_token", // Wrong type
		}

		tokenString, err := oauth2.CreateTestToken(claims, testSecret)
		require.NoError(t, err)

		result, err := oauth2.ValidateAccessToken(tokenString, testSecret)
		require.ErrorIs(t, err, oauth2.ErrInvalidTokenType)
		assert.Nil(t, result)
	})

	t.Run("accepts token without type claim (backwards compatibility)", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			},
			UserID: uuid.New().String(),
			Type:   "", // No type
		}

		// Manually create token to avoid CreateTestToken setting default type
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(testSecret))
		require.NoError(t, err)

		result, err := oauth2.ValidateAccessToken(tokenString, testSecret)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestJWTClaims_GetUserUUID(t *testing.T) {
	t.Parallel()

	t.Run("returns user_id when present", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New()
		claims := &oauth2.JWTClaims{
			UserID: userID.String(),
		}

		result, err := claims.GetUserUUID()
		require.NoError(t, err)
		assert.Equal(t, userID, result)
	})

	t.Run("falls back to subject when user_id is empty", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New()
		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
			UserID: "",
		}

		result, err := claims.GetUserUUID()
		require.NoError(t, err)
		assert.Equal(t, userID, result)
	})

	t.Run("returns error when both are empty", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			UserID: "",
		}

		result, err := claims.GetUserUUID()
		require.ErrorIs(t, err, oauth2.ErrNoUserID)
		assert.Equal(t, uuid.Nil, result)
	})

	t.Run("returns error for invalid UUID in user_id", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			UserID: "not-a-uuid",
		}

		result, err := claims.GetUserUUID()
		require.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
	})
}

func TestCreateTestToken(t *testing.T) {
	t.Parallel()

	t.Run("creates valid token with defaults", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			UserID:   uuid.New().String(),
			ClientID: "test-client",
		}

		tokenString, err := oauth2.CreateTestToken(claims, testSecret)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		// Verify token can be parsed back
		result, err := oauth2.ValidateAccessToken(tokenString, testSecret)
		require.NoError(t, err)
		assert.Equal(t, claims.UserID, result.UserID)
		assert.Equal(t, "access_token", result.Type) // Default type
	})

	t.Run("respects provided expiry", func(t *testing.T) {
		t.Parallel()

		customExpiry := time.Now().Add(1 * time.Hour)
		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(customExpiry),
			},
			UserID: uuid.New().String(),
		}

		tokenString, err := oauth2.CreateTestToken(claims, testSecret)
		require.NoError(t, err)

		result, err := oauth2.ValidateAccessToken(tokenString, testSecret)
		require.NoError(t, err)
		assert.Equal(t, customExpiry.Unix(), result.ExpiresAt.Unix())
	})
}
