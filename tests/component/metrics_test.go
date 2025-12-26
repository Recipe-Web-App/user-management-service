package component

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	internalConfig "github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
)

func TestMetricsEndpoint(t *testing.T) {
	t.Parallel()

	// We manually construct the container to avoid real DB connections
	container := &app.Container{
		Config: &internalConfig.Config{
			Server: internalConfig.ServerConfig{
				Port: 8080,
			},
		},
	}

	// Mock the metrics service
	mockMetricsSvc := &mockMetricsService{}
	container.MetricsService = mockMetricsSvc

	srv := server.NewServerWithContainer(container)

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	// Make request
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		ts.URL+"/api/v1/user-management/metrics/performance",
		nil,
	)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Success bool                           `json:"success"`
		Data    dto.PerformanceMetricsResponse `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Success)

	// Verify values from mock
	assert.Equal(t, 100, response.Data.RequestCounts.TotalRequests)
	assert.InDelta(t, 42.0, response.Data.ResponseTimes.AverageMs, 0.1)
}

// Mock service.
type mockMetricsService struct{}

func (m *mockMetricsService) GetPerformanceMetrics(ctx context.Context) (*dto.PerformanceMetricsResponse, error) {
	return &dto.PerformanceMetricsResponse{
		RequestCounts: dto.RequestCounts{
			TotalRequests: 100,
		},
		ResponseTimes: dto.ResponseTimes{
			AverageMs: 42.0,
		},
	}, nil
}

func (m *mockMetricsService) GetCacheMetrics(ctx context.Context) (*dto.CacheMetricsResponse, error) {
	return &dto.CacheMetricsResponse{
		MemoryUsage:      "1024",
		MemoryUsageHuman: "1KB",
		KeysCount:        100,
		HitRate:          0.9,
	}, nil
}

func (m *mockMetricsService) GetSystemMetrics(ctx context.Context) (*dto.SystemMetricsResponse, error) {
	return &dto.SystemMetricsResponse{
		System: dto.SystemInfo{
			CPUUsagePercent: 25.5,
		},
		Process: dto.ProcessInfo{
			NumThreads: 10,
		},
	}, nil
}

func TestMetricsEndpoint_System(t *testing.T) {
	t.Parallel()

	container := &app.Container{
		Config: &internalConfig.Config{
			Server: internalConfig.ServerConfig{
				Port: 8080,
			},
		},
	}

	mockMetricsSvc := &mockMetricsService{}
	container.MetricsService = mockMetricsSvc

	srv := server.NewServerWithContainer(container)

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		ts.URL+"/api/v1/user-management/metrics/system",
		nil,
	)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Success bool                      `json:"success"`
		Data    dto.SystemMetricsResponse `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Success)

	assert.InDelta(t, 25.5, response.Data.System.CPUUsagePercent, 0.01)
	assert.Equal(t, 10, response.Data.Process.NumThreads)
}
