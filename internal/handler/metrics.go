//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"net/http"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
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
func (h *MetricsHandler) GetCacheMetrics(w http.ResponseWriter, _ *http.Request) {
	// Note: Use MetricsService for this as well in future
	SuccessResponse(w, http.StatusOK, dto.CacheMetricsResponse{
		MemoryUsage:      "268435456",
		MemoryUsageHuman: "256MB",
		KeysCount:        5000,
		HitRate:          0.95,
		ConnectedClients: 10,
		EvictedKeys:      50,
		ExpiredKeys:      1200,
	})
}

// ClearCache handles POST /metrics/cache/clear.
func (h *MetricsHandler) ClearCache(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.CacheClearResponse{
		Message:      "Cache cleared successfully",
		Pattern:      "*",
		ClearedCount: 500,
	})
}

// GetSystemMetrics handles GET /metrics/system.
func (h *MetricsHandler) GetSystemMetrics(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.SystemMetricsResponse{
		Timestamp: time.Now(),
		System: dto.SystemInfo{
			CPUUsagePercent:    25.5,
			MemoryTotalGB:      16.0,
			MemoryUsedGB:       8.5,
			MemoryUsagePercent: 53.1,
			DiskTotalGB:        500.0,
			DiskUsedGB:         150.0,
			DiskUsagePercent:   30.0,
		},
		Process: dto.ProcessInfo{
			MemoryRSSMB: 128.5,
			MemoryVMSMB: 512.0,
			CPUPercent:  2.5,
			NumThreads:  12,
			OpenFiles:   45,
		},
		UptimeSeconds: 86400,
	})
}

// GetDetailedHealthMetrics handles GET /metrics/health/detailed.
func (h *MetricsHandler) GetDetailedHealthMetrics(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.DetailedHealthMetricsResponse{
		Timestamp:     time.Now(),
		OverallStatus: "healthy",
		Services: dto.ServicesHealth{
			Redis: dto.RedisHealth{
				Status:           "healthy",
				ResponseTimeMs:   1.2,
				MemoryUsage:      "256MB",
				ConnectedClients: 10,
				HitRatePercent:   95.0,
			},
			Database: dto.DatabaseHealth{
				Status:            "healthy",
				ResponseTimeMs:    2.5,
				ActiveConnections: 10,
				MaxConnections:    100,
			},
		},
		Application: dto.ApplicationInfo{
			Version:     "1.0.0",
			Environment: "development",
			Features: dto.ApplicationFeatures{
				Authentication:  "enabled",
				Caching:         "enabled",
				Monitoring:      "enabled",
				SecurityHeaders: "enabled",
			},
		},
	})
}
