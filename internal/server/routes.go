package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	customMiddleware "github.com/jsamuelsen/recipe-web-app/user-management-service/internal/middleware"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	Health       *handler.HealthHandler
	User         *handler.UserHandler
	Social       *handler.SocialHandler
	Notification *handler.NotificationHandler
	Admin        *handler.AdminHandler
	Metrics      *handler.MetricsHandler
}

// RegisterRoutesWithHandlers creates routes with injected handlers.
func RegisterRoutesWithHandlers(h Handlers, authCfg customMiddleware.AuthConfig) http.Handler {
	r := chi.NewRouter()

	setupMiddleware(r)

	// Prometheus metrics endpoint (public - no auth)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1/user-management", func(r chi.Router) {
		// Health routes - public (kubernetes probes)
		registerHealthRoutes(r, h)

		// Protected routes - require authentication
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.Auth(authCfg))
			registerUserRoutes(r, h)
			registerNotificationRoutes(r, h)
			registerAdminRoutes(r, h)
			registerMetricsRoutes(r, h)
		})
	})

	return r
}

func setupMiddleware(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Metrics)
	r.Use(customMiddleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5)) //nolint:mnd // compression level

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

	timeout := 60 * time.Second //nolint:mnd // default timeout
	if config.Instance != nil {
		timeout = config.Instance.Server.Timeout
	}
	r.Use(middleware.Timeout(timeout))
}

func registerHealthRoutes(r chi.Router, h Handlers) {
	r.Get("/health", h.Health.Health)
	r.Get("/ready", h.Health.Ready)
}

func registerUserRoutes(r chi.Router, h Handlers) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/search", h.User.SearchUsers)
		r.Put("/profile", h.User.UpdateUserProfile)
		r.Post("/account/delete-request", h.User.RequestAccountDeletion)
		r.Delete("/account", h.User.ConfirmAccountDeletion)

		r.Route("/{user_id}", func(r chi.Router) {
			r.Get("/", h.User.GetUserByID)
			r.Get("/profile", h.User.GetUserProfile)
			r.Get("/following", h.Social.GetFollowing)
			r.Get("/followers", h.Social.GetFollowers)
			r.Get("/activity", h.Social.GetUserActivity)
			r.Post("/follow/{target_user_id}", h.Social.FollowUser)
			r.Delete("/follow/{target_user_id}", h.Social.UnfollowUser)
		})
	})
}

func registerNotificationRoutes(r chi.Router, h Handlers) {
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.Notification.GetNotifications)
		r.Delete("/", h.Notification.DeleteNotifications)
		r.Put("/read-all", h.Notification.MarkAllNotificationsRead)
		r.Get("/preferences", h.Notification.GetNotificationPreferences)
		r.Put("/preferences", h.Notification.UpdateNotificationPreferences)
		r.Put("/{notification_id}/read", h.Notification.MarkNotificationRead)
	})
}

func registerAdminRoutes(r chi.Router, h Handlers) {
	r.Route("/admin", func(r chi.Router) {
		r.Get("/users/stats", h.Admin.GetUserStats)
		r.Post("/cache/clear", h.Admin.ClearCache)
	})
}

func registerMetricsRoutes(r chi.Router, h Handlers) {
	r.Route("/metrics", func(r chi.Router) {
		r.Get("/performance", h.Metrics.GetPerformanceMetrics)
		r.Get("/cache", h.Metrics.GetCacheMetrics)
		r.Get("/cache", h.Metrics.GetCacheMetrics)
		// r.Post("/cache/clear", h.Metrics.ClearCache) // Moved to Admin
		r.Get("/system", h.Metrics.GetSystemMetrics)
		r.Get("/health/detailed", h.Metrics.GetDetailedHealthMetrics)
	})
}
