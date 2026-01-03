package performance_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/dto"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
)

var (
	benchmarkHandler   http.Handler
	benchmarkContainer *app.Container
)

func TestMain(m *testing.M) {
	// Change to project root to load config correctly (since config path is "./config")
	err := os.Chdir("../../")
	if err != nil {
		panic(err)
	}

	// Load real configuration (reads from env vars set by Makefile)
	cfg := config.Load()

	// Disable logging for benchmarks
	slog.SetDefault(slog.New(slog.DiscardHandler))

	// Create container with config
	container, err := app.NewContainer(app.ContainerConfig{
		Config: cfg,
	})
	if err != nil {
		panic(err)
	}

	benchmarkContainer = container

	// Inject mocks for benchmarks where external dependencies aren't available
	container.AdminService = &MockAdminService{}

	// Initialize the router with container
	srv := server.NewServerWithContainer(container)
	benchmarkHandler = srv.Handler

	code := m.Run()
	os.Exit(code)
}

// MockAdminService for benchmarks.
type MockAdminService struct{}

func (m *MockAdminService) ClearCache(ctx context.Context, keyPattern string) (*dto.CacheClearResponse, error) {
	return &dto.CacheClearResponse{
		Message:      "Cache cleared successfully",
		Pattern:      keyPattern,
		ClearedCount: 50,
	}, nil
}
