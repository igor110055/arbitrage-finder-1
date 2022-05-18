package helpers

import (
	"calc/common/config"
	"calc/internal/adapters/db/postgres"
	"calc/internal/adapters/db/postgres/migrations"
	"net/http"

	"github.com/golang-migrate/migrate/v4/source/httpfs"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/urfave/cli/v2"
)

func NewMigrateCmd(cfg *config.DB) *cli.Command {
	return &cli.Command{
		Name:  "migrate-up",
		Usage: "apply migrations on db",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "c",
				Required: false,
			},
		},
		Action: func(ctx *cli.Context) error {
			cfg := &postgres.Config{
				User:       cfg.User,
				Password:   cfg.Password,
				Host:       cfg.Host,
				Name:       cfg.Name,
				DisableTLS: cfg.DisableTLS,
				CertPath:   cfg.CertPath,
			}
			src, err := httpfs.New(http.FS(migrations.F), ".")
			if err != nil {
				return err
			}

			return postgres.MigrateUp(cfg, "httpfs", src)
		},
	}
}
