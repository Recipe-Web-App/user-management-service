//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"net/http"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// MetricsHandler handles metrics HTTP endpoints.
type MetricsHandler struct {
	metricsService service.MetricsService
}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler(metricsService service.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
	}
}

// GetPerformanceMetrics handles GET /metrics/performance.
func (h *MetricsHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricsService.GetPerformanceMetrics(r.Context())
	if err != nil {
		InternalErrorResponse(w)
		return
	}

	SuccessResponse(w, http.StatusOK, metrics)
}

// GetCacheMetrics handles GET /metrics/cache.
func (h *MetricsHandler) GetCacheMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricsService.GetCacheMetrics(r.Context())
	if err != nil {
		// If Redis is unavailable, return 503
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	SuccessResponse(w, http.StatusOK, metrics)
}

// GetSystemMetrics handles GET /metrics/system.
// GetSystemMetrics handles GET /metrics/system.
func (h *MetricsHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricsService.GetSystemMetrics(r.Context())
	if err != nil {
		InternalErrorResponse(w)
		return
	}

	SuccessResponse(w, http.StatusOK, metrics)
}

// GetDetailedHealthMetrics handles GET /metrics/health/detailed.
func (h *MetricsHandler) GetDetailedHealthMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricsService.GetDetailedHealthMetrics(r.Context())
	if err != nil {
		InternalErrorResponse(w)
		return
	}

	SuccessResponse(w, http.StatusOK, metrics)
}
