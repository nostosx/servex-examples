package user

import (
	"context"
	"errors"
	"time"

	goosesdk "github.com/pressly/goose/v3"

	"github.com/nostosx/servex-contrib/database/goose"
	gormx "github.com/nostosx/servex-contrib/database/gorm"
	"github.com/nostosx/servex-contrib/lifecycle"

	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/infrastructure/store"
)

// Migrate explicitly applies schema changes using the same pool factory as Build.
// Serving never calls this function. Run one migration process per deployment.
func Migrate(ctx context.Context, cfg config.Config) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if cfg.Composition.PostgresDSN == "" {
		return errors.New("migration requires composition.postgres_dsn")
	}
	scope, err := lifecycle.NewScope(10 * time.Second)
	if err != nil {
		return err
	}
	defer scope.Finish(&err)
	db, err := gormx.OpenPostgres("postgres", gormx.DefaultPostgresConfig(cfg.Composition.PostgresDSN), nil, nil)
	if err != nil {
		return err
	}
	if err := scope.Add(db); err != nil {
		return err
	}
	migrator, err := goose.New(goosesdk.DialectPostgres, db.SQL(), store.Migrations())
	if err != nil {
		return err
	}
	_, err = migrator.Up(ctx)
	return err
}
