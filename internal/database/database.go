// Package database provides a global interface for PostgreSQL database interactions.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
)

// Service represents a service that interacts with a database.
type Service struct {
	db *sql.DB
}

// NewWithDB creates a new database service with an existing connection (for testing).
func NewWithDB(db *sql.DB) *Service {
	return &Service{db: db}
}

// GetDB returns the underlying sql.DB instance.
func (s *Service) GetDB() *sql.DB {
	return s.db
}

var Instance *Service

// New creates a new database service with the given config.
func New(cfg *config.PostgresConfig) (*Service, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.Schema,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.DefaultMaxOpenConns)
	db.SetMaxIdleConns(cfg.DefaultMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DefaultConnMaxLifetime)

	return &Service{db: db}, nil
}

// Init initializes the global database instance.
//
// Deprecated: Use New() with dependency injection instead.
func Init() {
	svc, err := New(&config.Instance.Postgres)
	if err != nil {
		slog.Error("failed to open database, continuing without db", "error", err)
		return
	}

	Instance = svc
}

// Health checks the health of the database connection.
// Returns a map with status information.
func (s *Service) Health(ctx context.Context) map[string]string {
	stats := make(map[string]string)

	if s == nil || s.db == nil {
		stats["status"] = "down"
		stats["message"] = "database instance is nil"

		return stats
	}

	// Use provided context with a timeout fallback
	checkCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := s.db.PingContext(checkCtx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = err.Error()

		return stats
	}

	stats["status"] = "up"
	stats["message"] = "database is healthy"

	// Include some pool stats
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	return stats
}

// Close closes the database connection.
func (s *Service) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	slog.Info("closing database connection")

	err := s.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}
