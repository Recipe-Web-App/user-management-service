package dependency_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClearCacheEndpoint(t *testing.T) {
	t.Parallel()

	reqBody := `{"keyPattern": "user:*"}`
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/api/v1/user-management/admin/cache/clear",
		bytes.NewBufferString(reqBody),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", uuid.New().String())

	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var data map[string]any

	err = json.Unmarshal(rr.Body.Bytes(), &data)
	require.NoError(t, err)

	assert.Equal(t, "user:*", data["pattern"])
	clearedCount, ok := data["clearedCount"].(float64)
	require.True(t, ok, "clearedCount should be float64")
	assert.InDelta(t, 10.0, clearedCount, 0.1) // Mock returns 10
}
