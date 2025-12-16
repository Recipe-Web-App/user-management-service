// Package repository defines interfaces for data access layer.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// HealthChecker defines the contract for components that can report health status and be closed.
type HealthChecker interface {
	Health(ctx context.Context) map[string]string
	Close() error
}

// TokenStore defines the contract for storing and retrieving tokens.
type TokenStore interface {
	StoreDeleteToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error
	GetDeleteToken(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteDeleteToken(ctx context.Context, userID uuid.UUID) error
}
