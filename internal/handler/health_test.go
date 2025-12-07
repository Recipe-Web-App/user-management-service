package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()
	// Create a request to pass to our handler.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
	require.NoError(t, err)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handler.HealthHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	// Check the response body is what we expect.
	expected := map[string]string{"status": "UP"}

	var actual map[string]string

	err = json.NewDecoder(rr.Body).Decode(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual, "handler returned unexpected body")
}

func TestReadyHandler(t *testing.T) {
	t.Parallel()
	// Create a request to pass to our handler.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", nil)
	require.NoError(t, err)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handler.ReadyHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	// Check the response body is what we expect.
	expected := map[string]string{"status": "READY"}

	var actual map[string]string

	err = json.NewDecoder(rr.Body).Decode(&actual)
	require.NoError(t, err)
	assert.Equal(t, expected, actual, "handler returned unexpected body")
}
