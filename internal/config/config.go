package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Config holds the configuration settings for the application.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`   // HTTP server address
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`  // Base URL for shortened links
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./storage.txt"` // Path to the storage file
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`                   // Database connection string // postgres://shlink:shlink@localhost:54321/shlink?sslmode=disable
}

// ParseAndLoadConfig reads configuration from environment variables and command-line flags.
func ParseAndLoadConfig() (*Config, error) {
	cfg := &Config{}

	// Parsing environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %v", err)
	}

	// Defining command-line flags
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.Parse()

	return cfg, nil
}
