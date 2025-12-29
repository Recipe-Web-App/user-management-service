package oauth2_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/oauth2"
)

//nolint:funlen // table-driven test
func TestOAuth2Client_GetClientCredentialsToken(t *testing.T) {
	t.Parallel()

	t.Run("successfully obtains token", func(t *testing.T) {
		t.Parallel()

		expectedToken := oauth2.TokenResponse{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "read write",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/oauth2/token", r.URL.Path)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			// Verify Basic Auth
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "test-client-id", user)
			assert.Equal(t, "test-client-secret", pass)

			// Verify form data
			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
			assert.Equal(t, "read write", r.FormValue("scope"))

			w.Header().Set("Content-Type", "application/json")
			assert.NoError(t, json.NewEncoder(w).Encode(expectedToken))
		}))
		defer server.Close()

		cfg := &config.OAuth2Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			BaseAuthURL:  server.URL,
			GetTokenPath: "/oauth2/token",
		}

		client := oauth2.NewOAuth2ClientWithHTTP(cfg, server.Client())

		token, err := client.GetClientCredentialsToken(context.Background(), []string{"read", "write"})
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, expectedToken.AccessToken, token.AccessToken)
		assert.Equal(t, expectedToken.TokenType, token.TokenType)
		assert.Equal(t, expectedToken.ExpiresIn, token.ExpiresIn)
	})

	t.Run("returns error for missing credentials", func(t *testing.T) {
		t.Parallel()

		cfg := &config.OAuth2Config{
			ClientID:     "",
			ClientSecret: "",
			BaseAuthURL:  "http://localhost:8080",
			GetTokenPath: "/oauth2/token",
		}

		client := oauth2.NewOAuth2Client(cfg)

		token, err := client.GetClientCredentialsToken(context.Background(), nil)
		require.ErrorIs(t, err, oauth2.ErrMissingCredentials)
		assert.Nil(t, token)
	})

	t.Run("handles error response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(oauth2.OAuth2Error{
				Error:            "invalid_client",
				ErrorDescription: "Client authentication failed",
			})
		}))
		defer server.Close()

		cfg := &config.OAuth2Config{
			ClientID:     "wrong-client",
			ClientSecret: "wrong-secret",
			BaseAuthURL:  server.URL,
			GetTokenPath: "/oauth2/token",
		}

		client := oauth2.NewOAuth2ClientWithHTTP(cfg, server.Client())

		token, err := client.GetClientCredentialsToken(context.Background(), nil)
		require.Error(t, err)
		require.ErrorIs(t, err, oauth2.ErrMissingCredentials) // invalid_client maps to this
		assert.Nil(t, token)
	})

	t.Run("handles network error", func(t *testing.T) {
		t.Parallel()

		cfg := &config.OAuth2Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			BaseAuthURL:  "http://localhost:99999", // Invalid port
			GetTokenPath: "/oauth2/token",
		}

		client := oauth2.NewOAuth2Client(cfg)

		token, err := client.GetClientCredentialsToken(context.Background(), nil)
		require.Error(t, err)
		require.ErrorIs(t, err, oauth2.ErrTokenFetchFailed)
		assert.Nil(t, token)
	})
}

//nolint:funlen // table-driven test
func TestOAuth2Client_IntrospectToken(t *testing.T) {
	t.Parallel()

	t.Run("successfully introspects active token", func(t *testing.T) {
		t.Parallel()

		expectedResp := oauth2.IntrospectResponse{
			Active:   true,
			Sub:      "user-123",
			UserID:   "user-123",
			ClientID: "client-456",
			Scope:    "read write",
			Exp:      1735689600,
			Iat:      1735686000,
			Type:     "access_token",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/oauth2/introspect", r.URL.Path)

			// Verify Basic Auth
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "test-client-id", user)
			assert.Equal(t, "test-client-secret", pass)

			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "token-to-introspect", r.FormValue("token"))
			assert.Equal(t, "access_token", r.FormValue("token_type_hint"))

			w.Header().Set("Content-Type", "application/json")
			assert.NoError(t, json.NewEncoder(w).Encode(expectedResp))
		}))
		defer server.Close()

		cfg := &config.OAuth2Config{
			ClientID:          "test-client-id",
			ClientSecret:      "test-client-secret",
			BaseAuthURL:       server.URL,
			IntrospectionPath: "/oauth2/introspect",
		}

		client := oauth2.NewOAuth2ClientWithHTTP(cfg, server.Client())

		resp, err := client.IntrospectToken(context.Background(), "token-to-introspect")
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Active)
		assert.Equal(t, expectedResp.UserID, resp.UserID)
		assert.Equal(t, expectedResp.ClientID, resp.ClientID)
		assert.Equal(t, expectedResp.Scope, resp.Scope)
	})

	t.Run("returns error for inactive token", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(oauth2.IntrospectResponse{
				Active: false,
			})
		}))
		defer server.Close()

		cfg := &config.OAuth2Config{
			ClientID:          "test-client-id",
			ClientSecret:      "test-client-secret",
			BaseAuthURL:       server.URL,
			IntrospectionPath: "/oauth2/introspect",
		}

		client := oauth2.NewOAuth2ClientWithHTTP(cfg, server.Client())

		resp, err := client.IntrospectToken(context.Background(), "expired-token")
		require.ErrorIs(t, err, oauth2.ErrTokenInactive)
		assert.NotNil(t, resp) // Response is still returned for inspection
		assert.False(t, resp.Active)
	})

	t.Run("returns error for missing token", func(t *testing.T) {
		t.Parallel()

		cfg := &config.OAuth2Config{
			ClientID:          "test-client-id",
			ClientSecret:      "test-client-secret",
			BaseAuthURL:       "http://localhost:8080",
			IntrospectionPath: "/oauth2/introspect",
		}

		client := oauth2.NewOAuth2Client(cfg)

		resp, err := client.IntrospectToken(context.Background(), "")
		require.ErrorIs(t, err, oauth2.ErrMissingToken)
		assert.Nil(t, resp)
	})

	t.Run("returns error for missing credentials", func(t *testing.T) {
		t.Parallel()

		cfg := &config.OAuth2Config{
			ClientID:          "",
			ClientSecret:      "",
			IntrospectionPath: "/oauth2/introspect",
		}

		client := oauth2.NewOAuth2Client(cfg)

		resp, err := client.IntrospectToken(context.Background(), "some-token")
		require.ErrorIs(t, err, oauth2.ErrMissingCredentials)
		assert.Nil(t, resp)
	})
}

//nolint:funlen // table-driven test
func TestOAuth2Client_RevokeToken(t *testing.T) {
	t.Parallel()

	t.Run("successfully revokes token", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/oauth2/revoke", r.URL.Path)

			// Verify Basic Auth
			user, pass, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "test-client-id", user)
			assert.Equal(t, "test-client-secret", pass)

			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "token-to-revoke", r.FormValue("token"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := &config.OAuth2Config{
			ClientID:        "test-client-id",
			ClientSecret:    "test-client-secret",
			BaseAuthURL:     server.URL,
			RevokeTokenPath: "/oauth2/revoke",
		}

		client := oauth2.NewOAuth2ClientWithHTTP(cfg, server.Client())

		err := client.RevokeToken(context.Background(), "token-to-revoke")
		require.NoError(t, err)
	})

	t.Run("returns error for missing token", func(t *testing.T) {
		t.Parallel()

		cfg := &config.OAuth2Config{
			ClientID:        "test-client-id",
			ClientSecret:    "test-client-secret",
			BaseAuthURL:     "http://localhost:8080",
			RevokeTokenPath: "/oauth2/revoke",
		}

		client := oauth2.NewOAuth2Client(cfg)

		err := client.RevokeToken(context.Background(), "")
		require.ErrorIs(t, err, oauth2.ErrMissingToken)
	})

	t.Run("returns error for missing credentials", func(t *testing.T) {
		t.Parallel()

		cfg := &config.OAuth2Config{
			ClientID:        "",
			ClientSecret:    "",
			RevokeTokenPath: "/oauth2/revoke",
		}

		client := oauth2.NewOAuth2Client(cfg)

		err := client.RevokeToken(context.Background(), "some-token")
		require.ErrorIs(t, err, oauth2.ErrMissingCredentials)
	})
}

func TestIntrospectResponse_GetScopes(t *testing.T) {
	t.Parallel()

	t.Run("parses space-delimited scopes", func(t *testing.T) {
		t.Parallel()

		resp := &oauth2.IntrospectResponse{
			Scope: "read write admin",
		}

		scopes := resp.GetScopes()
		assert.Equal(t, []string{"read", "write", "admin"}, scopes)
	})

	t.Run("returns nil for empty scope", func(t *testing.T) {
		t.Parallel()

		resp := &oauth2.IntrospectResponse{
			Scope: "",
		}

		scopes := resp.GetScopes()
		assert.Nil(t, scopes)
	})

	t.Run("handles single scope", func(t *testing.T) {
		t.Parallel()

		resp := &oauth2.IntrospectResponse{
			Scope: "read",
		}

		scopes := resp.GetScopes()
		assert.Equal(t, []string{"read"}, scopes)
	})
}
