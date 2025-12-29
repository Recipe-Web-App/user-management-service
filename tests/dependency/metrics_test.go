package dependency_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
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
	req.Header.Set("X-User-Id", uuid.New().String())

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "requestCounts")
	assert.Contains(t, response, "database")
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
	req.Header.Set("X-User-Id", uuid.New().String())

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "hitRate")
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
	req.Header.Set("X-User-Id", uuid.New().String())

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "system")
	assert.Contains(t, response, "process")
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
	req.Header.Set("X-User-Id", uuid.New().String())

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "overallStatus")
	assert.Contains(t, response, "services")
	assert.Contains(t, response, "application")
}
