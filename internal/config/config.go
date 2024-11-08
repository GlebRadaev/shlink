package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./storage.txt"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""` // postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
}

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
