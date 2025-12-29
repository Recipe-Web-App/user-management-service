package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/middleware"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/oauth2"
)

const testJWTSecret = "test-secret-key-minimum-32-chars!"

// MockOAuth2Client is a mock implementation of oauth2.Client.
type MockOAuth2Client struct {
	mock.Mock
}

//nolint:wrapcheck // test mock returns test errors directly
func (m *MockOAuth2Client) GetClientCredentialsToken(
	ctx context.Context,
	scopes []string,
) (*oauth2.TokenResponse, error) {
	args := m.Called(ctx, scopes)

	resp, ok := args.Get(0).(*oauth2.TokenResponse)
	if !ok {
		return nil, args.Error(1)
	}

	return resp, args.Error(1)
}

//nolint:wrapcheck // test mock returns test errors directly
func (m *MockOAuth2Client) IntrospectToken(ctx context.Context, token string) (*oauth2.IntrospectResponse, error) {
	args := m.Called(ctx, token)

	resp, ok := args.Get(0).(*oauth2.IntrospectResponse)
	if !ok {
		return nil, args.Error(1)
	}

	return resp, args.Error(1)
}

//nolint:wrapcheck // test mock returns test errors directly
func (m *MockOAuth2Client) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)

	return args.Error(0)
}

//nolint:funlen // table-driven test
func TestAuth_HeaderMode(t *testing.T) {
	t.Parallel()

	cfg := middleware.AuthConfig{
		OAuth2Enabled: false, // Header mode
	}

	t.Run("extracts user from X-User-Id header", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New()

		var capturedUser *middleware.AuthenticatedUser

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			var ok bool

			capturedUser, ok = middleware.GetAuthenticatedUser(r.Context())
			assert.True(t, ok)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-Id", userID.String())

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, capturedUser)
		assert.Equal(t, userID, capturedUser.UserID)
		assert.Equal(t, "local", capturedUser.ClientID)
		assert.False(t, capturedUser.IsService)
	})

	t.Run("returns 401 when X-User-Id header is missing", func(t *testing.T) {
		t.Parallel()

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
	})

	t.Run("returns 401 when X-User-Id header is invalid UUID", func(t *testing.T) {
		t.Parallel()

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-Id", "not-a-uuid")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

//nolint:funlen // table-driven test
func TestAuth_JWTMode(t *testing.T) {
	t.Parallel()

	cfg := middleware.AuthConfig{
		OAuth2Enabled:        true,
		IntrospectionEnabled: false,
		JWTSecret:            testJWTSecret,
	}

	t.Run("validates valid JWT token", func(t *testing.T) {
		t.Parallel()

		userID := uuid.New()
		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID.String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID:   userID.String(),
			ClientID: "test-client",
			Scopes:   []string{"read", "write"},
			Type:     "access_token",
		}

		tokenString, err := oauth2.CreateTestToken(claims, testJWTSecret)
		require.NoError(t, err)

		var capturedUser *middleware.AuthenticatedUser

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			var ok bool

			capturedUser, ok = middleware.GetAuthenticatedUser(r.Context())
			assert.True(t, ok)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, capturedUser)
		assert.Equal(t, userID, capturedUser.UserID)
		assert.Equal(t, "test-client", capturedUser.ClientID)
		assert.Equal(t, []string{"read", "write"}, capturedUser.Scopes)
		assert.False(t, capturedUser.IsService)
	})

	t.Run("handles service token (no user ID)", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			},
			ClientID: "service-client",
			Scopes:   []string{"service"},
			Type:     "access_token",
		}

		tokenString, err := oauth2.CreateTestToken(claims, testJWTSecret)
		require.NoError(t, err)

		var capturedUser *middleware.AuthenticatedUser

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			var ok bool

			capturedUser, ok = middleware.GetAuthenticatedUser(r.Context())
			assert.True(t, ok)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, capturedUser)
		assert.Equal(t, uuid.Nil, capturedUser.UserID)
		assert.Equal(t, "service-client", capturedUser.ClientID)
		assert.True(t, capturedUser.IsService)
	})

	t.Run("returns 401 for missing Authorization header", func(t *testing.T) {
		t.Parallel()

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "Bearer", rr.Header().Get("WWW-Authenticate"))
	})

	t.Run("returns 401 for invalid token format", func(t *testing.T) {
		t.Parallel()

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // Basic auth instead of Bearer

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 401 for expired token", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			},
			Type: "access_token",
		}

		tokenString, err := oauth2.CreateTestToken(claims, testJWTSecret)
		require.NoError(t, err)

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 401 for invalid signature", func(t *testing.T) {
		t.Parallel()

		claims := &oauth2.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   uuid.New().String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			},
			Type: "access_token",
		}

		// Sign with different secret
		tokenString, err := oauth2.CreateTestToken(claims, "different-secret-32-characters!!")
		require.NoError(t, err)

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

//nolint:funlen // table-driven test
func TestAuth_IntrospectionMode(t *testing.T) {
	t.Parallel()

	t.Run("validates token via introspection", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockOAuth2Client)
		userID := uuid.New()

		mockClient.On("IntrospectToken", mock.Anything, "valid-token").
			Return(&oauth2.IntrospectResponse{
				Active:   true,
				Sub:      userID.String(),
				UserID:   userID.String(),
				ClientID: "introspect-client",
				Scope:    "read write",
			}, nil)

		cfg := middleware.AuthConfig{
			OAuth2Enabled:        true,
			IntrospectionEnabled: true,
			OAuth2Client:         mockClient,
		}

		var capturedUser *middleware.AuthenticatedUser

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			var ok bool

			capturedUser, ok = middleware.GetAuthenticatedUser(r.Context())
			assert.True(t, ok)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, capturedUser)
		assert.Equal(t, userID, capturedUser.UserID)
		assert.Equal(t, "introspect-client", capturedUser.ClientID)
		assert.Equal(t, []string{"read", "write"}, capturedUser.Scopes)

		mockClient.AssertExpectations(t)
	})

	t.Run("returns 401 for inactive token", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockOAuth2Client)

		mockClient.On("IntrospectToken", mock.Anything, "inactive-token").
			Return(&oauth2.IntrospectResponse{
				Active: false,
			}, oauth2.ErrTokenInactive)

		cfg := middleware.AuthConfig{
			OAuth2Enabled:        true,
			IntrospectionEnabled: true,
			OAuth2Client:         mockClient,
		}

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer inactive-token")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockClient.AssertExpectations(t)
	})

	t.Run("returns 401 when introspection fails", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockOAuth2Client)

		mockClient.On("IntrospectToken", mock.Anything, "error-token").
			Return(nil, oauth2.ErrIntrospectionFailed)

		cfg := middleware.AuthConfig{
			OAuth2Enabled:        true,
			IntrospectionEnabled: true,
			OAuth2Client:         mockClient,
		}

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer error-token")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockClient.AssertExpectations(t)
	})

	t.Run("handles service token via introspection", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockOAuth2Client)

		mockClient.On("IntrospectToken", mock.Anything, "service-token").
			Return(&oauth2.IntrospectResponse{
				Active:   true,
				ClientID: "service-client",
				Scope:    "service admin",
				// No user_id or sub - this is a service token
			}, nil)

		cfg := middleware.AuthConfig{
			OAuth2Enabled:        true,
			IntrospectionEnabled: true,
			OAuth2Client:         mockClient,
		}

		var capturedUser *middleware.AuthenticatedUser

		handler := middleware.Auth(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			var ok bool

			capturedUser, ok = middleware.GetAuthenticatedUser(r.Context())
			assert.True(t, ok)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer service-token")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		require.NotNil(t, capturedUser)
		assert.Equal(t, uuid.Nil, capturedUser.UserID)
		assert.Equal(t, "service-client", capturedUser.ClientID)
		assert.True(t, capturedUser.IsService)
		assert.Equal(t, []string{"service", "admin"}, capturedUser.Scopes)

		mockClient.AssertExpectations(t)
	})
}
