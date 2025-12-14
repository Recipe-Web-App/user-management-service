package dependency_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"github.com/spf13/viper"
)

var testHandler http.Handler

func TestMain(m *testing.M) {
	// Point viper to the project root config directory
	viper.AddConfigPath("../../config")

	// Load the real configuration from files
	cfg := config.Load()

	// Create container with real config
	container, _ := app.NewContainer(app.ContainerConfig{
		Config: cfg,
	})

	// Initialize the router with container
	srv := server.NewServerWithContainer(container)
	testHandler = srv.Handler

	code := m.Run()
	os.Exit(code)
}
