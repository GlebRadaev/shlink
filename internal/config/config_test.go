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
	os.Unsetenv("BASE_URL")
}

func TestParseAndLoadConfig_ValidEnvVars(t *testing.T) {
	// Reset flags, arguments, and environment variables before the test
	resetFlagsAndArgs()
	resetEnv()

	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("BASE_URL", "http://localhost")

	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
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
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL) // corrected to the right default value
}

func TestParseAndLoadConfig_CommandLineArgs(t *testing.T) {
	// Reset flags, arguments, and environment variables before the test
	resetFlagsAndArgs()
	resetEnv()

	// Set environment variables for the test
	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("BASE_URL", "http://localhost")

	// Set command-line arguments for the test
	os.Args = []string{"cmd", "-a", "127.0.0.1:9090", "-b", "http://short.ly"}

	// Call the function to parse the configuration
	cfg, err := config.ParseAndLoadConfig()
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	// Check that the configuration values are loaded from command-line arguments
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddress)
	assert.Equal(t, "http://short.ly", cfg.BaseURL)
}
