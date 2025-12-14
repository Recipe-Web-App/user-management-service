// Package dto contains Data Transfer Objects for API request/response handling.
package dto

// Response represents a standard API response wrapper.
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error represents an API error response.
type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// HealthResponse represents health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadyResponse represents readiness check response.
type ReadyResponse struct {
	Status   string            `json:"status"`
	Database map[string]string `json:"database,omitempty"`
	Redis    map[string]string `json:"redis,omitempty"`
}
