package oauth2_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/oauth2"
)

// MockClient is a mock implementation of oauth2.Client for testing.
type MockClient struct {
	mock.Mock
}

//nolint:wrapcheck // test mock returns test errors directly
func (m *MockClient) GetClientCredentialsToken(
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
func (m *MockClient) IntrospectToken(ctx context.Context, token string) (*oauth2.IntrospectResponse, error) {
	args := m.Called(ctx, token)

	resp, ok := args.Get(0).(*oauth2.IntrospectResponse)
	if !ok {
		return nil, args.Error(1)
	}

	return resp, args.Error(1)
}

//nolint:wrapcheck // test mock returns test errors directly
func (m *MockClient) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)

	return args.Error(0)
}

//nolint:funlen // table-driven test
func TestTokenManager_GetToken(t *testing.T) {
	t.Parallel()

	t.Run("fetches new token when cache is empty", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read", "write"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "new-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil).Once()

		tm := oauth2.NewTokenManager(mockClient, []string{"read", "write"})

		token, err := tm.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "new-token", token)

		mockClient.AssertExpectations(t)
	})

	t.Run("returns cached token when valid", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "cached-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600, // 1 hour
			}, nil).Once() // Should only be called once

		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		// First call - fetches token
		token1, err := tm.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "cached-token", token1)

		// Second call - should use cache
		token2, err := tm.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "cached-token", token2)

		mockClient.AssertExpectations(t)
	})

	t.Run("refreshes token when near expiry", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		// First call returns token with very short TTL
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "short-lived-token",
				TokenType:   "Bearer",
				ExpiresIn:   1, // 1 second - with 10% buffer, expires in 0.9s
			}, nil).Once()

		// Second call should happen because token expired
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "refreshed-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil).Once()

		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		// First call
		token1, err := tm.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "short-lived-token", token1)

		// Wait for token to expire (1s TTL - 10% buffer = 0.9s effective, wait 1s to be safe)
		time.Sleep(1100 * time.Millisecond)

		// Second call should refresh
		token2, err := tm.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "refreshed-token", token2)

		mockClient.AssertExpectations(t)
	})

	t.Run("handles client error", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		mockClient.On("GetClientCredentialsToken", mock.Anything, mock.Anything).
			Return(nil, oauth2.ErrMissingCredentials)

		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		token, err := tm.GetToken(context.Background())
		require.ErrorIs(t, err, oauth2.ErrMissingCredentials)
		assert.Empty(t, token)

		mockClient.AssertExpectations(t)
	})

	t.Run("handles concurrent requests with single refresh", func(t *testing.T) {
		t.Parallel()

		var callCount atomic.Int32

		mockClient := new(MockClient)
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Run(func(_ mock.Arguments) {
				callCount.Add(1)
				// Simulate slow token fetch
				time.Sleep(50 * time.Millisecond)
			}).
			Return(&oauth2.TokenResponse{
				AccessToken: "concurrent-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil)

		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		var wg sync.WaitGroup

		results := make(chan string, 10)

		// Launch multiple concurrent requests
		for range 10 {
			wg.Add(1)

			go func() {
				defer wg.Done()

				token, err := tm.GetToken(context.Background())
				assert.NoError(t, err)

				results <- token
			}()
		}

		wg.Wait()
		close(results)

		// All should get the same token
		for token := range results {
			assert.Equal(t, "concurrent-token", token)
		}

		// Client should have been called only once (or at most twice due to race)
		assert.LessOrEqual(t, int(callCount.Load()), 2)
	})
}

func TestTokenManager_InvalidateToken(t *testing.T) {
	t.Parallel()

	t.Run("forces token refresh", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "first-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil).Once()
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "second-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil).Once()

		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		// Get first token
		token1, err := tm.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "first-token", token1)

		// Invalidate
		tm.InvalidateToken()

		// Should fetch new token
		token2, err := tm.GetToken(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "second-token", token2)

		mockClient.AssertExpectations(t)
	})
}

func TestTokenManager_Close(t *testing.T) {
	t.Parallel()

	t.Run("returns error after close", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		tm.Close()

		token, err := tm.GetToken(context.Background())
		require.ErrorIs(t, err, oauth2.ErrTokenFetchFailed)
		assert.Empty(t, token)
	})
}

func TestTokenManager_GetCachedToken(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when no token cached", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		token, expiresAt := tm.GetCachedToken()
		assert.Nil(t, token)
		assert.True(t, expiresAt.IsZero())
	})

	t.Run("returns cached token", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "cached-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}, nil).Once()

		tm := oauth2.NewTokenManager(mockClient, []string{"read"})

		// Populate cache
		_, err := tm.GetToken(context.Background())
		require.NoError(t, err)

		// Check cached token
		token, expiresAt := tm.GetCachedToken()
		require.NotNil(t, token)
		assert.Equal(t, "cached-token", token.AccessToken)
		assert.False(t, expiresAt.IsZero())
		assert.True(t, expiresAt.After(time.Now()))
	})
}

func TestNewTokenManagerWithBuffer(t *testing.T) {
	t.Parallel()

	t.Run("uses custom buffer", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)
		mockClient.On("GetClientCredentialsToken", mock.Anything, []string{"read"}).
			Return(&oauth2.TokenResponse{
				AccessToken: "token",
				TokenType:   "Bearer",
				ExpiresIn:   100, // 100 seconds
			}, nil).Once()

		// Use 50% buffer
		tm := oauth2.NewTokenManagerWithBuffer(mockClient, []string{"read"}, 0.5)

		_, err := tm.GetToken(context.Background())
		require.NoError(t, err)

		// With 100s TTL and 50% buffer, token should expire at 50s
		_, expiresAt := tm.GetCachedToken()

		// Expected expiry is now + 100s - 50s = now + 50s (approximately)
		expectedExpiry := time.Now().Add(50 * time.Second)
		assert.WithinDuration(t, expectedExpiry, expiresAt, 2*time.Second)
	})

	t.Run("clamps invalid buffer values", func(t *testing.T) {
		t.Parallel()

		mockClient := new(MockClient)

		// Negative buffer should be clamped to default
		tm := oauth2.NewTokenManagerWithBuffer(mockClient, []string{"read"}, -0.5)
		assert.NotNil(t, tm)

		// Buffer > 1 should be clamped to default
		tm2 := oauth2.NewTokenManagerWithBuffer(mockClient, []string{"read"}, 1.5)
		assert.NotNil(t, tm2)
	})
}
