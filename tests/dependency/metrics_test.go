package dependency_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const responseDataMapMsg = "response data should be a map"

func TestPerformanceMetricsEndpoint(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/performance",
		nil,
	)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok, responseDataMapMsg)

	assert.Contains(t, data, "requestCounts")
	assert.Contains(t, data, "database")
}

func TestCacheMetricsEndpoint(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/cache",
		nil,
	)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok, responseDataMapMsg)

	assert.Contains(t, data, "hitRate")
}

func TestSystemMetricsEndpoint(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/system",
		nil,
	)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok, responseDataMapMsg)

	assert.Contains(t, data, "system")
	assert.Contains(t, data, "process")
}

func TestDetailedHealthMetricsEndpoint(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/user-management/metrics/health/detailed",
		nil,
	)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]any)
	require.True(t, ok, responseDataMapMsg)

	assert.Contains(t, data, "overallStatus")
	assert.Contains(t, data, "services")
	assert.Contains(t, data, "application")
}
