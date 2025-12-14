package component_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
)

var testHandler http.Handler

func TestMain(m *testing.M) {
	// Initialize config with necessary values for router setup
	cfg := &config.Config{
		Cors: config.CorsConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			MaxAge:         time.Duration(300) * time.Second,
		},
		Server: config.ServerConfig{
			Timeout: 60 * time.Second,
		},
	}
	config.Instance = cfg

	// Create container with mock dependencies (nil for component tests)
	container, _ := app.NewContainer(app.ContainerConfig{
		Config: cfg,
	})

	// Initialize the router with container
	srv := server.NewServerWithContainer(container)
	testHandler = srv.Handler

	code := m.Run()
	os.Exit(code)
}
