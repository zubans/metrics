package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddr      string `env:"ADDRESS"`
	FlagLogLevel string `env:"LOG_LEVEL"`
}

func NewServerConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.FlagLogLevel, "l", "info", "log level")

	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return nil
	}

	return &cfg
}
