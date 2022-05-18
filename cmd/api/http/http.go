package http

import (
	"calc/cmd/api/http/handlers"
	"calc/cmd/api/http/handlers/requests"
	"calc/common/config"
	"calc/foundation/jwt"
	"calc/internal/adapters/client/sender"
	"calc/internal/adapters/client/sender/mobizon"
	"calc/internal/adapters/client/sender/mocks"
	"calc/internal/adapters/db/postgres"
	"calc/internal/adapters/db/postgres/migrations"
	"calc/internal/services/auth"
	"calc/internal/services/exchange"
	"calc/internal/services/refresh_token_keeper"
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func NewCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: "API server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "c",
				Required: false,
			},
		},
		Action: func(ctx *cli.Context) error {
			return run(cfg)
		},
	}
}

func run(cfg *config.Config) error {
	log.Info().Msgf("http: Started: Application initializing: version %q", cfg.Version)
	defer log.Info().Msg("http: Completed")

	log.Info().
		Interface("environment", cfg.Env).
		Msg("http: Unmarshal ENVs to environment structure")

	// =========================================================================
	// Start Database
	log.Info().Msgf("http: Initializing database support %q", cfg.Database.Host)

	dbCfg := &postgres.Config{
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Host:         cfg.Database.Host,
		Name:         cfg.Database.Name,
		DisableTLS:   cfg.Database.DisableTLS,
		CertPath:     cfg.Database.CertPath,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	}

	db, err := postgres.NewDB(dbCfg)
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Stack().Err(err).Msg("closing db")
		}
	}()

	src, err := httpfs.New(http.FS(migrations.F), ".")
	if err != nil {
		return err
	}

	if err := postgres.MigrateUp(dbCfg, "httpfs", src); err != nil {
		log.Error().Stack().Err(err).Msg("http: failed to migrate database")
		return err
	}

	// =========================================================================
	// Initialize authentication support
	log.Info().Msg("api: Started: Initializing authentication support")

	jwtAuth, err := jwt.NewFromFiles(
		cfg.Auth.Algorithm,
		cfg.Auth.PrivateKeyFile,
		cfg.Auth.PublicKeyFile,
		jwt.WithAccessTokenLifetime(cfg.Auth.AccessLifetime),
		jwt.WithRefreshTokenLifetime(cfg.Auth.RefreshLifetime),
		jwt.WithRefreshTokenKeeper(refresh_token_keeper.NewService(db.RefreshToken())),
	)
	if err != nil {
		return errors.Wrap(err, "failed to init JWT authenticator from .pem files")
	}

	// =========================================================================
	// Initialize services

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var smsSenderClient sender.Sender
	switch cfg.Env {
	case "dev":
		smsSenderClient = mocks.NewDummySMSSender()
	case "prod":
		smsSenderClient = mobizon.NewMobizon(cfg.Sender)
	default:
		return errors.New("unknown env")
	}

	authService := auth.NewService(
		db.Account(),
		db.PhoneConfirmation(),
		smsSenderClient,
		jwtAuth,
		cfg.Auth.MaxAttempts,
	)

	exchangeService := exchange.NewService(ctx, cfg, db.Arbitrage())

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.

	log.Info().Msg("http: Initializing debugging support")

	go func() {
		log.Info().Msgf("http: Debug Listening %s:%d", cfg.Server.Host, cfg.Server.DebugPort)

		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.DebugPort), http.DefaultServeMux); err != nil {
			log.Error().Stack().Err(err).Msgf("http: Debug Listener closed: %v", err)
		}
	}()

	// =========================================================================
	// Start API Service

	log.Info().Msg("api: Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HttpPort),
		Handler: handlers.API(
			cfg.Version,
			db,
			jwtAuth,
			authService,
			exchangeService,
		),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	requests.SetupValidator()

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Info().Msgf("http: API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdown:
		log.Info().Msgf("http: %v: Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancelShutdown()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(shutdownCtx); err != nil {
			_ = api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}

		log.Info().Msgf("http: %v: Completed shutdown", sig)
	}

	return nil
}
