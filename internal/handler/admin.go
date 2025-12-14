//nolint:mnd // placeholder values for stub handlers
package handler

import (
	"net/http"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
)

// AdminHandler handles admin HTTP endpoints.
type AdminHandler struct{}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// GetRedisSessionStats handles GET /admin/redis/session-stats.
func (h *AdminHandler) GetRedisSessionStats(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.RedisSessionStatsResponse{
		TotalSessions:  100,
		ActiveSessions: 42,
		MemoryUsage:    "256MB",
		TTLInfo: map[string]any{
			"min_ttl": "1h",
			"max_ttl": "24h",
			"avg_ttl": "12h",
		},
	})
}

// GetUserStats handles GET /admin/users/stats.
func (h *AdminHandler) GetUserStats(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.UserStatsResponse{
		TotalUsers:        1000,
		ActiveUsers:       850,
		InactiveUsers:     150,
		NewUsersToday:     12,
		NewUsersThisWeek:  75,
		NewUsersThisMonth: 250,
	})
}

// GetSystemHealth handles GET /admin/health.
func (h *AdminHandler) GetSystemHealth(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.SystemHealthResponse{
		Status:         "healthy",
		DatabaseStatus: "healthy",
		RedisStatus:    "healthy",
		UptimeSeconds:  86400,
		Version:        "1.0.0",
	})
}

// ForceLogoutUser handles POST /admin/users/{user_id}/force-logout.
func (h *AdminHandler) ForceLogoutUser(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.ForceLogoutResponse{
		Success:         true,
		Message:         "User force-logout triggered",
		SessionsCleared: 3,
	})
}

// ClearAllSessions handles DELETE /admin/redis/sessions.
func (h *AdminHandler) ClearAllSessions(w http.ResponseWriter, _ *http.Request) {
	SuccessResponse(w, http.StatusOK, dto.ClearSessionsResponse{
		Success:         true,
		Message:         "All sessions cleared",
		SessionsCleared: 100,
	})
}
