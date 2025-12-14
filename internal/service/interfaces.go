package service

import "context"

// HealthServicer defines the health service interface for mocking.
type HealthServicer interface {
	GetHealth(ctx context.Context) HealthStatus
	GetReadiness(ctx context.Context) HealthStatus
}
