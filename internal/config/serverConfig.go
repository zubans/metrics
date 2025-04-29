package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"reflect"
	"time"
)

type Config struct {
	RunAddr         string        `env:"ADDRESS"`
	FlagLogLevel    string        `env:"LOG_LEVEL"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	Restore         bool          `env:"RESTORE"`
	DBCfg           string        `env:"DATABASE_DSN"`
}

func NewServerConfig() *Config {
	var db string
	var cfg Config
	var addr string
	var flagLogLevel string
	var storeInterval int
	var storagePath string
	var isRestore bool

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.IntVar(&storeInterval, "i", 300, "store to file interval")
	flag.StringVar(&storagePath, "f", "metric_storage.json", "file storage path")
	flag.StringVar(&db, "d", "postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable", "db credential")
	flag.BoolVar(&isRestore, "r", true, "bool value. Ability to restore metrics from file")

	flag.Parse()

	cfg.RunAddr = addr
	cfg.FlagLogLevel = flagLogLevel
	cfg.StoreInterval = time.Duration(storeInterval) * time.Second
	cfg.FileStoragePath = storagePath
	cfg.Restore = isRestore
	cfg.DBCfg = db

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
