// Package app provides application-level dependency management.
package app

import (
	"errors"
	"fmt"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/database"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/handler"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/oauth2"
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
	HealthService     service.HealthServicer
	UserService       service.UserService
	SocialService     service.SocialService
	MetricsService    service.MetricsService
	AdminService      service.AdminService
	PreferenceService service.PreferenceService

	// Handlers
	HealthHandler  handler.HealthHandler
	UserHandler    handler.UserHandler
	SocialHandler  handler.SocialHandler
	AdminHandler   handler.AdminHandler
	MetricsHandler handler.MetricsHandler

	// OAuth2
	OAuth2Client oauth2.Client
	TokenManager oauth2.TokenManager
}

// ContainerConfig holds options for building the container.
type ContainerConfig struct {
	Config         *config.Config
	Database       repository.HealthChecker        // Optional override for testing
	Cache          repository.HealthChecker        // Optional override for testing
	UserRepo       repository.UserRepository       // Optional override for testing
	SocialRepo     repository.SocialRepository     // Optional override for testing
	TokenStore     repository.TokenStore           // Optional override for testing
	PreferenceRepo repository.PreferenceRepository // Optional override for testing
}

// NewContainer creates a new dependency container.
func NewContainer(cfg ContainerConfig) (*Container, error) {
	c := &Container{
		Config: cfg.Config,
	}

	initInfrastructure(c, cfg)

	// Initialize services
	c.HealthService = service.NewHealthService(c.Database, c.Cache)

	// Initialize repositories and domain services
	userRepo, socialRepo, tokenStore, preferenceRepo := initRepositories(c, cfg)

	if userRepo != nil {
		c.UserService = service.NewUserService(userRepo, tokenStore)
	}

	if userRepo != nil && socialRepo != nil {
		c.SocialService = service.NewSocialService(userRepo, socialRepo)
	}

	if preferenceRepo != nil {
		c.PreferenceService = service.NewPreferenceService(preferenceRepo)
	}

	initMetricsService(c)
	initAdminService(c)
	initOAuth2(c, cfg)

	return c, nil
}

func initInfrastructure(c *Container, cfg ContainerConfig) {
	// Database
	if cfg.Database != nil {
		c.Database = cfg.Database
	} else if cfg.Config != nil {
		db, err := database.New(&cfg.Config.Postgres)
		if err == nil {
			c.Database = db
		}
	}

	// Cache
	if cfg.Cache != nil {
		c.Cache = cfg.Cache
	} else if cfg.Config != nil {
		cache, err := redis.New(&cfg.Config.Redis)
		if err == nil {
			c.Cache = cache
		}
	}
}

func initRepositories(c *Container, cfg ContainerConfig) (
	repository.UserRepository,
	repository.SocialRepository,
	repository.TokenStore,
	repository.PreferenceRepository,
) {
	var (
		dbService      *database.Service
		userRepo       repository.UserRepository
		socialRepo     repository.SocialRepository
		tokenStore     repository.TokenStore
		preferenceRepo repository.PreferenceRepository
	)

	if svc, ok := c.Database.(*database.Service); ok {
		dbService = svc
	}

	// User Repo
	if cfg.UserRepo != nil {
		userRepo = cfg.UserRepo
	} else if dbService != nil {
		userRepo = repository.NewUserRepository(dbService.GetDB())
	}

	// Social Repo
	if cfg.SocialRepo != nil {
		socialRepo = cfg.SocialRepo
	} else if dbService != nil {
		socialRepo = repository.NewSocialRepository(dbService.GetDB())
	}

	// Token Store
	if cfg.TokenStore != nil {
		tokenStore = cfg.TokenStore
	} else if redisService, ok := c.Cache.(*redis.Service); ok {
		tokenStore = redisService
	}

	// Preference Repo
	if cfg.PreferenceRepo != nil {
		preferenceRepo = cfg.PreferenceRepo
	} else if dbService != nil {
		preferenceRepo = repository.NewPreferenceRepository(dbService.GetDB())
	}

	return userRepo, socialRepo, tokenStore, preferenceRepo
}

func initMetricsService(c *Container) {
	dbService, ok := c.Database.(*database.Service)
	if !ok || dbService == nil {
		return
	}

	var redisClient service.RedisClient
	if rc, ok := c.Cache.(service.RedisClient); ok {
		redisClient = rc
	}

	systemCollector := service.NewSystemCollector()
	c.MetricsService = service.NewMetricsService(dbService, redisClient, systemCollector, c.Config)
}

// Close cleanly shuts down all dependencies.
func (c *Container) Close() error {
	var errs []error

	// Close TokenManager first (depends on OAuth2Client)
	if c.TokenManager != nil {
		c.TokenManager.Close()
	}

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

func initAdminService(c *Container) {
	var redisClient service.RedisCacheClient
	if rc, ok := c.Cache.(service.RedisCacheClient); ok {
		redisClient = rc
	}

	c.AdminService = service.NewAdminService(redisClient)
}

func initOAuth2(c *Container, cfg ContainerConfig) {
	if cfg.Config == nil || !cfg.Config.OAuth2.Enabled {
		return
	}

	// Initialize OAuth2 client for token introspection
	c.OAuth2Client = oauth2.NewOAuth2Client(&cfg.Config.OAuth2)

	// Initialize TokenManager for service-to-service authentication
	if cfg.Config.OAuth2.ServiceEnabled {
		c.TokenManager = oauth2.NewTokenManager(c.OAuth2Client, []string{"user:read", "user:write"})
	}
}
