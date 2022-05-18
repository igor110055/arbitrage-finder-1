package postgres

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/oklog/ulid/v2"
	"github.com/pkg/errors"
)

const (
	refreshTokensTable = "refresh_tokens"
)

type RefreshTokenRepo struct {
	db *DB
}

func (r *RefreshTokenRepo) Create(ctx context.Context, accountID uint64, token ulid.ULID) error {
	clauses := map[string]interface{}{
		"id":         token,
		"account_id": accountID,
	}

	q, args, err := r.db.Sq.Insert(refreshTokensTable).SetMap(clauses).ToSql()
	if err != nil {
		return errors.Wrap(err, "error build query `Get`")
	}

	if _, err := r.db.ExecContext(ctx, q, args); err != nil {
		return errors.Wrap(err, "failed to exec query `Get`")
	}

	return nil
}

func (r *RefreshTokenRepo) CountByAccountID(ctx context.Context, accountID uint64) (int64, error) {
	q, args, err := r.db.Sq.Select("COUNT(*)").From(refreshTokensTable).Where(squirrel.Eq{"account_id": accountID}).ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "error build query `CountByAccountID`")
	}

	var count int64

	if err := r.db.GetContext(ctx, &count, q, args); err != nil {
		return 0, errors.Wrap(err, "failed to exec query `CountByAccountID`")
	}

	return count, nil
}

func (r *RefreshTokenRepo) Delete(ctx context.Context, token ulid.ULID) (int64, error) {
	q, args, err := r.db.Sq.Delete(refreshTokensTable).Where(squirrel.Eq{"id": token}).ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "error build query `Delete`")
	}

	res, err := r.db.ExecContext(ctx, q, args)
	if err != nil {
		return 0, errors.Wrap(err, "failed to exec query `Delete`")
	}

	return res.RowsAffected()
}

func (r *RefreshTokenRepo) DeleteOldestByAccountID(ctx context.Context, accountID uint64) error {
	subQ, subArgs, err := r.db.Sq.Select("id").From(refreshTokensTable).Where(squirrel.Eq{"account_id": accountID}).OrderBy("created_at").Limit(1).ToSql()
	if err != nil {
		return errors.Wrap(err, "error build sub query `DeleteOldestByAccountID`")
	}

	q, args, err := r.db.Sq.Delete(refreshTokensTable).Where(fmt.Sprintf("id = (%s)", subQ)).ToSql()
	if err != nil {
		return errors.Wrap(err, "error build query `DeleteOldestByAccountID`")
	}

	args = append(args, subArgs...)

	if _, err := r.db.ExecContext(ctx, q, args); err != nil {
		return errors.Wrap(err, "failed to exec query `DeleteOldestByAccountID`")
	}

	return nil
}
