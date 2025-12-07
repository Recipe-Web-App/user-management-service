package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // Test modifies global config.Instance, cannot run in parallel
func TestNewServer(t *testing.T) {
	t.Run("uses default values when config is nil", func(t *testing.T) {
		// Ensure config.Instance is nil
		config.Instance = nil

		srv := NewServer()
		assert.Equal(t, ":8080", srv.Addr)
		assert.Equal(t, time.Minute, srv.IdleTimeout)
		assert.Equal(t, 10*time.Second, srv.ReadTimeout)
		assert.Equal(t, 30*time.Second, srv.WriteTimeout)
	})

	t.Run("uses config values when config is set", func(t *testing.T) {
		// Mock config
		config.Instance = &config.Config{
			Server: config.ServerConfig{
				Port:         9090,
				IdleTimeout:  2 * time.Minute,
				ReadTimeout:  20 * time.Second,
				WriteTimeout: 60 * time.Second,
			},
		}
		// Clean up
		defer func() { config.Instance = nil }()

		srv := NewServer()
		assert.Equal(t, ":9090", srv.Addr)
		assert.Equal(t, 2*time.Minute, srv.IdleTimeout)
		assert.Equal(t, 20*time.Second, srv.ReadTimeout)
		assert.Equal(t, 60*time.Second, srv.WriteTimeout)
	})
}

func TestRegisterRoutes_HealthReady(t *testing.T) {
	t.Parallel()
	// Setup handlers
	handler := RegisterRoutes()

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	t.Run("health endpoint returns 200", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			ts.URL+"/api/v1/user-management/health",
			nil,
		)
		require.NoError(t, err)

		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ready endpoint returns 200", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			ts.URL+"/api/v1/user-management/ready",
			nil,
		)
		require.NoError(t, err)

		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
