package admin

import (
	"calc/cmd/admin/helpers"
	"calc/common/config"
	"github.com/urfave/cli/v2"
)

func NewCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "admin",
		Usage: "Admin commands",
		Subcommands: []*cli.Command{
			helpers.NewGenKeysCmd(cfg.Auth),
			helpers.NewMigrateCmd(cfg.Database),
		},
	}
}
