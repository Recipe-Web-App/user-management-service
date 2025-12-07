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

func createConfigFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600)
	require.NoError(t, err)
}
