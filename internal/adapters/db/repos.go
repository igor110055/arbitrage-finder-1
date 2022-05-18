package db

import (
	"calc/internal/adapters/db/filters"
	"calc/internal/domain"
	"context"
	"github.com/oklog/ulid/v2"
)

type RefreshTokenRepo interface {
	Create(ctx context.Context, accountID uint64, tokenID ulid.ULID) error
	Delete(ctx context.Context, tokenID ulid.ULID) (int64, error)
	CountByAccountID(ctx context.Context, accountID uint64) (int64, error)
	DeleteOldestByAccountID(ctx context.Context, accountID uint64) error
}

type AccountRepo interface {
	Create(ctx context.Context, account *domain.Account) (*domain.Account, error)
	FindByID(ctx context.Context, id uint64) (*domain.Account, error)
	FindByPhone(ctx context.Context, phone string) (*domain.Account, error)
	Update(ctx context.Context, account *domain.Account) error
}

type PhoneConfirmationRepo interface {
	Create(ctx context.Context, confirmation *domain.PhoneConfirmation) (*domain.PhoneConfirmation, error)
	FindByID(ctx context.Context, id uint64) (*domain.PhoneConfirmation, error)
	FindByAccountID(ctx context.Context, accountID uint64) (*domain.PhoneConfirmation, error)
	Update(ctx context.Context, confirmation *domain.PhoneConfirmation) error
}

type ArbitrageRepo interface {
	Save(ctx context.Context, arbitrage *domain.Arbitrage) (*domain.Arbitrage, error)
	Create(ctx context.Context, arbitrage *domain.Arbitrage) (*domain.Arbitrage, error)
	FindAllByFilter(ctx context.Context, filter filters.ArbitrageParams) ([]*domain.Arbitrage, error)
	FindByID(ctx context.Context, id uint64) (*domain.Arbitrage, error)
	FindByPair(ctx context.Context, pair string) (*domain.Arbitrage, error)
	Update(ctx context.Context, arbitrage *domain.Arbitrage) (int64, error)
}
