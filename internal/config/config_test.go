package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/GlebRadaev/shlink/internal/config"

	"github.com/stretchr/testify/assert"
)

func resetFlagsAndArgs() {
	// Reset flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// Restore command-line arguments to their original state
	os.Args = []string{"cmd"}
}

func resetEnv() {
	// Clear environment variables
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("GRPC_SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("CONFIG")
}

func TestParseAndLoadConfig_ValidEnvVars(t *testing.T) {
	// Reset flags, arguments, and environment variables before the test
	resetFlagsAndArgs()
	resetEnv()

	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("GRPC_SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost")

	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "localhost:9090", cfg.GRPCServerAddress)
	assert.Equal(t, "http://localhost", cfg.BaseURL)
}

func TestParseAndLoadConfig_MissingEnvVarsAndArgs(t *testing.T) {
	// Reset flags, arguments, and environment variables before the test
	resetFlagsAndArgs()
	resetEnv()

	// Call the function to parse the configuration
	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	// Check that the configuration is not nil
	assert.NotNil(t, cfg)

	// Check that the configuration values are set to default values
	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "localhost:9090", cfg.GRPCServerAddress)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
}

func TestParseAndLoadConfig_CommandLineArgs(t *testing.T) {
	// Reset flags, arguments, and environment variables before the test
	resetFlagsAndArgs()
	resetEnv()

	// Set environment variables for the test
	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("GRPC_SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost")

	// Set command-line arguments for the test
	os.Args = []string{"cmd", "-a", "127.0.0.1:9090", "-g", "127.0.0.1:9091", "-b", "http://short.ly"}

	// Call the function to parse the configuration
	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	// Check that the configuration values are loaded from command-line arguments
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddress)
	assert.Equal(t, "127.0.0.1:9091", cfg.GRPCServerAddress)
	assert.Equal(t, "http://short.ly", cfg.BaseURL)
}

func TestParseAndLoadConfig_ValidJSONConfig(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv()

	jsonConfig := `{
		"server_address": "192.168.1.0:8080",
		"grpc_server_address": "192.168.1.0:9090",
		"base_url": "http://192.168.1.0:8080",
		"file_storage_path": "./json_storage.txt",
		"enable_https": true
	}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(jsonConfig)
	assert.NoError(t, err)
	tmpFile.Close()

	os.Setenv("CONFIG", tmpFile.Name())

	cfg, err := config.ParseAndLoadConfig()
	assert.NoError(t, err)

	assert.Equal(t, "192.168.1.0:8080", cfg.ServerAddress)
	assert.Equal(t, "192.168.1.0:9090", cfg.GRPCServerAddress)
	assert.Equal(t, "http://192.168.1.0:8080", cfg.BaseURL)
	assert.Equal(t, "./json_storage.txt", cfg.FileStoragePath)
	assert.True(t, cfg.EnableHTTPS)
}

func TestParseAndLoadConfig_InvalidJSON(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv()

	tmpFile, err := os.CreateTemp("", "config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{ "server_address": "localhost:8080", `)
	assert.NoError(t, err)
	tmpFile.Close()

	os.Setenv("CONFIG", tmpFile.Name())

	_, err = config.ParseAndLoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load config from JSON")
}
