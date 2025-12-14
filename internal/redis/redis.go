// Package redis provides a global interface for Redis database interactions.
package redis

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/redis/go-redis/v9"
)

// Service represents a service that interacts with Redis.
type Service struct {
	client     *redis.Client
	prevStatus string
	mu         sync.Mutex
}

var Instance *Service

// New creates a new Redis service with the given config.
func New(cfg *config.RedisConfig) (*Service, error) {
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	}

	client := redis.NewClient(opts)

	slog.Info("redis client initialized", "addr", opts.Addr)

	return &Service{
		client:     client,
		prevStatus: "unknown",
	}, nil
}

// Init initializes the global Redis instance.
//
// Deprecated: Use New() with dependency injection instead.
func Init() {
	if config.Instance == nil {
		slog.Error("config not loaded, cannot initialize redis")
		return
	}

	svc, err := New(&config.Instance.Redis)
	if err != nil {
		slog.Error("failed to initialize redis", "error", err)
		return
	}

	Instance = svc
}

// Health checks the health of the Redis connection.
// Returns a map with status information.
func (s *Service) Health(ctx context.Context) map[string]string {
	stats := make(map[string]string)

	if s == nil || s.client == nil {
		stats["status"] = "down"
		stats["message"] = "redis instance is nil"

		if s != nil {
			s.logStateChange("down")
		}

		return stats
	}

	// Use provided context with a timeout fallback
	checkCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	_, err := s.client.Ping(checkCtx).Result()
	if err != nil {
		stats["status"] = "down"
		stats["error"] = err.Error()

		s.logStateChange("down")

		return stats
	}

	stats["status"] = "up"
	stats["message"] = "redis is healthy"

	s.logStateChange("up")

	// Pool stats
	poolStats := s.client.PoolStats()
	stats["hits"] = strconv.FormatUint(uint64(poolStats.Hits), 10)
	stats["misses"] = strconv.FormatUint(uint64(poolStats.Misses), 10)
	stats["timeouts"] = strconv.FormatUint(uint64(poolStats.Timeouts), 10)
	stats["total_conns"] = strconv.FormatUint(uint64(poolStats.TotalConns), 10)
	stats["idle_conns"] = strconv.FormatUint(uint64(poolStats.IdleConns), 10)
	stats["stale_conns"] = strconv.FormatUint(uint64(poolStats.StaleConns), 10)

	return stats
}

// Close closes the Redis connection.
func (s *Service) Close() error {
	if s == nil || s.client == nil {
		return nil
	}

	slog.Info("closing redis connection")

	err := s.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close redis connection: %w", err)
	}

	return nil
}

// logStateChange logs the status only when it changes.
func (s *Service) logStateChange(currentStatus string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.prevStatus != currentStatus {
		slog.Info("redis health status changed", "previous", s.prevStatus, "current", currentStatus)
		s.prevStatus = currentStatus
	}
}
