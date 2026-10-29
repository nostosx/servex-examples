// Package application assembles the monolith host.
package application

import (
	"context"

	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/services/user"
)

// Run hosts the user service using its shared assembly and lifecycle.
func Run(ctx context.Context, cfg config.Config) error {
	return user.Run(ctx, cfg)
}
