package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
)

const (
	defaultTimeout      = 30 * time.Second
	contentTypeForm     = "application/x-www-form-urlencoded"
	grantClientCreds    = "client_credentials"
	tokenTypeHintAccess = "access_token"
)

// Client defines the interface for OAuth2 operations.
type Client interface {
	// GetClientCredentialsToken obtains a token using the client credentials flow.
	GetClientCredentialsToken(ctx context.Context, scopes []string) (*TokenResponse, error)

	// IntrospectToken validates a token via the introspection endpoint.
	IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error)

	// RevokeToken revokes an access or refresh token.
	RevokeToken(ctx context.Context, token string) error
}

// OAuth2Client implements the Client interface for communicating with the auth service.
type OAuth2Client struct {
	httpClient *http.Client
	config     *config.OAuth2Config
	logger     *slog.Logger
}

// NewOAuth2Client creates a new OAuth2 client with the provided configuration.
func NewOAuth2Client(cfg *config.OAuth2Config) *OAuth2Client {
	return &OAuth2Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		config: cfg,
		logger: slog.Default(),
	}
}

// NewOAuth2ClientWithHTTP creates a new OAuth2 client with a custom HTTP client.
// This is useful for testing with mock servers.
func NewOAuth2ClientWithHTTP(cfg *config.OAuth2Config, httpClient *http.Client) *OAuth2Client {
	return &OAuth2Client{
		httpClient: httpClient,
		config:     cfg,
		logger:     slog.Default(),
	}
}

// GetClientCredentialsToken obtains a token using the client credentials flow.
func (c *OAuth2Client) GetClientCredentialsToken(ctx context.Context, scopes []string) (*TokenResponse, error) {
	if c.config.ClientID == "" || c.config.ClientSecret == "" {
		return nil, ErrMissingCredentials
	}

	tokenURL := c.buildURL(c.config.GetTokenPath)

	data := url.Values{
		"grant_type": {grantClientCreds},
	}

	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatingRequest, err)
	}

	req.Header.Set("Content-Type", contentTypeForm)
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("token request failed", "error", err)

		return nil, fmt.Errorf("%w: %w", ErrTokenFetchFailed, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, "token request")
	}

	var tokenResp TokenResponse

	decodeErr := json.NewDecoder(resp.Body).Decode(&tokenResp)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodingResponse, decodeErr)
	}

	return &tokenResp, nil
}

// IntrospectToken validates a token via the introspection endpoint.
func (c *OAuth2Client) IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error) {
	if c.config.ClientID == "" || c.config.ClientSecret == "" {
		return nil, ErrMissingCredentials
	}

	if token == "" {
		return nil, ErrMissingToken
	}

	introspectURL := c.buildURL(c.config.IntrospectionPath)

	data := url.Values{
		"token":           {token},
		"token_type_hint": {tokenTypeHintAccess},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatingRequest, err)
	}

	req.Header.Set("Content-Type", contentTypeForm)
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("introspection request failed", "error", err)

		return nil, fmt.Errorf("%w: %w", ErrIntrospectionFailed, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, "introspection")
	}

	var introspectResp IntrospectResponse

	decodeErr := json.NewDecoder(resp.Body).Decode(&introspectResp)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodingResponse, decodeErr)
	}

	// If token is not active, return a specific error
	if !introspectResp.Active {
		return &introspectResp, ErrTokenInactive
	}

	return &introspectResp, nil
}

// RevokeToken revokes an access or refresh token.
func (c *OAuth2Client) RevokeToken(ctx context.Context, token string) error {
	if c.config.ClientID == "" || c.config.ClientSecret == "" {
		return ErrMissingCredentials
	}

	if token == "" {
		return ErrMissingToken
	}

	revokeURL := c.buildURL(c.config.RevokeTokenPath)

	data := url.Values{
		"token":           {token},
		"token_type_hint": {tokenTypeHintAccess},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreatingRequest, err)
	}

	req.Header.Set("Content-Type", contentTypeForm)
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("revocation request failed", "error", err)

		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// Per RFC 7009, revocation endpoint returns 200 even if token was already revoked
	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp, "revocation")
	}

	return nil
}

// buildURL constructs the full URL for an OAuth2 endpoint.
func (c *OAuth2Client) buildURL(path string) string {
	baseURL := strings.TrimSuffix(c.config.BaseAuthURL, "/")
	path = strings.TrimPrefix(path, "/")

	return fmt.Sprintf("%s/%s", baseURL, path)
}

// handleErrorResponse extracts error information from an OAuth2 error response.
func (c *OAuth2Client) handleErrorResponse(resp *http.Response, operation string) error {
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		c.logger.Warn("failed to read error response",
			"operation", operation,
			"status", resp.StatusCode,
			"error", readErr,
		)

		return fmt.Errorf("%w: status %d", ErrOAuth2ServerError, resp.StatusCode)
	}

	var oauth2Err OAuth2Error

	unmarshalErr := json.Unmarshal(body, &oauth2Err)
	if unmarshalErr != nil {
		// Not a standard OAuth2 error response
		c.logger.Warn("non-standard error response",
			"operation", operation,
			"status", resp.StatusCode,
			"body", string(body),
		)

		return fmt.Errorf("%w: status %d", ErrOAuth2ServerError, resp.StatusCode)
	}

	c.logger.Warn("OAuth2 error",
		"operation", operation,
		"error", oauth2Err.Error,
		"description", oauth2Err.ErrorDescription,
	)

	switch oauth2Err.Error {
	case ErrorInvalidClient:
		return ErrMissingCredentials
	case ErrorInvalidGrant:
		return ErrInvalidToken
	default:
		return fmt.Errorf("%w: %s", ErrOAuth2ServerError, oauth2Err.ErrorDescription)
	}
}
