// Package config loads and validates process settings.
package config

import (
	"errors"
	"flag"
	"io"
	"net"
	"os"
	"strings"
)

// Config contains operator-controlled settings for enabled transports.
type Config struct {
	Address     string
	GRPCAddress string
	Token       string
}

// Parse reads flags and environment without creating services. Help returns
// flag.ErrHelp so command entries can exit before starting the application.
func Parse(name string, args []string, output io.Writer) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&cfg.Address, "addr", "127.0.0.1:8080", "loopback HTTP listen address")
	flags.StringVar(&cfg.GRPCAddress, "grpc-addr", "127.0.0.1:9090", "loopback gRPC listen address")
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	if flags.NArg() != 0 {
		return Config{}, errors.New("unexpected positional arguments")
	}
	cfg.Token = os.Getenv("API_TOKEN")
	return cfg, nil
}

// Validate checks settings before creating listeners or application resources.
func (c Config) Validate() error {
	if len(c.Token) < 32 || strings.ContainsFunc(c.Token, func(r rune) bool { return r < '!' || r > '~' }) {
		return errors.New("API_TOKEN must contain at least 32 printable ASCII characters without whitespace")
	}
	if !loopbackAddress(c.Address) || !loopbackAddress(c.GRPCAddress) {
		return errors.New("plaintext listeners must use literal loopback IPs")
	}
	return nil
}

func loopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	ip := net.ParseIP(host)
	return err == nil && ip != nil && ip.IsLoopback()
}
