package db

import (
	"context"
	"database/sql"
)

type DB interface {
	StatusCheck(ctx context.Context) error
	SelectContext(ctx context.Context, query string, dest interface{}, args []interface{}, disableLog ...bool) error
	GetContext(ctx context.Context, dest interface{}, query string, args []interface{}, disableLog ...bool) error
	ExecContext(ctx context.Context, query string, args []interface{}, disableLog ...bool) (sql.Result, error)
	RunTx(ctx context.Context) (context.Context, func() error, func() error, error)
	Close() error
	Account() AccountRepo
	PhoneConfirmation() PhoneConfirmationRepo
	RefreshToken() RefreshTokenRepo
	Arbitrage() ArbitrageRepo
}
