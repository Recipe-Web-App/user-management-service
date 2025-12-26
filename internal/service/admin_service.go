package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// AdminService handles administrative operations.
type AdminService interface {
	ClearCache(ctx context.Context, keyPattern string) (*dto.CacheClearResponse, error)
}

// RedisCacheClient defines the interface for Redis cache operations needed by AdminService.
type RedisCacheClient interface {
	ClearCache(ctx context.Context, pattern string) (int, error)
}

type adminService struct {
	redis RedisCacheClient
}

// NewAdminService creates a new admin service.
func NewAdminService(redis RedisCacheClient) AdminService {
	return &adminService{
		redis: redis,
	}
}

var (
	ErrRedisNotInitialized = errors.New("redis client is not initialized")
)

// ClearCache clears cache entries matching the given pattern.
func (s *adminService) ClearCache(ctx context.Context, keyPattern string) (*dto.CacheClearResponse, error) {
	if s.redis == nil {
		return nil, ErrRedisNotInitialized
	}

	count, err := s.redis.ClearCache(ctx, keyPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to clear cache: %w", err)
	}

	return &dto.CacheClearResponse{
		Message:      "Cache cleared successfully",
		Pattern:      keyPattern,
		ClearedCount: count,
	}, nil
}
