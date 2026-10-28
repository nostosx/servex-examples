package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"

	"github.com/nostosx/servex/app"

	"github.com/nostosx/servex-examples/infrastructure/config"
	service "github.com/nostosx/servex-examples/services/user"
)

func main() {
	if err := run(); err != nil {
		slog.Error("user stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Parse("user", os.Args[1:], os.Stderr)
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
