package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest // Mocks global logger
func TestLogger(t *testing.T) {
	// Setup a logger that writes to a buffer
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, nil)
	logger := slog.New(handler)

	// Save existing default logger and restore it after test
	originalLogger := slog.Default()

	slog.SetDefault(logger)
	defer slog.SetDefault(originalLogger)

	// Create a dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap with Logger middleware
	middlewareStack := Logger(nextHandler)

	// Create a request
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	w := httptest.NewRecorder()

	// Execute
	middlewareStack.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Verify logs
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Request handled")
	assert.Contains(t, logOutput, "path=/test-path")
	assert.Contains(t, logOutput, "status=200")
}
