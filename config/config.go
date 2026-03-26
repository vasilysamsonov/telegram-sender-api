package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type (
	Config struct {
		HTTP HTTP
		Log  Log
	}

	HTTP struct {
		Port            string        `env:"APP_PORT" envDefault:"8086"`
		Timeout         time.Duration `env:"HTTP_TIMEOUT" envDefault:"10s"`
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL" envDefault:"info"`
	}
)

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
