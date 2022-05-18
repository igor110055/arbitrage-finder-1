package postgres

import (
	"calc/internal/domain"
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"time"
)

const phoneConfirmationsTable = "phone_confirmations_challenges"

type PhoneConfirmation struct {
	ID                uint64    `db:"id"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
	AccountID         uint64    `db:"account_id"`
	Phone             string    `db:"phone"`
	Code              string    `db:"code"`
	RemainingAttempts int       `db:"remaining_attempts"`
	Used              bool      `db:"used"`
}

type PhoneConfirmationRepo struct {
	db *DB
}

func (r *PhoneConfirmationRepo) Create(ctx context.Context, confirmation *domain.PhoneConfirmation) (*domain.PhoneConfirmation, error) {
	clauses := map[string]interface{}{
		"account_id":         confirmation.AccountID,
		"phone":              confirmation.Phone,
		"code":               confirmation.Code,
		"remaining_attempts": confirmation.RemainingAttempts,
		"used":               confirmation.Used,
	}

	q, args, err := r.db.Sq.Insert(phoneConfirmationsTable).SetMap(clauses).Suffix("RETURNING id").ToSql()
	if err != nil {
		return nil, err
	}

	if err := r.db.GetContext(ctx, &confirmation.ID, q, args); err != nil {
		return nil, err
	}

	return confirmation, nil
}

func (r *PhoneConfirmationRepo) FindByID(ctx context.Context, id uint64) (*domain.PhoneConfirmation, error) {
	q, args, err := r.db.Sq.Select("*").From(phoneConfirmationsTable).Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query `FindByID`")
	}

	var dbConfirmation PhoneConfirmation

	err = r.db.GetContext(ctx, &dbConfirmation, q, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec query `FindByID`")
	}

	return &domain.PhoneConfirmation{
		ID:                dbConfirmation.ID,
		AccountID:         dbConfirmation.AccountID,
		CreatedAt:         dbConfirmation.CreatedAt,
		UpdatedAt:         dbConfirmation.UpdatedAt,
		Phone:             dbConfirmation.Phone,
		Code:              dbConfirmation.Code,
		RemainingAttempts: dbConfirmation.RemainingAttempts,
		Used:              dbConfirmation.Used,
	}, nil
}

func (r *PhoneConfirmationRepo) FindByAccountID(ctx context.Context, accountID uint64) (*domain.PhoneConfirmation, error) {
	q, args, err := r.db.Sq.Select("*").From(phoneConfirmationsTable).Where(squirrel.Eq{"account_id": accountID}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query `FindByAccountID`")
	}

	var dbConfirmation PhoneConfirmation

	err = r.db.GetContext(ctx, &dbConfirmation, q, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec query `FindByAccountID`")
	}

	return &domain.PhoneConfirmation{
		ID:                dbConfirmation.ID,
		AccountID:         dbConfirmation.AccountID,
		CreatedAt:         dbConfirmation.CreatedAt,
		UpdatedAt:         dbConfirmation.UpdatedAt,
		Phone:             dbConfirmation.Phone,
		Code:              dbConfirmation.Code,
		RemainingAttempts: dbConfirmation.RemainingAttempts,
		Used:              dbConfirmation.Used,
	}, nil
}

func (r *PhoneConfirmationRepo) Update(ctx context.Context, confirmation *domain.PhoneConfirmation) error {
	clauses := map[string]interface{}{
		"code":               confirmation.Code,
		"remaining_attempts": confirmation.RemainingAttempts,
		"used":               confirmation.Used,
	}

	q, args, err := r.db.Sq.Update(phoneConfirmationsTable).SetMap(clauses).Where(squirrel.Eq{"id": confirmation.ID}).ToSql()
	if err != nil {
		return errors.Wrap(err, "error build query `Update`")
	}

	if _, err := r.db.ExecContext(ctx, q, args); err != nil {
		return errors.Wrap(err, "failed to exec query `Update`")
	}

	return nil
}
