package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRedisCacheClient struct {
	mock.Mock
}

func (m *MockRedisCacheClient) ClearCache(ctx context.Context, pattern string) (int, error) {
	args := m.Called(ctx, pattern)
	return args.Int(0), args.Error(1)
}

func TestAdminService_ClearCache(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockRedis := new(MockRedisCacheClient)
		svc := NewAdminService(mockRedis)
		ctx := context.Background()
		pattern := "user:*"

		mockRedis.On("ClearCache", ctx, pattern).Return(10, nil)

		resp, err := svc.ClearCache(ctx, pattern)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 10, resp.ClearedCount)
		assert.Equal(t, pattern, resp.Pattern)
		assert.Equal(t, "Cache cleared successfully", resp.Message)
		mockRedis.AssertExpectations(t)
	})

	t.Run("redis error", func(t *testing.T) {
		t.Parallel()

		mockRedis := new(MockRedisCacheClient)
		svc := NewAdminService(mockRedis)
		ctx := context.Background()
		pattern := "user:*"

		mockRedis.On("ClearCache", ctx, pattern).Return(0, assert.AnError)

		resp, err := svc.ClearCache(ctx, pattern)

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to clear cache")
		mockRedis.AssertExpectations(t)
	})

	t.Run("nil redis client", func(t *testing.T) {
		t.Parallel()

		svc := NewAdminService(nil)
		ctx := context.Background()

		resp, err := svc.ClearCache(ctx, "*")

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "redis client is not initialized")
	})
}
