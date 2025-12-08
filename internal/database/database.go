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

var Instance *Service

// Init initializes the global database instance.
func Init() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		config.Instance.Postgres.Host,
		config.Instance.Postgres.Port,
		config.Instance.Postgres.User,
		config.Instance.Postgres.Password,
		config.Instance.Postgres.Database,
		config.Instance.Postgres.Schema,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		// This should theoretically not happen with pgx driver unless the driver name is wrong
		// or DSN is drastically malformed. sql.Open doesn't usually connect.
		slog.Error("failed to open database, continuing without db", "error", err)
		return
	}

	// Set some reasonable defaults for a connection pool
	db.SetMaxOpenConns(config.Instance.Postgres.DefaultMaxOpenConns)
	db.SetMaxIdleConns(config.Instance.Postgres.DefaultMaxIdleConns)
	db.SetConnMaxLifetime(config.Instance.Postgres.DefaultConnMaxLifetime)

	Instance = &Service{
		db: db,
	}

	// We do NOT ping here. We want non-fatal startup.
	// We'll let the readiness probe handle checking if it's actually up.
}

// Health checks the health of the database connection.
// Returns a map with status information.
func (s *Service) Health() map[string]string {
	stats := make(map[string]string)

	if s == nil || s.db == nil {
		stats["status"] = "down"
		stats["message"] = "database instance is nil"

		return stats
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.PingContext(ctx)
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
