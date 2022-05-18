package main

import (
	"calc/cmd/admin"
	"calc/cmd/api/http"
	"calc/common/config"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v2"
	_ "net/http/pprof"
	"os"
	"strings"
)

func main() {
	var args config.Args
	_, err := flags.Parse(&args)
	if err != nil {
		panic(err)
	}

	cfg, err := config.NewConfig(args.ConfigFilename)
	if err != nil {
		panic(err)
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	switch strings.ToLower(cfg.Logger.Format) {
	case "json":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = zerolog.New(os.Stdout).With().Caller().Logger()
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	default:
		log.Error().
			Msgf("unknown log format %s", cfg.Logger.Format)
		os.Exit(1)
	}

	lvl, err := zerolog.ParseLevel(strings.ToLower(cfg.Logger.Level))
	if err != nil {
		log.Error().
			Stack().
			Err(err).
			Msg("parse log level")
		os.Exit(1)
	}

	zerolog.SetGlobalLevel(lvl)

	app := &cli.App{
		Name:    cfg.AppName,
		Version: cfg.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "c",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			http.NewCmd(cfg),
			admin.NewCmd(cfg),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error().Stack().Err(err).Msg("CLI error")
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
