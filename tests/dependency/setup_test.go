package dependency_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/spf13/viper"
)

var testHandler http.Handler

// MockHealthChecker mocks health checking.
type MockHealthChecker struct{}

func (m *MockHealthChecker) Health(ctx context.Context) map[string]string {
	return map[string]string{
		"status": "up",
	}
}

func (m *MockHealthChecker) Close() error {
	return nil
}

func TestMain(m *testing.M) {
	// Point viper to the project root config directory
	viper.AddConfigPath("../../config")

	// Load the real configuration from files
	cfg := config.Load()

	// Create container with injected mock dependencies for health checks
	mockHealth := &MockHealthChecker{}
	container, _ := app.NewContainer(app.ContainerConfig{
		Config:   cfg,
		Database: mockHealth,
		Cache:    mockHealth,
	})

	// Initialize the router with container
	srv := server.NewServerWithContainer(container)
	testHandler = srv.Handler

	code := m.Run()
	os.Exit(code)
}
