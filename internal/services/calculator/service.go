package calculator

import (
	"calc/common/config"
	"calc/internal/adapters/db"
	"calc/internal/domain"
	"context"
)

type CalculateService interface {
	Save(data *domain.Data) error
}

type calculateService struct {
	ctx           context.Context
	arbitrageRepo db.ArbitrageRepo
	pairs         map[string]*calculator
}

func NewCalculateService(ctx context.Context, cfg *config.Config, arbitrageRepo db.ArbitrageRepo) CalculateService {
	pairs := make(map[string]*calculator)
	for _, pair := range cfg.Exchanges.Pairs {
		pairs[pair] = NewCalculator(pair)
	}

	return &calculateService{
		ctx:           ctx,
		arbitrageRepo: arbitrageRepo,
		pairs:         pairs,
	}
}

func (s *calculateService) Save(data *domain.Data) error {
	arbitrage := s.pairs[data.Pair].Put(data)
	if arbitrage != nil && arbitrage.SellExchange != "" && arbitrage.BuyExchange != "" {
		if _, err := s.arbitrageRepo.Save(s.ctx, arbitrage); err != nil {
			return err
		}
	}

	return nil
}
