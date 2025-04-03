package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr string
}

func NewServerConfig() *Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	return &cfg
}
