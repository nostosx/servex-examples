package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"time"

	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/logging"
	"github.com/nostosx/servex-examples/services/user"
	"github.com/nostosx/servex/app"
)

func main() {
	if err := run(); err != nil {
		logging.Report(err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load(context.Background(), "migrate", "config/config.toml", os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	ctx, stop := app.SignalContext(context.Background())
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	return user.Migrate(ctx, cfg)
}
