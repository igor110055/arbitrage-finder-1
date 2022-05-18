package postgres

import (
	"calc/internal/adapters/db"
	"calc/internal/domain"
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"time"
)

const accountsTable = "accounts"

type Account struct {
	ID        uint64               `db:"id"`
	CreatedAt time.Time            `db:"created_at"`
	UpdatedAt time.Time            `db:"updated_at"`
	Phone     string               `db:"phone"`
	Password  string               `db:"password"`
	Status    domain.AccountStatus `db:"status"`
}

type AccountRepo struct {
	db *DB
}

func (r *AccountRepo) Create(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	clauses := map[string]interface{}{
		"phone":    account.Phone,
		"password": account.Password,
		"status":   account.Status,
	}

	q, args, err := r.db.Sq.Insert(accountsTable).SetMap(clauses).Suffix("RETURNING id").ToSql()
	if err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &account.ID, q, args); err != nil {
		var pgError *pgconn.PgError

		if errors.As(err, &pgError) {
			if pgError.Code == DuplicateKeyValueCode {
				return nil, db.ErrAlreadyExists
			}

			return nil, pgError
		}

		return nil, errors.Wrap(err, "failed to exec query `Get`")
	}

	return account, nil
}

func (r *AccountRepo) FindByID(ctx context.Context, id uint64) (*domain.Account, error) {
	q, args, err := r.db.Sq.Select("*").From(accountsTable).Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query `FindByID`")
	}

	var dbAccount Account

	err = r.db.GetContext(ctx, &dbAccount, q, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec query `FindByID`")
	}

	return &domain.Account{
		ID:       dbAccount.ID,
		Phone:    dbAccount.Phone,
		Password: dbAccount.Password,
		Status:   dbAccount.Status,
	}, nil
}

func (r *AccountRepo) FindByPhone(ctx context.Context, phone string) (*domain.Account, error) {
	q, args, err := r.db.Sq.Select("*").From(accountsTable).Where(squirrel.Eq{"phone": phone}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query `FindByPhone`")
	}

	var dbAccount Account

	err = r.db.GetContext(ctx, &dbAccount, q, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec query `FindByPhone`")
	}

	return &domain.Account{
		ID:       dbAccount.ID,
		Phone:    dbAccount.Phone,
		Password: dbAccount.Password,
		Status:   dbAccount.Status,
	}, nil
}

func (r *AccountRepo) Update(ctx context.Context, account *domain.Account) error {
	clauses := map[string]interface{}{
		"password": account.Password,
		"status":   account.Status,
	}

	q, args, err := r.db.Sq.Update(accountsTable).SetMap(clauses).Where(squirrel.Eq{"id": account.ID}).ToSql()
	if err != nil {
		return errors.Wrap(err, "error build query `Update`")
	}

	if _, err := r.db.ExecContext(ctx, q, args); err != nil {
		return errors.Wrap(err, "failed to exec query `Update`")
	}

	return nil
}
