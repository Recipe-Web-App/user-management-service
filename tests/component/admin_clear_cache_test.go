package component_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRedisCacheClient for component tests.
type MockRedisCacheClient struct {
	mock.Mock
}

// Sentinel errors for mocks.
var (
	ErrInvalidTypeAssertion = errors.New("invalid type assertion")
	ErrMockNil              = errors.New("mock returned nil")
)

func (m *MockRedisCacheClient) ClearCache(ctx context.Context, pattern string) (int, error) {
	args := m.Called(ctx, pattern)
	return args.Int(0), args.Error(1)
}

func (m *MockRedisCacheClient) GetCacheMetrics(ctx context.Context) (*dto.CacheMetricsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, ErrMockNil
	}

	resp, ok := args.Get(0).(*dto.CacheMetricsResponse)
	if !ok {
		err := args.Error(1)
		if err != nil {
			return nil, fmt.Errorf("mock error: %w", err)
		}

		return nil, ErrInvalidTypeAssertion
	}

	err := args.Error(1)
	if err != nil {
		return resp, fmt.Errorf("mock error: %w", err)
	}

	return resp, nil
}

func (m *MockRedisCacheClient) Health(ctx context.Context) map[string]string {
	args := m.Called(ctx)

	resp, ok := args.Get(0).(map[string]string)
	if !ok {
		return map[string]string{}
	}

	return resp
}

func TestClearCache_Success(t *testing.T) {
	t.Parallel()

	// Setup Mocks
	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	mockAdminRedis := new(MockRedisCacheClient)

	// Create Service
	userService := service.NewUserService(mockRepo, mockTokenStore)
	adminService := service.NewAdminService(mockAdminRedis)

	// Create Container
	c := &app.Container{
		UserService:  userService,
		AdminService: adminService,
		Config:       config.Instance,
	}
	// Setup Health so router doesn't panic
	c.HealthService = service.NewHealthService(nil, nil)

	// Create Server
	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	// Mock Expectation
	pattern := "user:*"
	mockAdminRedis.On("ClearCache", mock.Anything, pattern).Return(42, nil)

	// Execute
	reqBody := `{"keyPattern": "user:*"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-management/admin/cache/clear", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp struct {
		Success bool                   `json:"success"`
		Data    dto.CacheClearResponse `json:"data"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)
	assert.Equal(t, 42, apiResp.Data.ClearedCount)
	assert.Equal(t, pattern, apiResp.Data.Pattern)

	mockAdminRedis.AssertExpectations(t)
}

func TestClearCache_Success_EmptyBody(t *testing.T) {
	t.Parallel()

	// Setup Mocks
	mockRepo := new(MockUserRepo)
	mockTokenStore := new(MockTokenStore)
	mockAdminRedis := new(MockRedisCacheClient)

	// Create Service
	userService := service.NewUserService(mockRepo, mockTokenStore)
	adminService := service.NewAdminService(mockAdminRedis)

	// Create Container
	c := &app.Container{
		UserService:  userService,
		AdminService: adminService,
		Config:       config.Instance,
	}
	// Setup Health so router doesn't panic
	c.HealthService = service.NewHealthService(nil, nil)

	// Create Server
	srv := server.NewServerWithContainer(c)
	handler := srv.Handler

	// Mock Expectation - Expect "*" as default pattern
	pattern := "*"
	mockAdminRedis.On("ClearCache", mock.Anything, pattern).Return(100, nil)

	// Execute with NO body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user-management/admin/cache/clear", nil)
	// Even without body, content-type is often not present, or maybe application/json
	// If the binder checks header first, we might need to be careful.
	// But binder.BindJSON check r.Body == nil first.

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var apiResp struct {
		Success bool                   `json:"success"`
		Data    dto.CacheClearResponse `json:"data"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &apiResp)
	require.NoError(t, err)
	require.True(t, apiResp.Success)
	assert.Equal(t, 100, apiResp.Data.ClearedCount)
	assert.Equal(t, pattern, apiResp.Data.Pattern)

	mockAdminRedis.AssertExpectations(t)
}
