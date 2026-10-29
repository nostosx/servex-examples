package config

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRejectsUnsafeConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*Config)
	}{
		{"short token", func(c *Config) { c.Token = "short" }},
		{"whitespace token", func(c *Config) { c.Token = strings.Repeat(" ", 32) }},
		{"public HTTP", func(c *Config) { c.Address = "0.0.0.0:8080" }},
		{"public RPC", func(c *Config) { c.GRPCAddress = ":9090" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{Address: "127.0.0.1:0", GRPCAddress: "127.0.0.1:0", Token: strings.Repeat("a", 32)}
			tc.edit(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("unsafe settings accepted")
			}
		})
	}
}

func TestLoadRejectsArguments(t *testing.T) {
	for _, arg := range []string{"--bad", "unexpected"} {
		t.Run(arg, func(t *testing.T) {
			if _, err := Load(context.Background(), "test", []string{arg}, io.Discard); err == nil {
				t.Fatal("invalid argument accepted")
			}
		})
	}
	if _, err := Load(context.Background(), "test", []string{"--help"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatal(err)
	}
}

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	cfg, err := Load(context.Background(), "test", nil, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "127.0.0.1:8080" {
		t.Fatal(cfg)
	}
	if cfg.GRPCAddress != "127.0.0.1:9090" {
		t.Fatal(cfg)
	}
	if cfg.LogFile != "" || cfg.LogLevel != "info" {
		t.Fatal(cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"APP_ADDRESS", "APP_GRPC_ADDRESS", "APP_LOG_LEVEL", "APP_LOG_FILE", "APP_TOKEN", "API_TOKEN"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadPrecedence(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "settings.yaml")
	content := "address: 127.0.0.1:8101\ngrpc_address: 127.0.0.1:9101\nlog_level: warn\ntoken: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	load := func(args ...string) Config {
		t.Helper()
		cfg, err := Load(t.Context(), "test", append([]string{"--config", path}, args...), io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		return cfg
	}
	cfg := load()
	if cfg.LogLevel != "warn" || cfg.Address != "127.0.0.1:8101" || cfg.GRPCAddress != "127.0.0.1:9101" {
		t.Fatal(cfg)
	}
	t.Setenv("API_TOKEN", strings.Repeat("b", 32))
	if cfg := load(); cfg.Token != strings.Repeat("b", 32) {
		t.Fatal("legacy token did not override file")
	}
	t.Setenv("APP_TOKEN", strings.Repeat("c", 32))
	if cfg := load(); cfg.Token != strings.Repeat("c", 32) {
		t.Fatal("APP_TOKEN did not override API_TOKEN")
	}
	t.Setenv("APP_ADDRESS", "127.0.0.1:8102")
	t.Setenv("APP_GRPC_ADDRESS", "127.0.0.1:9102")
	t.Setenv("APP_LOG_LEVEL", "debug")
	t.Setenv("APP_LOG_FILE", "from-env.log")
	cfg = load()
	if cfg.LogFile != "from-env.log" || cfg.LogLevel != "debug" || cfg.Address != "127.0.0.1:8102" || cfg.GRPCAddress != "127.0.0.1:9102" {
		t.Fatal(cfg)
	}
	// An explicitly supplied default must still override a different environment value.
	cfg = load("--log-file", "", "--log-level", "info", "--addr", "127.0.0.1:8080", "--grpc-addr", "127.0.0.1:9090")
	if cfg.LogFile != "" || cfg.LogLevel != "info" || cfg.Address != "127.0.0.1:8080" || cfg.GRPCAddress != "127.0.0.1:9090" {
		t.Fatal(cfg)
	}
	t.Setenv("APP_TOKEN", "")
	if _, err := Load(t.Context(), "test", []string{"--config", path}, io.Discard); err == nil {
		t.Fatal("explicit empty token accepted")
	}
}

func TestLoadRejectsInvalidSources(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.yaml")
	if err := os.WriteFile(path, []byte("log_level: [unterminated"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"missing file", []string{"--config", filepath.Join(dir, "missing.yaml")}},
		{"malformed file", []string{"--config", path}},
		{"empty path", []string{"--config", ""}},
		{"invalid level", []string{"--log-level", "invalid"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Load(t.Context(), "test", tc.args, io.Discard); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := Load(ctx, "test", nil, io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := Load(t.Context(), "test", []string{"--config", path, "--help"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatal(err)
	}
}
