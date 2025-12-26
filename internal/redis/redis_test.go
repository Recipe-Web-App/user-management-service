package redis

import (
	"context"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	t.Parallel() // Enable parallel execution

	mr, err := miniredis.Run()
	require.NoError(t, err)

	defer mr.Close()

	port, _ := strconv.Atoi(mr.Port())

	// Setup mock config
	config.Instance = &config.Config{
		Redis: config.RedisConfig{
			Host:     mr.Host(),
			Port:     port,
			Database: 0,
			Password: "",
		},
	}

	// Test Init
	Init()
	require.NotNil(t, Instance)
	require.NotNil(t, Instance.client)

	ctx := context.Background()

	// Test Health (Up)
	health := Instance.Health(ctx)
	assert.Equal(t, "up", health["status"])
	assert.Equal(t, "redis is healthy", health["message"])

	// Test Close
	err = Instance.Close()
	require.NoError(t, err)

	// Test Health (Down after close)
	// Note: go-redis client might still try to reconnect, but closed client usually returns error
	health = Instance.Health(ctx)
	assert.Equal(t, "down", health["status"])
}

func TestHealthNilInstance(t *testing.T) {
	t.Parallel() // Enable parallel execution

	var s *Service = nil // Explicitly nil service pointer

	stats := s.Health(context.Background())
	assert.Equal(t, "down", stats["status"])
	assert.Equal(t, "redis instance is nil", stats["message"])
}

func TestClearCache(t *testing.T) {
	t.Parallel()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	defer mr.Close()

	port, _ := strconv.Atoi(mr.Port())
	cfg := &config.RedisConfig{
		Host:     mr.Host(),
		Port:     port,
		Database: 0,
	}

	svc, err := New(cfg)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, svc.Close())
	}()

	ctx := context.Background()

	// Setup keys
	err = svc.client.Set(ctx, "user:1", "data", 0).Err()
	require.NoError(t, err)
	err = svc.client.Set(ctx, "user:2", "data", 0).Err()
	require.NoError(t, err)
	err = svc.client.Set(ctx, "session:1", "data", 0).Err()
	require.NoError(t, err)

	// Test deleting specific pattern
	count, err := svc.ClearCache(ctx, "user:*")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Verify keys
	exists, err := svc.client.Exists(ctx, "user:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	exists, err = svc.client.Exists(ctx, "session:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Test deleting all (default)
	count, err = svc.ClearCache(ctx, "*")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1) // At least session:1
}
