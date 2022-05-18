package postgres

import (
	"calc/internal/adapters/db"
	"calc/internal/adapters/db/filters"
	"calc/internal/domain"
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

const arbitragesTable = "arbitrages"

type Arbitrage struct {
	Pair         string    `db:"pair"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	BuyExchange  string    `db:"buy_exchange"`
	SellExchange string    `db:"sell_exchange"`
	BuyPrice     float64   `db:"buy_price"`
	SellPrice    float64   `db:"sell_price"`
	Profit       float64   `db:"profit"`
}

type ArbitrageRepo struct {
	db *DB
}

func (r *ArbitrageRepo) Save(ctx context.Context, arbitrage *domain.Arbitrage) (*domain.Arbitrage, error) {
	n, err := r.Update(ctx, arbitrage)
	if err != nil {
		log.Error().Stack().Err(err).Msg("failed to exec `Save`")

		return nil, err
	}

	if n == 0 {
		return r.Create(ctx, arbitrage)
	}

	return arbitrage, nil
}

func (r *ArbitrageRepo) Create(ctx context.Context, arbitrage *domain.Arbitrage) (*domain.Arbitrage, error) {
	clauses := map[string]interface{}{
		"pair":          arbitrage.Pair,
		"buy_exchange":  arbitrage.BuyExchange,
		"sell_exchange": arbitrage.SellExchange,
		"buy_price":     arbitrage.BuyPrice,
		"sell_price":    arbitrage.SellPrice,
		"profit":        arbitrage.Profit,
	}

	q, args, err := r.db.Sq.Insert(arbitragesTable).SetMap(clauses).ToSql()
	if err != nil {
		return nil, err
	}

	if _, err := r.db.ExecContext(ctx, q, args, true); err != nil {
		var pgError *pgconn.PgError

		if errors.As(err, &pgError) {
			if pgError.Code == DuplicateKeyValueCode {
				return nil, db.ErrAlreadyExists
			}

			return nil, pgError
		}

		return nil, errors.Wrap(err, "failed to exec query `Create`")
	}

	return arbitrage, nil
}

func (r *ArbitrageRepo) FindAllByFilter(ctx context.Context, filter filters.ArbitrageParams) ([]*domain.Arbitrage, error) {
	sb := r.db.Sq.Select("*").From(arbitragesTable).Limit(uint64(filter.Limit))

	sb = sb.OrderBy(fmt.Sprintf("%s %s", filter.SortBy, filter.SortDir))

	q, args, err := sb.ToSql()
	if err != nil {
		log.Error().Stack().Err(err).Msg("failed to build query `FindAllByFilter`")
		return nil, errors.Wrap(err, "failed to build query `FindAllByFilter`")
	}

	var dbArbitrage []Arbitrage
	err = r.db.SelectContext(ctx, q, &dbArbitrage, args)
	if err != nil {
		log.Error().Stack().Err(err).Msg("failed to build query `FindAllByFilter`")
		return nil, errors.Wrap(err, "failed to exec query `FindAllByFilter`")
	}

	var arbitrages []*domain.Arbitrage
	for _, arbitrage := range dbArbitrage {
		arbitrages = append(arbitrages, &domain.Arbitrage{
			Pair:         arbitrage.Pair,
			BuyExchange:  arbitrage.BuyExchange,
			SellExchange: arbitrage.SellExchange,
			BuyPrice:     arbitrage.BuyPrice,
			SellPrice:    arbitrage.SellPrice,
			Profit:       arbitrage.Profit,
		})
	}

	return arbitrages, nil
}

func (r *ArbitrageRepo) FindByID(ctx context.Context, id uint64) (*domain.Arbitrage, error) {
	q, args, err := r.db.Sq.Select("*").From(arbitragesTable).Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query `FindByID`")
	}

	var dbArbitrage Arbitrage

	err = r.db.GetContext(ctx, &dbArbitrage, q, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec query `FindByID`")
	}

	return &domain.Arbitrage{
		Pair:         dbArbitrage.Pair,
		BuyExchange:  dbArbitrage.BuyExchange,
		SellExchange: dbArbitrage.SellExchange,
		BuyPrice:     dbArbitrage.BuyPrice,
		SellPrice:    dbArbitrage.SellPrice,
		Profit:       dbArbitrage.Profit,
	}, nil
}

func (r *ArbitrageRepo) FindByPair(ctx context.Context, pair string) (*domain.Arbitrage, error) {
	q, args, err := r.db.Sq.Select("*").From(arbitragesTable).Where(squirrel.Eq{"pair": pair}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query `FindByPair`")
	}

	var dbArbitrage Arbitrage

	err = r.db.GetContext(ctx, &dbArbitrage, q, args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec query `FindByPair`")
	}

	return &domain.Arbitrage{
		Pair:         dbArbitrage.Pair,
		BuyExchange:  dbArbitrage.BuyExchange,
		SellExchange: dbArbitrage.SellExchange,
		BuyPrice:     dbArbitrage.BuyPrice,
		SellPrice:    dbArbitrage.SellPrice,
		Profit:       dbArbitrage.Profit,
	}, nil
}

func (r *ArbitrageRepo) Update(ctx context.Context, arbitrage *domain.Arbitrage) (int64, error) {
	clauses := map[string]interface{}{
		"buy_exchange":  arbitrage.BuyExchange,
		"sell_exchange": arbitrage.SellExchange,
		"buy_price":     arbitrage.BuyPrice,
		"sell_price":    arbitrage.SellPrice,
		"profit":        arbitrage.Profit,
	}

	q, args, err := r.db.Sq.Update(arbitragesTable).SetMap(clauses).Where(squirrel.Eq{"pair": arbitrage.Pair}).ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "error build query `Update`")
	}

	result, err := r.db.ExecContext(ctx, q, args, true)
	if err != nil {
		return 0, errors.Wrap(err, "failed to exec query `Update`")
	}

	return result.RowsAffected()
}
