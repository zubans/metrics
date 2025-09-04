package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func resetFlagsAndArgs(tb testing.TB, args []string) {
	tb.Helper()
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	os.Args = args
}

func writeAgentJSON(tb testing.TB, dir string, v map[string]any) string {
	tb.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		tb.Fatalf("marshal json: %v", err)
	}
	p := filepath.Join(dir, "agent.json")
	if err := os.WriteFile(p, b, 0o600); err != nil {
		tb.Fatalf("write json: %v", err)
	}
	return p
}

func clearAgentEnv(tb testing.TB) {
	tb.Helper()
	_ = os.Unsetenv("CONFIG")
	_ = os.Unsetenv("ADDRESS")
	_ = os.Unsetenv("REPORT_INTERVAL")
	_ = os.Unsetenv("POLL_INTERVAL")
	_ = os.Unsetenv("CRYPTO_KEY")
}

func TestAgentConfig_FileOnly(t *testing.T) {
	t.Cleanup(func() { clearAgentEnv(t) })
	clearAgentEnv(t)
	dir := t.TempDir()
	p := writeAgentJSON(t, dir, map[string]any{
		"address":         "example:9999",
		"report_interval": "1s",
		"poll_interval":   "3s",
		"crypto_key":      "/pub.pem",
	})
	if err := os.Setenv("CONFIG", p); err != nil {
		t.Fatal(err)
	}
	resetFlagsAndArgs(t, []string{"agent"})

	cfg := NewAgentConfig()
	if cfg == nil {
		t.Fatal("cfg is nil")
	}
	if cfg.AddressServer != "example:9999" {
		t.Fatalf("address=%q", cfg.AddressServer)
	}
	if cfg.SendInterval != 1*time.Second {
		t.Fatalf("send=%v", cfg.SendInterval)
	}
	if cfg.PollInterval != 3*time.Second {
		t.Fatalf("poll=%v", cfg.PollInterval)
	}
	if cfg.CryptoKey != "/pub.pem" {
		t.Fatalf("crypto=%q", cfg.CryptoKey)
	}
}

func TestAgentConfig_EnvOverridesFile(t *testing.T) {
	t.Cleanup(func() { clearAgentEnv(t) })
	clearAgentEnv(t)
	dir := t.TempDir()
	p := writeAgentJSON(t, dir, map[string]any{
		"address":         "file:1",
		"report_interval": "2s",
		"poll_interval":   "4s",
		"crypto_key":      "file.pem",
	})
	_ = os.Setenv("CONFIG", p)
	_ = os.Setenv("ADDRESS", "env:2")
	_ = os.Setenv("REPORT_INTERVAL", "5s")
	_ = os.Setenv("POLL_INTERVAL", "7s")
	_ = os.Setenv("CRYPTO_KEY", "env.pem")
	resetFlagsAndArgs(t, []string{"agent"})

	cfg := NewAgentConfig()
	if cfg.AddressServer != "env:2" {
		t.Fatalf("address=%q", cfg.AddressServer)
	}
	if cfg.SendInterval != 5*time.Second {
		t.Fatalf("send=%v", cfg.SendInterval)
	}
	if cfg.PollInterval != 7*time.Second {
		t.Fatalf("poll=%v", cfg.PollInterval)
	}
	if cfg.CryptoKey != "env.pem" {
		t.Fatalf("crypto=%q", cfg.CryptoKey)
	}
}

func TestAgentConfig_FlagsOverrideEnvAndFile(t *testing.T) {
	t.Cleanup(func() { clearAgentEnv(t) })
	clearAgentEnv(t)
	dir := t.TempDir()
	p := writeAgentJSON(t, dir, map[string]any{
		"address":         "file:1",
		"report_interval": "2s",
		"poll_interval":   "4s",
		"crypto_key":      "file.pem",
	})
	_ = os.Setenv("CONFIG", p)
	_ = os.Setenv("ADDRESS", "env:2")
	_ = os.Setenv("REPORT_INTERVAL", "5s")
	_ = os.Setenv("POLL_INTERVAL", "7s")
	_ = os.Setenv("CRYPTO_KEY", "env.pem")

	resetFlagsAndArgs(t, []string{"agent",
		"-a", "flag:3",
		"-r", "9",
		"-p", "11",
		"-crypto-key", "flag.pem",
	})

	cfg := NewAgentConfig()
	if cfg.AddressServer != "flag:3" {
		t.Fatalf("address=%q", cfg.AddressServer)
	}
	if cfg.SendInterval != 9*time.Second {
		t.Fatalf("send=%v", cfg.SendInterval)
	}
	if cfg.PollInterval != 11*time.Second {
		t.Fatalf("poll=%v", cfg.PollInterval)
	}
	if cfg.CryptoKey != "flag.pem" {
		t.Fatalf("crypto=%q", cfg.CryptoKey)
	}
}
