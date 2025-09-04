package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	AddressServer string        `env:"ADDRESS"`
	SendInterval  time.Duration `env:"REPORT_INTERVAL"`
	PollInterval  time.Duration `env:"POLL_INTERVAL"`
	CryptoKey     string        `env:"CRYPTO_KEY"`
	GRPCAddress   string        `env:"GRPC_ADDRESS"`
	UseGRPC       bool          `env:"USE_GRPC"`
}

type agentFileConfig struct {
	Address        *string `json:"address"`
	ReportInterval *string `json:"report_interval"`
	PollInterval   *string `json:"poll_interval"`
	CryptoKey      *string `json:"crypto_key"`
	GRPCAddress    *string `json:"grpc_address"`
	UseGRPC        *bool   `json:"use_grpc"`
}

func NewAgentConfig() *AgentConfig {
	cfg := AgentConfig{
		AddressServer: "localhost:8080",
		SendInterval:  10 * time.Second,
		PollInterval:  2 * time.Second,
		CryptoKey:     "",
		GRPCAddress:   "localhost:8090",
		UseGRPC:       false,
	}

	configEnvPath := os.Getenv("CONFIG")

	var (
		addrFlag      string
		repIntFlag    int
		pollIntFlag   int
		cryptoFlag    string
		grpcAddrFlag  string
		useGRPCFlag   bool
		configFlag    string
		configFlagAlt string
	)
	flag.StringVar(&addrFlag, "a", cfg.AddressServer, "address and port to run server")
	flag.IntVar(&repIntFlag, "r", int(cfg.SendInterval/time.Second), "report send interval")
	flag.IntVar(&pollIntFlag, "p", int(cfg.PollInterval/time.Second), "poll interval")
	flag.StringVar(&cryptoFlag, "crypto-key", cfg.CryptoKey, "path to RSA public key (PEM)")
	flag.StringVar(&grpcAddrFlag, "grpc-addr", cfg.GRPCAddress, "gRPC server address")
	flag.BoolVar(&useGRPCFlag, "use-grpc", cfg.UseGRPC, "use gRPC instead of HTTP")
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
			var fc agentFileConfig
			if err := json.Unmarshal(data, &fc); err == nil {
				if fc.Address != nil {
					cfg.AddressServer = *fc.Address
				}
				if fc.ReportInterval != nil {
					if d, err := time.ParseDuration(*fc.ReportInterval); err == nil {
						cfg.SendInterval = d
					}
				}
				if fc.PollInterval != nil {
					if d, err := time.ParseDuration(*fc.PollInterval); err == nil {
						cfg.PollInterval = d
					}
				}
				if fc.CryptoKey != nil {
					cfg.CryptoKey = *fc.CryptoKey
				}
				if fc.GRPCAddress != nil {
					cfg.GRPCAddress = *fc.GRPCAddress
				}
				if fc.UseGRPC != nil {
					cfg.UseGRPC = *fc.UseGRPC
				}
			}
		}
	}

	err := env.Parse(&cfg)
	if err != nil {
		return nil
	}

	applyAgentFlagOverrides(&cfg, addrFlag, repIntFlag, pollIntFlag, cryptoFlag, grpcAddrFlag, useGRPCFlag)

	return &cfg
}

func applyAgentFlagOverrides(cfg *AgentConfig, addrFlag string, repIntFlag int, pollIntFlag int, cryptoFlag string, grpcAddrFlag string, useGRPCFlag bool) {
	setFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})
	if setFlags["a"] {
		cfg.AddressServer = addrFlag
	}
	if setFlags["r"] {
		cfg.SendInterval = time.Duration(repIntFlag) * time.Second
	}
	if setFlags["p"] {
		cfg.PollInterval = time.Duration(pollIntFlag) * time.Second
	}
	if setFlags["crypto-key"] {
		cfg.CryptoKey = cryptoFlag
	}
	if setFlags["grpc-addr"] {
		cfg.GRPCAddress = grpcAddrFlag
	}
	if setFlags["use-grpc"] {
		cfg.UseGRPC = useGRPCFlag
	}
}
