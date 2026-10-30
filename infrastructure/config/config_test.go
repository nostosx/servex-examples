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
			if _, err := Load(context.Background(), "test", "config/config.toml", []string{arg}, io.Discard); err == nil {
				t.Fatal("invalid argument accepted")
			}
		})
	}
	if _, err := Load(context.Background(), "test", "config/config.toml", []string{"--help"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatal(err)
	}
}

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	cfg, err := Load(context.Background(), "test", "config/config.toml", nil, io.Discard)
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
	path := filepath.Join(t.TempDir(), "settings.toml")
	content := "address = \"127.0.0.1:8101\"\ngrpc_address = \"127.0.0.1:9101\"\nlog_level = \"warn\"\ntoken = \"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	load := func(args ...string) Config {
		t.Helper()
		cfg, err := Load(t.Context(), "test", "config/config.toml", append([]string{"--config", path}, args...), io.Discard)
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
	if _, err := Load(t.Context(), "test", "config/config.toml", []string{"--config", path}, io.Discard); err == nil {
		t.Fatal("explicit empty token accepted")
	}
}

func TestLoadRejectsInvalidSources(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.toml")
	if err := os.WriteFile(path, []byte("log_level = [unterminated"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"missing file", []string{"--config", filepath.Join(dir, "missing.toml")}},
		{"malformed file", []string{"--config", path}},
		{"empty path", []string{"--config", ""}},
		{"invalid level", []string{"--log-level", "invalid"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Load(t.Context(), "test", "config/config.toml", tc.args, io.Discard); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := Load(ctx, "test", "config/config.toml", nil, io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := Load(t.Context(), "test", "config/config.toml", []string{"--config", path, "--help"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatal(err)
	}
}

func TestLoadProcessFiles(t *testing.T) {
	clearConfigEnv(t)
	t.Chdir(t.TempDir())
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	paths := []string{"config/config.toml", "services/order/config/config.toml", "services/payment/config/config.toml"}
	levels := []string{"warn", "debug", "error"}
	for i, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("log_level = \""+levels[i]+"\"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for i, path := range paths {
		cfg, err := Load(t.Context(), "test", path, nil, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.LogLevel != levels[i] {
			t.Fatal("process loaded another configuration", cfg)
		}
	}
	cfg, err := Load(t.Context(), "test", paths[1], []string{"--config", paths[2]}, io.Discard)
	if err != nil || cfg.LogLevel != "error" {
		t.Fatal(cfg, err)
	}
	if err := os.Remove(paths[1]); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load(t.Context(), "test", paths[1], nil, io.Discard)
	if err != nil || cfg.LogLevel != "info" {
		t.Fatal("missing service file must not load root settings", cfg, err)
	}
}

func TestLoadExampleFiles(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	t.Chdir(filepath.Join("..", ".."))
	for _, path := range []string{"config/config.toml", "services/user/config/config.toml"} {
		// Require the checked-in files so a missing example cannot silently use defaults.
		cfg, err := Load(t.Context(), "test", path, []string{"--config", path}, io.Discard)
		if err != nil {
			t.Fatalf("load example %s: %v", path, err)
		}
		if cfg.Address != "127.0.0.1:8080" || cfg.LogLevel != "info" || cfg.LogFile != "" {
			t.Fatal(path, cfg)
		}
		if cfg.GRPCAddress != "127.0.0.1:9090" {
			t.Fatal(path, cfg)
		}
	}
}

func TestLoadCompositionFileAndEnvironment(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[composition]
postgres_dsn = "postgres://from-file"
redis_address = "127.0.0.1:6379"
demo_username = "demo"
demo_password_hash = "explicit-hash"
insecure_cookies = true
allowed_origins = ["http://127.0.0.1:3000"]
[composition.s3]
bucket = "notes"
access_key = "test-key"
secret_key = "test-secret"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_COMPOSITION__POSTGRES_DSN", "postgres://from-env")
	cfg, err := Load(t.Context(), "test", path, []string{"--config", path}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	c := cfg.Composition
	if c.PostgresDSN != "postgres://from-env" || c.RedisAddress != "127.0.0.1:6379" || !c.InsecureCookies || c.DemoUsername != "demo" || c.S3.Bucket != "notes" || len(c.AllowedOrigins) != 1 {
		t.Fatalf("nested configuration not preserved: %+v", c)
	}
}
