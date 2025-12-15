package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const wrongStatusCode = "handler returned wrong status code"
const unexpectedStatusMsg = "handler returned unexpected status"

// mockHealthService implements service.HealthServicer for testing.
type mockHealthService struct {
	healthStatus    service.HealthStatus
	readinessStatus service.HealthStatus
}

func (m *mockHealthService) GetHealth(_ context.Context) service.HealthStatus {
	return m.healthStatus
}

func (m *mockHealthService) GetReadiness(_ context.Context) service.HealthStatus {
	return m.readinessStatus
}

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	mockSvc := &mockHealthService{
		healthStatus: service.HealthStatus{Status: "UP"},
	}
	h := handler.NewHealthHandler(mockSvc)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	h.Health(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, wrongStatusCode)

	var actual service.HealthStatus

	err = json.NewDecoder(rr.Body).Decode(&actual)
	require.NoError(t, err)
	assert.Equal(t, "UP", actual.Status, unexpectedStatusMsg)
}

func TestReadyHandler(t *testing.T) {
	t.Parallel()

	mockSvc := &mockHealthService{
		readinessStatus: service.HealthStatus{
			Status:   "DEGRADED",
			Database: map[string]string{"status": "down", "message": "database not configured"},
			Redis:    map[string]string{"status": "down", "message": "cache not configured"},
		},
	}
	h := handler.NewHealthHandler(mockSvc)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	h.Ready(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, wrongStatusCode)

	var actual service.HealthStatus

	err = json.NewDecoder(rr.Body).Decode(&actual)
	require.NoError(t, err)
	assert.Equal(t, "DEGRADED", actual.Status, unexpectedStatusMsg)
	assert.NotNil(t, actual.Database, "handler should return database stats")
}

func TestReadyHandlerAllUp(t *testing.T) {
	t.Parallel()

	mockSvc := &mockHealthService{
		readinessStatus: service.HealthStatus{
			Status:   "READY",
			Database: map[string]string{"status": "up", "message": "database is healthy"},
			Redis:    map[string]string{"status": "up", "message": "redis is healthy"},
		},
	}
	h := handler.NewHealthHandler(mockSvc)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	h.Ready(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, wrongStatusCode)

	var actual service.HealthStatus

	err = json.NewDecoder(rr.Body).Decode(&actual)
	require.NoError(t, err)
	assert.Equal(t, "READY", actual.Status, unexpectedStatusMsg)
}
