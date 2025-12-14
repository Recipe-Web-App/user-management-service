package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/metrics"
)

// Metrics is middleware that records Prometheus metrics for each request.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		metrics.RequestsInFlight.Inc()
		defer metrics.RequestsInFlight.Dec()

		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		// Get route pattern for consistent labeling (avoids high cardinality)
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = "unknown"
		}

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(ww.Status())

		metrics.RequestsTotal.WithLabelValues(r.Method, routePattern, status).Inc()
		metrics.RequestDuration.WithLabelValues(r.Method, routePattern).Observe(duration)
	})
}
