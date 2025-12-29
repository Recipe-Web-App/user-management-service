package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/middleware"
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
		User:         handler.NewUserHandler(container.UserService),
		Social:       handler.NewSocialHandler(container.SocialService),
		Notification: handler.NewNotificationHandler(container.NotificationService),
		Admin:        handler.NewAdminHandler(container.UserService, container.AdminService),
		Metrics:      handler.NewMetricsHandler(container.MetricsService),
	}

	// Build auth middleware config
	authCfg := buildAuthConfig(container)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      RegisterRoutesWithHandlers(handlers, authCfg),
		IdleTimeout:  idleTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return server
}

// buildAuthConfig creates the auth middleware configuration from the container.
func buildAuthConfig(container *app.Container) middleware.AuthConfig {
	cfg := container.Config

	// Default: OAuth2 disabled, use X-User-Id header
	if cfg == nil {
		return middleware.AuthConfig{
			OAuth2Enabled: false,
		}
	}

	// OAuth2 is enabled if either:
	// - oauth2.enabled is true (explicit incoming auth setting)
	// - oauth2.service_enabled is true (service-to-service mode implies incoming auth)
	oauth2Enabled := cfg.OAuth2.Enabled || cfg.OAuth2.ServiceEnabled

	if !oauth2Enabled {
		return middleware.AuthConfig{
			OAuth2Enabled: false,
		}
	}

	// OAuth2 enabled: check for introspection vs local JWT
	return middleware.AuthConfig{
		OAuth2Enabled:        true,
		IntrospectionEnabled: cfg.OAuth2.IntrospectionEnabled,
		JWTSecret:            cfg.OAuth2.JWTSecret,
		OAuth2Client:         container.OAuth2Client,
	}
}
