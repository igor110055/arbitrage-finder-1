package postgres

import (
	"calc/internal/adapters/db"
	"context"
	"database/sql"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	DuplicateKeyValueCode        = "23505"
	ForeignKeyViolationErrorCode = "23503"

	txCtxKey = "tx"
)

type DB struct {
	logger                zerolog.Logger
	DB                    *sqlx.DB
	Sq                    *squirrel.StatementBuilderType
	accountRepo           db.AccountRepo
	phoneConfirmationRepo db.PhoneConfirmationRepo
	jwtKeeperRepo         db.RefreshTokenRepo
	arbitrageRepo         db.ArbitrageRepo
}

func NewDB(config *Config) (db.DB, error) {
	dbConnection, err := sqlx.Connect("pgx", config.URL().String())
	if err != nil {
		return nil, err
	}

	dbConnection.SetMaxOpenConns(config.MaxOpenConns)
	dbConnection.SetConnMaxLifetime(connMaxLifetime)

	if config.MaxIdleConns != 0 {
		dbConnection.SetMaxIdleConns(config.MaxIdleConns)
	}

	stmt := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	database := &DB{
		DB:     dbConnection,
		Sq:     &stmt,
		logger: log.Logger.With().Str("logger", "db_log_hook").Logger(),
	}

	if err := database.StatusCheck(context.Background()); err != nil {
		return nil, err
	}

	return database, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func (r *DB) StatusCheck(ctx context.Context) error {
	_, err := r.ExecContext(ctx, `SELECT true`, nil)
	return err
}

// MigrateUp applies migrations
func MigrateUp(cfg *Config, srcName string, src source.Driver) error {
	// Migrate sql scripts in testing database
	m, err := migrate.NewWithSourceInstance(srcName, src, cfg.URL().String())
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (r *DB) SelectContext(ctx context.Context, query string, dest interface{}, args []interface{}, disableLog ...bool) error {
	if len(disableLog) == 0 || disableLog[0] == false {
		r.logger.Info().Msgf("query: %s; args: %v", query, args)
	}

	if tx := r.getTx(ctx); tx != nil {
		if err := tx.SelectContext(ctx, dest, query, args...); err != nil {
			return err
		}
	} else if err := r.DB.SelectContext(ctx, dest, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *DB) GetContext(ctx context.Context, dest interface{}, query string, args []interface{}, disableLog ...bool) error {
	if len(disableLog) == 0 || disableLog[0] == false {
		r.logger.Info().Msgf("query: %s; args: %v", query, args)
	}

	if tx := r.getTx(ctx); tx != nil {
		if err := tx.GetContext(ctx, dest, query, args...); err != nil {
			return err
		}
	} else if err := r.DB.GetContext(ctx, dest, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *DB) ExecContext(ctx context.Context, query string, args []interface{}, disableLog ...bool) (sql.Result, error) {
	if len(disableLog) == 0 || disableLog[0] == false {
		r.logger.Info().Msgf("query: %s; args: %v", query, args)
	}

	if tx := r.getTx(ctx); tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}

	return r.DB.ExecContext(ctx, query, args...)
}

func (r *DB) RunTx(ctx context.Context) (context.Context, func() error, func() error, error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	return context.WithValue(ctx, txCtxKey, tx), tx.Commit, tx.Rollback, nil
}

func (r *DB) Close() error {
	return r.DB.Close()
}

func (r *DB) getTx(ctx context.Context) *sqlx.Tx {
	tx := ctx.Value(txCtxKey)
	if tx != nil {
		return tx.(*sqlx.Tx)
	}

	return nil
}

func (r *DB) Account() db.AccountRepo {
	if r.accountRepo != nil {
		return r.accountRepo
	}

	r.accountRepo = &AccountRepo{
		db: r,
	}

	return r.accountRepo
}

func (r *DB) PhoneConfirmation() db.PhoneConfirmationRepo {
	if r.phoneConfirmationRepo != nil {
		return r.phoneConfirmationRepo
	}

	r.phoneConfirmationRepo = &PhoneConfirmationRepo{
		db: r,
	}

	return r.phoneConfirmationRepo
}

func (r *DB) RefreshToken() db.RefreshTokenRepo {
	if r.jwtKeeperRepo != nil {
		return r.jwtKeeperRepo
	}

	r.jwtKeeperRepo = &RefreshTokenRepo{
		db: r,
	}

	return r.jwtKeeperRepo
}

func (r *DB) Arbitrage() db.ArbitrageRepo {
	if r.arbitrageRepo != nil {
		return r.arbitrageRepo
	}

	r.arbitrageRepo = &ArbitrageRepo{
		db: r,
	}

	return r.arbitrageRepo
}
