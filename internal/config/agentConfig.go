package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"reflect"
	"time"
)

type AgentConfig struct {
	AddressServer string        `env:"ADDRESS"`
	SendInterval  time.Duration `env:"REPORT_INTERVAL"`
	PollInterval  time.Duration `env:"POLL_INTERVAL"`
}

func NewAgentConfig() *AgentConfig {
	var cfg AgentConfig
	var repInt int
	var pollInt int
	var addr string

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&repInt, "r", 10, "report send interval")
	flag.IntVar(&pollInt, "p", 2, "poll interval")
	flag.Parse()

	cfg.AddressServer = addr
	cfg.SendInterval = time.Duration(repInt) * time.Second
	cfg.PollInterval = time.Duration(pollInt) * time.Second

	err := env.ParseWithFuncs(&cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(time.Duration(0)): func(value string) (interface{}, error) {
			num, err := time.ParseDuration(value)
			if err == nil {
				return num, nil
			}
			seconds, err := time.ParseDuration(value + "s")
			if err != nil {
				return nil, err
			}
			return seconds, nil
		},
	},
	)
	if err != nil {
		return nil
	}

	return &cfg
}
