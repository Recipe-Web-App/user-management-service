package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/redis"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(map[string]string{"status": "UP"})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(jsonResp)
	if err != nil {
		slog.Error("failed to write health response", "error", err)
	}
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	status := "READY"
	statusCode := http.StatusOK
	ctx := r.Context()

	dbStats := database.Instance.Health(ctx)
	redisStats := redis.Instance.Health(ctx)

	if dbStats["status"] != "up" || redisStats["status"] != "up" {
		status = "DEGRADED"
		// Requirement: Return 200 OK even if DB is down to keep the pod in service.
		// Errors will be handled gracefully by individual endpoints.
	}

	resp := map[string]any{
		"status":   status,
		"database": dbStats,
		"redis":    redisStats,
	}

	jsonResp, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_, err := w.Write(jsonResp)
	if err != nil {
		slog.Error("failed to write readiness response", "error", err)
	}
}
