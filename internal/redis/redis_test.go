package redis

import (
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

	// Test Health (Up)
	health := Instance.Health()
	assert.Equal(t, "up", health["status"])
	assert.Equal(t, "redis is healthy", health["message"])

	// Test Close
	err = Instance.Close()
	require.NoError(t, err)

	// Test Health (Down after close)
	// Note: go-redis client might still try to reconnect, but closed client usually returns error
	health = Instance.Health()
	assert.Equal(t, "down", health["status"])
}

func TestHealthNilInstance(t *testing.T) {
	t.Parallel() // Enable parallel execution

	var s *Service = nil // Explicitly nil service pointer

	stats := s.Health()
	assert.Equal(t, "down", stats["status"])
	assert.Equal(t, "redis instance is nil", stats["message"])
}
