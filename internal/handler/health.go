package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
	jsonResp, _ := json.Marshal(map[string]string{"status": "READY"})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(jsonResp)
	if err != nil {
		slog.Error("failed to write readiness response", "error", err)
	}
}
