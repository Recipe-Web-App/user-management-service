package component_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
)

var testHandler http.Handler

func TestMain(m *testing.M) {
	// Initialize config with necessary values for router setup
	config.Instance = &config.Config{
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

	// Initialize the router
	testHandler = server.RegisterRoutes()

	code := m.Run()
	os.Exit(code)
}
