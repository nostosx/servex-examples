package main

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/nostosx/servex/app"

	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/logging"
	service "github.com/nostosx/servex-examples/services/user"
)

func main() {
	if err := run(); err != nil {
		logging.Report(err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(context.Background(), "user", "services/user/config/config.toml", os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	ctx, stop := app.SignalContext(context.Background())
	defer stop()
	return service.Run(ctx, cfg)
}
