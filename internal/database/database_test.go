package database

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // accessing global Instance
func TestInit(t *testing.T) {
	// Save existing instance to restore after test
	originalInstance := Instance
	originalConfig := config.Instance

	defer func() {
		Instance = originalInstance
		config.Instance = originalConfig
	}()

	config.Instance = &config.Config{
		Postgres: config.PostgresConfig{
			Host:                   "localhost",
			Port:                   5432,
			User:                   "user",
			Password:               "password",
			Database:               "dbname",
			Schema:                 "public",
			DefaultMaxOpenConns:    10,
			DefaultMaxIdleConns:    10,
			DefaultConnMaxLifetime: time.Minute,
		},
	}

	Init()

	assert.NotNil(t, Instance)
	assert.NotNil(t, Instance.db)

	stats := Instance.Health(context.Background())
	assert.NotNil(t, stats)
	// Expect down because we can't connect to this dummy DB
	assert.Equal(t, "down", stats["status"])

	err := Instance.Close()
	assert.NoError(t, err)
}

//nolint:paralleltest // accessing global Instance
func TestHealthNilInstance(t *testing.T) {
	// Save existing instance
	originalInstance := Instance

	defer func() { Instance = originalInstance }()

	// Set nil
	Instance = nil

	var s *Service

	stats := s.Health(context.Background())

	assert.Equal(t, "down", stats["status"])
	assert.Equal(t, "database instance is nil", stats["message"])
}

func TestCloseNilInstance(t *testing.T) {
	t.Parallel()

	var s *Service

	err := s.Close()
	assert.NoError(t, err)
}

func TestHealthUp(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Expect Ping
	mock.ExpectPing()

	s := &Service{db: db}
	stats := s.Health(context.Background())

	assert.Equal(t, "up", stats["status"])
	assert.Equal(t, "database is healthy", stats["message"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthPingError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	defer func() { _ = db.Close() }()

	// Expect Ping to fail
	mock.ExpectPing().WillReturnError(assert.AnError)

	s := &Service{db: db}
	stats := s.Health(context.Background())

	assert.Equal(t, "down", stats["status"])
	assert.Equal(t, assert.AnError.Error(), stats["error"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
