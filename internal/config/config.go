package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config holds the configuration settings for the application.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`   // HTTP server address
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`  // Base URL for shortened links
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./storage.txt"` // Path to the storage file
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`                   // Database connection string // postgres://shlink:shlink@localhost:54321/shlink?sslmode=disable
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" envDefault:"false"`
	CertPath        string `env:"CERT_PATH" envDefault:"./certs/cert.pem"`
	KeyPath         string `env:"KEY_PATH" envDefault:"./certs/key.pem"`
	ConfigPath      string `env:"CONFIG"`
}

// ParseAndLoadConfig reads configuration from environment variables and command-line flags.
func ParseAndLoadConfig() (*Config, error) {
	cfg := &Config{}

	configPath := os.Getenv("CONFIG")

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %v", err)
	}

	flag.StringVar(&cfg.ConfigPath, "c", configPath, "Path to JSON config file")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base address for shortened URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Path to file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database connection string")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS mode")
	flag.Parse()

	if cfg.ConfigPath != "" {
		jsonData, err := loadFromJSON(cfg.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from JSON: %v", err)
		}
		applyJSONConfig(cfg, jsonData)
	}

	return cfg, nil
}

// loadFromJSON loads config from JSON file
func loadFromJSON(path string) (map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var configData map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&configData); err != nil {
		return nil, err
	}

	return configData, nil
}

// applyJSONConfig updates config with values ​​from the JSON if they are not empty
func applyJSONConfig(cfg *Config, jsonData map[string]interface{}) {
	if val, ok := jsonData["server_address"].(string); ok && val != "" {
		cfg.ServerAddress = val
	}
	if val, ok := jsonData["base_url"].(string); ok && val != "" {
		cfg.BaseURL = val
	}
	if val, ok := jsonData["file_storage_path"].(string); ok && val != "" {
		cfg.FileStoragePath = val
	}
	if val, ok := jsonData["database_dsn"].(string); ok && val != "" {
		cfg.DatabaseDSN = val
	}
	if val, ok := jsonData["enable_https"].(bool); ok {
		cfg.EnableHTTPS = val
	}
	if val, ok := jsonData["cert_path"].(string); ok && val != "" {
		cfg.CertPath = val
	}
	if val, ok := jsonData["key_path"].(string); ok && val != "" {
		cfg.KeyPath = val
	}
}
