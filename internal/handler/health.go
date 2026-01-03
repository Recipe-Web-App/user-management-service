package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// HealthHandler handles health-related HTTP endpoints.
type HealthHandler struct {
	healthService service.HealthServicer
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(hs service.HealthServicer) *HealthHandler {
	return &HealthHandler{
		healthService: hs,
	}
}

// Health handles GET /health (liveness probe).
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	status := h.healthService.GetHealth(r.Context())
	h.writeJSON(w, http.StatusOK, status)
}

// Ready handles GET /ready (readiness probe).
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	status := h.healthService.GetReadiness(r.Context())
	h.writeJSON(w, http.StatusOK, status)
}

// writeJSON writes a JSON response.
func (h *HealthHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("failed to write response", "error", err)
	}
}
