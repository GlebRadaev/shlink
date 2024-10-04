package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

// ParseAndLoadConfig parses the configuration from environment variables and command-line flags
func ParseAndLoadConfig() *Config {
	cfg := &Config{}

	// Parsing environment variables
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Failed to parse config from env: %v", err)
	}

	// Defining command-line flags
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flag.Parse()

	return cfg
}
