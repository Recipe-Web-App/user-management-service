package notification_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/notification"
)

// MockTokenProvider is a mock implementation of TokenProvider.
type MockTokenProvider struct {
	mock.Mock
}

func (m *MockTokenProvider) GetToken(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestBaseClient_Do_Success(t *testing.T) {
	t.Parallel()

	expectedResponse := notification.BatchNotificationResponse{
		QueuedCount: 1,
		Message:     "success",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/test-path", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	mockTokenProvider := new(MockTokenProvider)
	mockTokenProvider.On("GetToken", mock.Anything).Return("test-token", nil)

	client := notification.NewBaseClient(server.URL, mockTokenProvider, 10*time.Second)

	var resp notification.BatchNotificationResponse

	err := client.Do(context.Background(), http.MethodPost, "/test-path", map[string]string{"key": "value"}, &resp)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.QueuedCount)
	assert.Equal(t, "success", resp.Message)
	mockTokenProvider.AssertExpectations(t)
}

func TestBaseClient_Do_Accepted(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"queued_count": 2}`))
	}))
	defer server.Close()

	client := notification.NewBaseClient(server.URL, nil, 10*time.Second)

	var resp notification.BatchNotificationResponse

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, &resp)

	require.NoError(t, err)
	assert.Equal(t, 2, resp.QueuedCount)
}

func TestBaseClient_Do_BadRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := notification.NewBaseClient(server.URL, nil, 10*time.Second)

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, notification.ErrNotificationBadRequest)
}

func TestBaseClient_Do_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := notification.NewBaseClient(server.URL, nil, 10*time.Second)

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, notification.ErrNotificationForbidden)
}

func TestBaseClient_Do_Forbidden(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := notification.NewBaseClient(server.URL, nil, 10*time.Second)

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, notification.ErrNotificationForbidden)
}

func TestBaseClient_Do_RateLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := notification.NewBaseClient(server.URL, nil, 10*time.Second)

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, notification.ErrNotificationRateLimit)
}

func TestBaseClient_Do_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := notification.NewBaseClient(server.URL, nil, 10*time.Second)

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, notification.ErrNotificationFailed)
}

func TestBaseClient_Do_TokenError(t *testing.T) {
	t.Parallel()

	mockTokenProvider := new(MockTokenProvider)
	mockTokenProvider.On("GetToken", mock.Anything).Return("", assert.AnError)

	client := notification.NewBaseClient("http://localhost", mockTokenProvider, 10*time.Second)

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get auth token")
	mockTokenProvider.AssertExpectations(t)
}

func TestBaseClient_Do_NoTokenProvider(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not have Authorization header when no token provider
		assert.Empty(t, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := notification.NewBaseClient(server.URL, nil, 10*time.Second)

	err := client.Do(context.Background(), http.MethodPost, "/test", nil, nil)

	require.NoError(t, err)
}

func TestBaseClient_Do_DefaultTimeout(t *testing.T) {
	t.Parallel()

	// Test with zero timeout (should use default)
	client := notification.NewBaseClient("http://localhost", nil, 0)
	assert.NotNil(t, client)
}

func TestBaseClient_TrimsTrailingSlash(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/test", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// URL with trailing slash
	client := notification.NewBaseClient(server.URL+"/", nil, 10*time.Second)

	err := client.Do(context.Background(), http.MethodGet, "/api/test", nil, nil)

	require.NoError(t, err)
}

func TestNewBaseClientWithHTTP(t *testing.T) {
	t.Parallel()

	customClient := &http.Client{Timeout: 5 * time.Second}
	client := notification.NewBaseClientWithHTTP("http://localhost", nil, customClient)

	assert.NotNil(t, client)
}
