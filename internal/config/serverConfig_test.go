package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func resetServerFlagsArgs(tb testing.TB, args []string) {
	tb.Helper()
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	os.Args = args
}

func writeServerJSON(tb testing.TB, dir string, v map[string]any) string {
	tb.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		tb.Fatalf("marshal: %v", err)
	}
	p := filepath.Join(dir, "server.json")
	if err := os.WriteFile(p, b, 0o600); err != nil {
		tb.Fatalf("write: %v", err)
	}
	return p
}

func clearServerEnv(tb testing.TB) {
	tb.Helper()
	_ = os.Unsetenv("CONFIG")
	_ = os.Unsetenv("ADDRESS")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("STORE_INTERVAL")
	_ = os.Unsetenv("FILE_STORAGE_PATH")
	_ = os.Unsetenv("RESTORE")
	_ = os.Unsetenv("DATABASE_DSN")
	_ = os.Unsetenv("CRYPTO_KEY")
}

func TestServerConfig_FileOnly(t *testing.T) {
	t.Cleanup(func() { clearServerEnv(t) })
	clearServerEnv(t)
	dir := t.TempDir()
	p := writeServerJSON(t, dir, map[string]any{
		"address":        "srv:1",
		"restore":        false,
		"store_interval": "42s",
		"store_file":     "file.db",
		"database_dsn":   "dsn://file",
		"crypto_key":     "file.pem",
	})
	_ = os.Setenv("CONFIG", p)
	resetServerFlagsArgs(t, []string{"server"})

	cfg := NewServerConfig()
	if cfg.RunAddr != "srv:1" {
		t.Fatalf("addr=%q", cfg.RunAddr)
	}
	if cfg.Restore != false {
		t.Fatalf("restore=%v", cfg.Restore)
	}
	if cfg.StoreInterval != 42*time.Second {
		t.Fatalf("storeInt=%v", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != "file.db" {
		t.Fatalf("file=%q", cfg.FileStoragePath)
	}
	if cfg.DBCfg != "dsn://file" {
		t.Fatalf("dsn=%q", cfg.DBCfg)
	}
	if cfg.CryptoKey != "file.pem" {
		t.Fatalf("crypto=%q", cfg.CryptoKey)
	}
}

func TestServerConfig_EnvOverridesFile(t *testing.T) {
	t.Cleanup(func() { clearServerEnv(t) })
	clearServerEnv(t)
	dir := t.TempDir()
	p := writeServerJSON(t, dir, map[string]any{
		"address":        "file:1",
		"restore":        false,
		"store_interval": "42s",
		"store_file":     "file.db",
		"database_dsn":   "dsn://file",
		"crypto_key":     "file.pem",
	})
	_ = os.Setenv("CONFIG", p)
	_ = os.Setenv("ADDRESS", "env:2")
	_ = os.Setenv("RESTORE", "true")
	_ = os.Setenv("STORE_INTERVAL", "2m")
	_ = os.Setenv("FILE_STORAGE_PATH", "env.db")
	_ = os.Setenv("DATABASE_DSN", "dsn://env")
	_ = os.Setenv("CRYPTO_KEY", "env.pem")
	resetServerFlagsArgs(t, []string{"server"})

	cfg := NewServerConfig()
	if cfg.RunAddr != "env:2" {
		t.Fatalf("addr=%q", cfg.RunAddr)
	}
	if cfg.Restore != true {
		t.Fatalf("restore=%v", cfg.Restore)
	}
	if cfg.StoreInterval != 2*time.Minute {
		t.Fatalf("storeInt=%v", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != "env.db" {
		t.Fatalf("file=%q", cfg.FileStoragePath)
	}
	if cfg.DBCfg != "dsn://env" {
		t.Fatalf("dsn=%q", cfg.DBCfg)
	}
	if cfg.CryptoKey != "env.pem" {
		t.Fatalf("crypto=%q", cfg.CryptoKey)
	}
}

func TestServerConfig_FlagsOverrideEnvAndFile(t *testing.T) {
	t.Cleanup(func() { clearServerEnv(t) })
	clearServerEnv(t)
	dir := t.TempDir()
	p := writeServerJSON(t, dir, map[string]any{
		"address":        "file:1",
		"restore":        false,
		"store_interval": "42s",
		"store_file":     "file.db",
		"database_dsn":   "dsn://file",
		"crypto_key":     "file.pem",
	})
	_ = os.Setenv("CONFIG", p)
	_ = os.Setenv("ADDRESS", "env:2")
	_ = os.Setenv("RESTORE", "true")
	_ = os.Setenv("STORE_INTERVAL", "2m")
	_ = os.Setenv("FILE_STORAGE_PATH", "env.db")
	_ = os.Setenv("DATABASE_DSN", "dsn://env")
	_ = os.Setenv("CRYPTO_KEY", "env.pem")

	resetServerFlagsArgs(t, []string{"server",
		"-a", "flag:3",
		"-l", "debug",
		"-i", "5",
		"-f", "flag.db",
		"-d", "dsn://flag",
		"-r=false",
		"-crypto-key", "flag.pem",
	})

	cfg := NewServerConfig()
	if cfg.RunAddr != "flag:3" {
		t.Fatalf("addr=%q", cfg.RunAddr)
	}
	if cfg.FlagLogLevel != "debug" {
		t.Fatalf("log=%q", cfg.FlagLogLevel)
	}
	if cfg.StoreInterval != 5*time.Second {
		t.Fatalf("storeInt=%v", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != "flag.db" {
		t.Fatalf("file=%q", cfg.FileStoragePath)
	}
	if cfg.DBCfg != "dsn://flag" {
		t.Fatalf("dsn=%q", cfg.DBCfg)
	}
	if cfg.Restore != false {
		t.Fatalf("restore=%v", cfg.Restore)
	}
	if cfg.CryptoKey != "flag.pem" {
		t.Fatalf("crypto=%q", cfg.CryptoKey)
	}
}
