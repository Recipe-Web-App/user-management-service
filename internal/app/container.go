// Package app provides application-level dependency management.
package app

import (
	"errors"
	"fmt"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/redis"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/repository"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/service"
)

// Container holds all application dependencies.
type Container struct {
	Config   *config.Config
	Database repository.HealthChecker
	Cache    repository.HealthChecker

	// Services
	HealthService service.HealthServicer
}

// ContainerConfig holds options for building the container.
type ContainerConfig struct {
	Config   *config.Config
	Database repository.HealthChecker // Optional override for testing
	Cache    repository.HealthChecker // Optional override for testing
}

// NewContainer creates a new dependency container.
func NewContainer(cfg ContainerConfig) (*Container, error) {
	c := &Container{
		Config: cfg.Config,
	}

	// Use provided dependencies or create new ones
	if cfg.Database != nil {
		c.Database = cfg.Database
	} else if cfg.Config != nil {
		db, err := database.New(&cfg.Config.Postgres)
		if err != nil {
			// Non-fatal: continue without database
			c.Database = nil
		} else {
			c.Database = db
		}
	}

	if cfg.Cache != nil {
		c.Cache = cfg.Cache
	} else if cfg.Config != nil {
		cache, err := redis.New(&cfg.Config.Redis)
		if err != nil {
			// Non-fatal: continue without cache
			c.Cache = nil
		} else {
			c.Cache = cache
		}
	}

	// Initialize services
	c.HealthService = service.NewHealthService(c.Database, c.Cache)

	return c, nil
}

// Close cleanly shuts down all dependencies.
func (c *Container) Close() error {
	var errs []error

	if c.Database != nil {
		err := c.Database.Close()
		if err != nil {
			errs = append(errs, fmt.Errorf("database close: %w", err))
		}
	}

	if c.Cache != nil {
		err := c.Cache.Close()
		if err != nil {
			errs = append(errs, fmt.Errorf("cache close: %w", err))
		}
	}

	return errors.Join(errs...)
}
