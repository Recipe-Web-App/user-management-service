package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHealthChecker implements repository.HealthChecker for testing.
type mockHealthChecker struct {
	healthStatus map[string]string
}

func (m *mockHealthChecker) Health(_ context.Context) map[string]string {
	return m.healthStatus
}

func (m *mockHealthChecker) Close() error {
	return nil
}

func TestNewServerWithContainer(t *testing.T) {
	t.Parallel()

	t.Run("uses default values when config is nil", func(t *testing.T) {
		t.Parallel()

		container := &app.Container{
			Config:        nil,
			HealthService: service.NewHealthService(nil, nil),
		}

		srv := NewServerWithContainer(container)
		assert.Equal(t, ":8080", srv.Addr)
		assert.Equal(t, time.Minute, srv.IdleTimeout)
		assert.Equal(t, 10*time.Second, srv.ReadTimeout)
		assert.Equal(t, 30*time.Second, srv.WriteTimeout)
	})

	t.Run("uses config values when config is set", func(t *testing.T) {
		t.Parallel()

		container := &app.Container{
			Config: &config.Config{
				Server: config.ServerConfig{
					Port:         9090,
					IdleTimeout:  2 * time.Minute,
					ReadTimeout:  20 * time.Second,
					WriteTimeout: 60 * time.Second,
				},
			},
			HealthService: service.NewHealthService(nil, nil),
		}

		srv := NewServerWithContainer(container)
		assert.Equal(t, ":9090", srv.Addr)
		assert.Equal(t, 2*time.Minute, srv.IdleTimeout)
		assert.Equal(t, 20*time.Second, srv.ReadTimeout)
		assert.Equal(t, 60*time.Second, srv.WriteTimeout)
	})
}

func TestRegisterRoutesWithHandlers_HealthReady(t *testing.T) {
	t.Parallel()

	mockDB := &mockHealthChecker{
		healthStatus: map[string]string{"status": "up", "message": "database is healthy"},
	}
	mockCache := &mockHealthChecker{
		healthStatus: map[string]string{"status": "up", "message": "redis is healthy"},
	}

	container := &app.Container{
		Config:        nil,
		Database:      mockDB,
		Cache:         mockCache,
		HealthService: service.NewHealthService(mockDB, mockCache),
	}

	srv := NewServerWithContainer(container)
	ts := httptest.NewServer(srv.Handler)
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

	t.Run("metrics endpoint returns 200", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			ts.URL+"/metrics",
			nil,
		)
		require.NoError(t, err)

		resp, err := ts.Client().Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
