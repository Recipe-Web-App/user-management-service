package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	customMiddleware "github.com/jsamuelsen/recipe-web-app/user-management-service/internal/middleware"
)

// RegisterRoutesWithHandlers creates routes with injected handlers.
func RegisterRoutesWithHandlers(healthHandler *handler.HealthHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	corsOptions := cors.Options{}
	if config.Instance != nil {
		corsOptions = cors.Options{
			AllowedOrigins:   config.Instance.Cors.AllowedOrigins,
			AllowedMethods:   config.Instance.Cors.AllowedMethods,
			AllowedHeaders:   config.Instance.Cors.AllowedHeaders,
			ExposedHeaders:   config.Instance.Cors.ExposedHeaders,
			AllowCredentials: config.Instance.Cors.AllowCredentials,
			MaxAge:           int(config.Instance.Cors.MaxAge.Seconds()),
		}
	}

	r.Use(cors.Handler(corsOptions))

	timeout := 60 * time.Second
	if config.Instance != nil {
		timeout = config.Instance.Server.Timeout
	}
	r.Use(middleware.Timeout(timeout))

	r.Route("/api/v1/user-management", func(r chi.Router) {
		r.Get("/health", healthHandler.Health)
		r.Get("/ready", healthHandler.Ready)
	})

	return r
}
