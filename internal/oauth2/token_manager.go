package oauth2

import (
	"context"
	"sync"
	"time"
)

const (
	// DefaultRefreshBuffer is the time before expiry when tokens should be refreshed.
	// Using 10% of the token lifetime means we refresh when 90% of the TTL has elapsed.
	DefaultRefreshBuffer = 0.1
)

// TokenManager handles token lifecycle for service-to-service authentication.
type TokenManager interface {
	// GetToken returns a valid access token, refreshing if necessary.
	GetToken(ctx context.Context) (string, error)

	// InvalidateToken forces a token refresh on the next GetToken call.
	InvalidateToken()

	// Close cleans up any resources.
	Close()
}

// DefaultTokenManager implements TokenManager with caching and automatic refresh.
type DefaultTokenManager struct {
	client        Client
	scopes        []string
	token         *TokenResponse
	expiresAt     time.Time
	mu            sync.RWMutex
	refreshBuffer float64 // Percentage of TTL to use as buffer (0.1 = 10%)
	closed        bool
}

// NewTokenManager creates a new TokenManager with default settings.
func NewTokenManager(client Client, scopes []string) *DefaultTokenManager {
	return &DefaultTokenManager{
		client:        client,
		scopes:        scopes,
		refreshBuffer: DefaultRefreshBuffer,
	}
}

// NewTokenManagerWithBuffer creates a new TokenManager with a custom refresh buffer.
// The buffer is specified as a percentage of the token TTL (e.g., 0.1 = 10%).
func NewTokenManagerWithBuffer(client Client, scopes []string, refreshBuffer float64) *DefaultTokenManager {
	if refreshBuffer < 0 || refreshBuffer > 1 {
		refreshBuffer = DefaultRefreshBuffer
	}

	return &DefaultTokenManager{
		client:        client,
		scopes:        scopes,
		refreshBuffer: refreshBuffer,
	}
}

// GetToken returns a valid access token, refreshing if necessary.
// This method is thread-safe and handles concurrent requests by ensuring
// only one refresh happens at a time.
func (tm *DefaultTokenManager) GetToken(ctx context.Context) (string, error) {
	// Fast path: check if we have a valid cached token
	tm.mu.RLock()

	if tm.closed {
		tm.mu.RUnlock()

		return "", ErrTokenFetchFailed
	}

	if tm.isTokenValid() {
		token := tm.token.AccessToken
		tm.mu.RUnlock()

		return token, nil
	}

	tm.mu.RUnlock()

	// Slow path: need to refresh the token
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return "", ErrTokenFetchFailed
	}

	// Double-check: another goroutine may have refreshed while we waited for the lock
	if tm.isTokenValid() {
		return tm.token.AccessToken, nil
	}

	// Fetch a new token
	tokenResp, err := tm.client.GetClientCredentialsToken(ctx, tm.scopes)
	if err != nil {
		return "", err //nolint:wrapcheck // oauth2 client errors are already wrapped
	}

	// Calculate expiry time with buffer
	ttl := time.Duration(tokenResp.ExpiresIn) * time.Second
	buffer := time.Duration(float64(ttl) * tm.refreshBuffer)
	tm.expiresAt = time.Now().Add(ttl - buffer)
	tm.token = tokenResp

	return tokenResp.AccessToken, nil
}

// InvalidateToken forces a token refresh on the next GetToken call.
func (tm *DefaultTokenManager) InvalidateToken() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.token = nil
	tm.expiresAt = time.Time{}
}

// Close marks the TokenManager as closed.
// After calling Close, GetToken will return an error.
func (tm *DefaultTokenManager) Close() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.closed = true
	tm.token = nil
}

// GetCachedToken returns the current cached token without refreshing.
// This is useful for testing and debugging.
func (tm *DefaultTokenManager) GetCachedToken() (*TokenResponse, time.Time) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.token == nil {
		return nil, time.Time{}
	}

	// Return a copy to prevent mutation
	tokenCopy := *tm.token

	return &tokenCopy, tm.expiresAt
}

// isTokenValid checks if the current cached token is still valid.
// Must be called with at least a read lock held.
func (tm *DefaultTokenManager) isTokenValid() bool {
	return tm.token != nil && time.Now().Before(tm.expiresAt)
}
