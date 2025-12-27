package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // t.Chdir modifies process-level working directory, cannot run in parallel
func TestLoad_Success(t *testing.T) {
	viper.Reset()

	// Create a temporary directory structure for the test
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create dummy config files
	createConfigFile(t, configDir, "server.yaml", `
server:
  port: 9090
  timeout: 5s
  idletimeout: 60s
  readtimeout: 5s
  writetimeout: 10s
`)
	createConfigFile(t, configDir, "cors.yaml", `
cors:
  allowedorigins: ["*"]
  allowedmethods: ["GET", "POST"]
  allowedheaders: ["Content-Type"]
  exposedheaders: ["Link"]
  allowcredentials: true
  maxage: 3600s
`)
	createConfigFile(t, configDir, "logging.yaml", `
logging:
  consoleenabled: true
  consolelevel: "debug"
  fileenabled: false
  filelevel: "info"

  format: "json"
`)
	createConfigFile(t, configDir, "database.yaml", `
postgres:
  defaultmaxopenconns: 10
  defaultmaxidleconns: 10
  defaultconnmaxlifetime: 60s
`)

	// Change working directory to the temp dir so "./config" resolves correctly
	t.Chdir(tmpDir)

	cfg := Load()

	assert.NotNil(t, cfg)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, 5*time.Second, cfg.Server.Timeout)
	assert.Equal(t, []string{"*"}, cfg.Cors.AllowedOrigins)
	assert.True(t, cfg.Logging.ConsoleEnabled)
	assert.Equal(t, "debug", cfg.Logging.ConsoleLevel)
}

//nolint:paralleltest // t.Chdir modifies process-level working directory, cannot run in parallel
func TestLoad_PanicOnMissingConfig(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create only cors and logging, missing server
	createConfigFile(t, configDir, "cors.yaml", "cors:\n  allowedorigins: ['*']")
	createConfigFile(t, configDir, "logging.yaml", "logging:\n  consoleenabled: true")

	t.Chdir(tmpDir)

	assert.PanicsWithValue(t, "server config file not found", func() {
		Load()
	})
}

func TestLoad_FromEnv(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create minimal config files to satisfy Load() requirements
	createConfigFile(t, configDir, "server.yaml", "server:\n  port: 8080")
	createConfigFile(t, configDir, "cors.yaml", "cors:\n  allowedorigins: ['*']")
	createConfigFile(t, configDir, "logging.yaml", "logging:\n  consoleenabled: true")
	createConfigFile(t, configDir, "database.yaml", "postgres:\n  defaultmaxopenconns: 10")

	t.Chdir(tmpDir)

	// Set environment variables using t.Setenv (cleans up automatically)
	t.Setenv("POSTGRES_HOST", "db.example.com")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "mydb")
	t.Setenv("POSTGRES_SCHEMA", "myschema")
	t.Setenv("POSTGRES_USER", "myuser")
	t.Setenv("POSTGRES_PASSWORD", "mypass")

	t.Setenv("REDIS_HOST", "redis.example.com")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_DB", "1")
	t.Setenv("REDIS_PASSWORD", "redispass")

	cfg := Load()

	assert.NotNil(t, cfg)
	assert.Equal(t, "db.example.com", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "mydb", cfg.Postgres.Database)
	assert.Equal(t, "myschema", cfg.Postgres.Schema)
	assert.Equal(t, "myuser", cfg.Postgres.User)
	assert.Equal(t, "mypass", cfg.Postgres.Password)

	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, 1, cfg.Redis.Database)
	assert.Equal(t, "redispass", cfg.Redis.Password)
}

func createConfigFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600)
	require.NoError(t, err)
}

//nolint:paralleltest
func TestLoad_EnvironmentDefaults(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create minimal config files
	createConfigFile(t, configDir, "server.yaml", "server:\n  port: 8080")
	createConfigFile(t, configDir, "cors.yaml", "cors:\n  allowedorigins: ['*']")
	createConfigFile(t, configDir, "logging.yaml", "logging:\n  consoleenabled: true")
	createConfigFile(t, configDir, "database.yaml", "postgres:\n  defaultmaxopenconns: 10")

	t.Chdir(tmpDir)

	// Case 1: Default value
	cfg := Load()
	assert.Equal(t, "development", cfg.Environment)
}

func TestLoad_EnvironmentVariableBinding(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create minimal config files
	createConfigFile(t, configDir, "server.yaml", "server:\n  port: 8080")
	createConfigFile(t, configDir, "cors.yaml", "cors:\n  allowedorigins: ['*']")
	createConfigFile(t, configDir, "logging.yaml", "logging:\n  consoleenabled: true")
	createConfigFile(t, configDir, "database.yaml", "postgres:\n  defaultmaxopenconns: 10")

	t.Chdir(tmpDir)

	// Test ENVIRONMENT variable binding
	t.Setenv("ENVIRONMENT", "staging")

	// Re-bind envs because viper.Reset() clears them
	// In the real application, Load() handles the binding.
	cfg := Load()
	assert.Equal(t, "staging", cfg.Environment)
}
