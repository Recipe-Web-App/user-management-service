// Package repository defines interfaces for data access layer.
package repository

import "context"

// HealthChecker defines the contract for components that can report health status and be closed.
type HealthChecker interface {
	Health(ctx context.Context) map[string]string
	Close() error
}
