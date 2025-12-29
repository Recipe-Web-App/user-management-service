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

const (
	corsConfigFileName         = "cors.yaml"
	corsConfigFileContents     = "cors:\n  allowedorigins: ['*']"
	databaseConfigFileName     = "database.yaml"
	databaseConfigFileContents = "postgres:\n  defaultmaxopenconns: 10"
	loggingConfigFileName      = "logging.yaml"
	loggingConfigFileContents  = "logging:\n  consoleenabled: true"
	serverConfigFileName       = "server.yaml"
	serverConfigFileContents   = "server:\n  port: 8080"
	oauth2ConfigFileName       = "oauth2.yaml"
	oauth2ConfigFileContents   = `oauth2:
  baseAuthUrl: "http://auth-service.local/api/v1/user-management/auth"
  getTokenPath: "/oauth2/token"
  revokeTokenPath: "/oauth2/revoke"
  introspectionPath: "/oauth2/introspect"`
)

//nolint:paralleltest // t.Chdir modifies process-level working directory, cannot run in parallel
func TestLoadSuccess(t *testing.T) {
	viper.Reset()

	// Create a temporary directory structure for the test
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create dummy config files
	createConfigFile(t, configDir, serverConfigFileName, `
server:
  port: 9090
  timeout: 5s
  idletimeout: 60s
  readtimeout: 5s
  writetimeout: 10s
`)
	createConfigFile(t, configDir, corsConfigFileName, `
cors:
  allowedorigins: ["*"]
  allowedmethods: ["GET", "POST"]
  allowedheaders: ["Content-Type"]
  exposedheaders: ["Link"]
  allowcredentials: true
  maxage: 3600s
`)
	createConfigFile(t, configDir, loggingConfigFileName, `
logging:
  consoleenabled: true
  consolelevel: "debug"
  fileenabled: false
  filelevel: "info"

  format: "json"
`)
	createConfigFile(t, configDir, databaseConfigFileName, `
postgres:
  defaultmaxopenconns: 10
  defaultmaxidleconns: 10
  defaultconnmaxlifetime: 60s
`)
	createConfigFile(t, configDir, oauth2ConfigFileName, oauth2ConfigFileContents)

	// Change working directory to the temp dir so "./config" resolves correctly
	t.Chdir(tmpDir)

	cfg := Load()

	assert.NotNil(t, cfg)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, 5*time.Second, cfg.Server.Timeout)
	assert.Equal(t, []string{"*"}, cfg.Cors.AllowedOrigins)
	assert.True(t, cfg.Logging.ConsoleEnabled)
	assert.Equal(t, "debug", cfg.Logging.ConsoleLevel)
	assert.Equal(t, "http://auth-service.local/api/v1/user-management/auth", cfg.OAuth2.BaseAuthURL)
	assert.Equal(t, "/oauth2/token", cfg.OAuth2.GetTokenPath)
	assert.Equal(t, "/oauth2/revoke", cfg.OAuth2.RevokeTokenPath)
	assert.Equal(t, "/oauth2/introspect", cfg.OAuth2.IntrospectionPath)
}

//nolint:paralleltest // t.Chdir modifies process-level working directory, cannot run in parallel
func TestLoadPanicOnMissingConfig(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create dummy config files for dependencies loaded BEFORE server (database, etc.)
	createConfigFile(t, configDir, corsConfigFileName, corsConfigFileContents)
	createConfigFile(t, configDir, loggingConfigFileName, loggingConfigFileContents)
	createConfigFile(t, configDir, databaseConfigFileName, databaseConfigFileContents)

	t.Chdir(tmpDir)

	assert.PanicsWithValue(t, "server config file not found", func() {
		Load()
	})
}

func TestLoadPanicOnInvalidOAuth2(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create valid base config
	createConfigFile(t, configDir, serverConfigFileName, serverConfigFileContents)
	createConfigFile(t, configDir, corsConfigFileName, corsConfigFileContents)
	createConfigFile(t, configDir, loggingConfigFileName, loggingConfigFileContents)
	createConfigFile(t, configDir, databaseConfigFileName, databaseConfigFileContents)

	t.Chdir(tmpDir)

	// Clear any OAuth2 env vars that might be set from .env.local
	t.Setenv("OAUTH2_CLIENT_ID", "")
	t.Setenv("OAUTH2_CLIENT_SECRET", "")
	t.Setenv("OAUTH2_SERVICE_ENABLED", "")
	t.Setenv("OAUTH2_INTROSPECTION_ENABLED", "")

	// Enable OAuth2 but don't set required fields
	t.Setenv("OAUTH2_ENABLED", "true")

	assert.PanicsWithValue(t, "oauth2.client_id is required when oauth2 is enabled", func() {
		Load()
	})
}

//nolint:paralleltest // t.Chdir modifies process-level working directory, cannot run in parallel
func TestLoadFromEnv(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create minimal config files to satisfy Load() requirements
	createConfigFile(t, configDir, serverConfigFileName, serverConfigFileContents)
	createConfigFile(t, configDir, corsConfigFileName, corsConfigFileContents)
	createConfigFile(t, configDir, loggingConfigFileName, loggingConfigFileContents)
	createConfigFile(t, configDir, databaseConfigFileName, databaseConfigFileContents)

	t.Chdir(tmpDir)

	// Set environment variables using helpers
	setPostgresEnv(t)
	setRedisEnv(t)
	setOAuth2Env(t)

	cfg := Load()

	assert.NotNil(t, cfg)
	assertPostgresConfig(t, cfg)
	assertRedisConfig(t, cfg)
	assertOAuth2Config(t, cfg)
}

func createConfigFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600)
	require.NoError(t, err)
}

//nolint:paralleltest
func TestLoadEnvironmentDefaults(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create minimal config files
	createConfigFile(t, configDir, serverConfigFileName, serverConfigFileContents)
	createConfigFile(t, configDir, corsConfigFileName, corsConfigFileContents)
	createConfigFile(t, configDir, loggingConfigFileName, loggingConfigFileContents)
	createConfigFile(t, configDir, databaseConfigFileName, databaseConfigFileContents)

	t.Chdir(tmpDir)

	// Clear environment variables that might be set from .env.local
	t.Setenv("ENVIRONMENT", "")

	// Case 1: Default value
	cfg := Load()
	assert.Equal(t, "development", cfg.Environment)
}

func TestLoadEnvironmentVariableBinding(t *testing.T) {
	viper.Reset()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	err := os.Mkdir(configDir, 0750)
	require.NoError(t, err)

	// Create minimal config files
	createConfigFile(t, configDir, serverConfigFileName, serverConfigFileContents)
	createConfigFile(t, configDir, corsConfigFileName, corsConfigFileContents)
	createConfigFile(t, configDir, loggingConfigFileName, loggingConfigFileContents)
	createConfigFile(t, configDir, databaseConfigFileName, databaseConfigFileContents)

	t.Chdir(tmpDir)

	// Test ENVIRONMENT variable binding
	t.Setenv("ENVIRONMENT", "staging")

	// Re-bind envs because viper.Reset() clears them
	// In the real application, Load() handles the binding.
	cfg := Load()
	assert.Equal(t, "staging", cfg.Environment)
}

func setPostgresEnv(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_HOST", "db.example.com")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "mydb")
	t.Setenv("POSTGRES_SCHEMA", "myschema")
	t.Setenv("POSTGRES_USER", "myuser")
	t.Setenv("POSTGRES_PASSWORD", "mypass")
}

func setRedisEnv(t *testing.T) {
	t.Helper()
	t.Setenv("REDIS_HOST", "redis.example.com")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_DB", "1")
	t.Setenv("REDIS_PASSWORD", "redispass")
}

func setOAuth2Env(t *testing.T) {
	t.Helper()
	t.Setenv("OAUTH2_ENABLED", "true")
	t.Setenv("OAUTH2_SERVICE_ENABLED", "true")
	t.Setenv("OAUTH2_INTROSPECTION_ENABLED", "true")
	t.Setenv("OAUTH2_CLIENT_ID", "my-client-id")
	t.Setenv("OAUTH2_CLIENT_SECRET", "my-client-secret")
	t.Setenv("OAUTH2_JWT_SECRET", "my-jwt-secret")
	t.Setenv("OAUTH2_AUTH_BASE_URL", "http://env-auth-service.local")
	t.Setenv("OAUTH2_GET_TOKEN_PATH", "/env/token")
	t.Setenv("OAUTH2_REVOKE_TOKEN_PATH", "/env/revoke")
	t.Setenv("OAUTH2_INTROSPECTION_PATH", "/env/introspect")
}

func assertPostgresConfig(t *testing.T, cfg *Config) {
	t.Helper()
	assert.Equal(t, "db.example.com", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "mydb", cfg.Postgres.Database)
	assert.Equal(t, "myschema", cfg.Postgres.Schema)
	assert.Equal(t, "myuser", cfg.Postgres.User)
	assert.Equal(t, "mypass", cfg.Postgres.Password)
}

func assertRedisConfig(t *testing.T, cfg *Config) {
	t.Helper()
	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, 1, cfg.Redis.Database)
	assert.Equal(t, "redispass", cfg.Redis.Password)
}

func assertOAuth2Config(t *testing.T, cfg *Config) {
	t.Helper()
	assert.True(t, cfg.OAuth2.Enabled)
	assert.True(t, cfg.OAuth2.ServiceEnabled)
	assert.True(t, cfg.OAuth2.IntrospectionEnabled)
	assert.Equal(t, "my-client-id", cfg.OAuth2.ClientID)
	assert.Equal(t, "my-client-secret", cfg.OAuth2.ClientSecret)
	assert.Equal(t, "my-jwt-secret", cfg.OAuth2.JWTSecret)
	assert.Equal(t, "http://env-auth-service.local", cfg.OAuth2.BaseAuthURL)
	assert.Equal(t, "/env/token", cfg.OAuth2.GetTokenPath)
	assert.Equal(t, "/env/revoke", cfg.OAuth2.RevokeTokenPath)
	assert.Equal(t, "/env/introspect", cfg.OAuth2.IntrospectionPath)
}
