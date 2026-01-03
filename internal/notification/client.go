package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// TokenProvider abstracts token retrieval for service-to-service auth.
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
}

// DownstreamClient defines a generic interface for HTTP clients communicating with downstream services.
type DownstreamClient interface {
	// Do executes an HTTP request with the given method, path, and body.
	// It handles authentication, encoding, and error responses.
	Do(ctx context.Context, method, path string, body any, response any) error
}

// BaseClient implements DownstreamClient with common functionality.
type BaseClient struct {
	httpClient    *http.Client
	baseURL       string
	tokenProvider TokenProvider
	logger        *slog.Logger
}

// NewBaseClient creates a new BaseClient with the given configuration.
func NewBaseClient(baseURL string, tokenProvider TokenProvider, timeout time.Duration) *BaseClient {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &BaseClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		tokenProvider: tokenProvider,
		logger:        slog.Default(),
	}
}

// NewBaseClientWithHTTP creates a BaseClient with a custom HTTP client (for testing).
func NewBaseClientWithHTTP(baseURL string, tokenProvider TokenProvider, httpClient *http.Client) *BaseClient {
	return &BaseClient{
		httpClient:    httpClient,
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		tokenProvider: tokenProvider,
		logger:        slog.Default(),
	}
}

// Do executes an HTTP request with authentication.
func (c *BaseClient) Do(ctx context.Context, method, path string, body any, response any) error {
	// 1. Build URL
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	// 2. Marshal body
	var bodyReader io.Reader

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 3. Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 4. Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 5. Add authorization token
	if c.tokenProvider != nil {
		token, tokenErr := c.tokenProvider.GetToken(ctx)
		if tokenErr != nil {
			return fmt.Errorf("failed to get auth token: %w", tokenErr)
		}

		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 6. Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNotificationFailed, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// 7. Handle response
	return c.handleResponse(resp, response)
}

func (c *BaseClient) handleResponse(resp *http.Response, response any) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		if response != nil {
			err := json.NewDecoder(resp.Body).Decode(response)
			if err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}

		return nil
	case http.StatusBadRequest:
		return ErrNotificationBadRequest
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrNotificationForbidden
	case http.StatusTooManyRequests:
		return ErrNotificationRateLimit
	default:
		return fmt.Errorf("%w: status %d", ErrNotificationFailed, resp.StatusCode)
	}
}
