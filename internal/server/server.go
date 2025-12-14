package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
)

// NewServerWithContainer creates server with injected dependencies.
func NewServerWithContainer(container *app.Container) *http.Server {
	cfg := container.Config

	port := 8080
	idleTimeout := time.Minute
	readTimeout := 10 * time.Second
	writeTimeout := 30 * time.Second

	if cfg != nil {
		port = cfg.Server.Port
		idleTimeout = cfg.Server.IdleTimeout
		readTimeout = cfg.Server.ReadTimeout
		writeTimeout = cfg.Server.WriteTimeout
	}

	// Create handlers with dependencies
	handlers := Handlers{
		Health:       handler.NewHealthHandler(container.HealthService),
		User:         handler.NewUserHandler(),
		Social:       handler.NewSocialHandler(),
		Notification: handler.NewNotificationHandler(),
		Admin:        handler.NewAdminHandler(),
		Metrics:      handler.NewMetricsHandler(),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      RegisterRoutesWithHandlers(handlers),
		IdleTimeout:  idleTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return server
}
