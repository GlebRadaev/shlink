package config

import (
	"net"
	"os"

	"github.com/spf13/pflag"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func ParseFlags() *Config {
	cfg := &Config{}

	pflag.StringVarP(&cfg.ServerAddress, "address", "a", "localhost:8080", "HTTP server address")
	pflag.StringVarP(&cfg.BaseURL, "base-url", "b", "http://localhost:8080", "Base address for shortened URL")

	pflag.Parse()

	// Проверяем, что адрес имеет правильный формат
	if _, err := net.ResolveTCPAddr("tcp", cfg.ServerAddress); err != nil {
		pflag.Usage()
		os.Exit(1)
	}

	return cfg
}
