package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/zubans/metrics/internal/logger"
	"go.uber.org/zap"
)

type Config struct {
	RunAddr         string        `env:"ADDRESS"`
	FlagLogLevel    string        `env:"LOG_LEVEL"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	Restore         bool          `env:"RESTORE"`
	DBCfg           string        `env:"DATABASE_DSN"`
	CryptoKey       string        `env:"CRYPTO_KEY"`
	TrustedSubnet   string        `env:"TRUSTED_SUBNET"`
	GRPCAddr        string        `env:"GRPC_ADDRESS"`
	EnableGRPC      bool          `env:"ENABLE_GRPC"`
}

type serverFileConfig struct {
	Address       *string `json:"address"`
	Restore       *bool   `json:"restore"`
	StoreInterval *string `json:"store_interval"`
	StoreFile     *string `json:"store_file"`
	DatabaseDSN   *string `json:"database_dsn"`
	CryptoKey     *string `json:"crypto_key"`
	TrustedSubnet *string `json:"trusted_subnet"`
	GRPCAddress   *string `json:"grpc_address"`
	EnableGRPC    *bool   `json:"enable_grpc"`
}

func NewServerConfig() *Config {
	cfg := Config{
		RunAddr:         "localhost:8080",
		FlagLogLevel:    "info",
		StoreInterval:   300 * time.Second,
		FileStoragePath: "metric_storage.json",
		Restore:         true,
		DBCfg:           "",
		CryptoKey:       "",
		TrustedSubnet:   "",
		GRPCAddr:        "localhost:8090",
		EnableGRPC:      false,
	}

	configEnvPath := os.Getenv("CONFIG")

	var (
		addrFlag       string
		flagLogLevel   string
		storeInterval  int
		storagePath    string
		db             string
		isRestore      bool
		cryptoFlag     string
		trustedFlag    string
		grpcAddrFlag   string
		enableGRPCFlag bool
		configFlag     string
		configFlagAlt  string
	)
	flag.StringVar(&addrFlag, "a", cfg.RunAddr, "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", cfg.FlagLogLevel, "log level")
	flag.IntVar(&storeInterval, "i", int(cfg.StoreInterval/time.Second), "store to file interval")
	flag.StringVar(&storagePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&db, "d", cfg.DBCfg, "db credential")
	flag.BoolVar(&isRestore, "r", cfg.Restore, "bool value. Ability to restore metrics from file")
	flag.StringVar(&cryptoFlag, "crypto-key", cfg.CryptoKey, "path to RSA private key (PEM)")
	flag.StringVar(&trustedFlag, "t", cfg.TrustedSubnet, "trusted subnet in CIDR notation")
	flag.StringVar(&grpcAddrFlag, "grpc-addr", cfg.GRPCAddr, "gRPC server address")
	flag.BoolVar(&enableGRPCFlag, "enable-grpc", cfg.EnableGRPC, "enable gRPC server")
	flag.StringVar(&configFlag, "config", "", "path to JSON config file")
	flag.StringVar(&configFlagAlt, "c", "", "path to JSON config file (short)")

	flag.Parse()

	configPath := configFlag
	if configPath == "" {
		configPath = configFlagAlt
	}
	if configPath == "" {
		configPath = configEnvPath
	}

	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			var fc serverFileConfig
			if err := json.Unmarshal(data, &fc); err == nil {
				if fc.Address != nil {
					cfg.RunAddr = *fc.Address
				}
				if fc.Restore != nil {
					cfg.Restore = *fc.Restore
				}
				if fc.StoreInterval != nil {
					if d, err := time.ParseDuration(*fc.StoreInterval); err == nil {
						cfg.StoreInterval = d
					}
				}
				if fc.StoreFile != nil {
					cfg.FileStoragePath = *fc.StoreFile
				}
				if fc.DatabaseDSN != nil {
					cfg.DBCfg = *fc.DatabaseDSN
				}
				if fc.CryptoKey != nil {
					cfg.CryptoKey = *fc.CryptoKey
				}
				if fc.TrustedSubnet != nil {
					cfg.TrustedSubnet = *fc.TrustedSubnet
				}
				if fc.GRPCAddress != nil {
					cfg.GRPCAddr = *fc.GRPCAddress
				}
				if fc.EnableGRPC != nil {
					cfg.EnableGRPC = *fc.EnableGRPC
				}
			}
		}
	}

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

	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	if setFlags["a"] {
		cfg.RunAddr = addrFlag
	}
	if setFlags["l"] {
		cfg.FlagLogLevel = flagLogLevel
	}
	if setFlags["i"] {
		cfg.StoreInterval = time.Duration(storeInterval) * time.Second
	}
	if setFlags["f"] {
		cfg.FileStoragePath = storagePath
	}
	if setFlags["d"] {
		cfg.DBCfg = db
	}
	if setFlags["r"] {
		cfg.Restore = isRestore
	}
	if setFlags["crypto-key"] {
		cfg.CryptoKey = cryptoFlag
	}
	if setFlags["t"] {
		cfg.TrustedSubnet = trustedFlag
	}
	if setFlags["grpc-addr"] {
		cfg.GRPCAddr = grpcAddrFlag
	}
	if setFlags["enable-grpc"] {
		cfg.EnableGRPC = enableGRPCFlag
	}

	return &cfg
}

func RecoveryServer() {
	if r := recover(); r != nil {
		logger.Log.Info("CRITICAL panic occurred", zap.Any("error", r))
		log.Printf("CRITICAL error %v", r)

	}
}
