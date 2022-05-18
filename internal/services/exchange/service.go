package exchange

import (
	"calc/common/config"
	"calc/internal/adapters/client/exchanges"
	"calc/internal/adapters/db"
	"calc/internal/adapters/db/filters"
	"calc/internal/domain"
	"calc/internal/services/calculator"
	"context"
)

type Service struct {
	arbitrageRepo    db.ArbitrageRepo
	exchangeFactory  *exchanges.ExchangeFactory
	calculateService calculator.CalculateService
}

func NewService(ctx context.Context, cfg *config.Config, arbitrageRepo db.ArbitrageRepo) *Service {
	calculateService := calculator.NewCalculateService(ctx, cfg, arbitrageRepo)

	return &Service{
		exchangeFactory:  exchanges.NewExchangeFactory(ctx, cfg, calculateService),
		calculateService: calculateService,
		arbitrageRepo:    arbitrageRepo,
	}
}

type SignUpArgs struct {
	Phone    string
	Password string
}

func (s *Service) Exchanges(ctx context.Context) []string {
	return s.exchangeFactory.List()
}

func (s *Service) Pairs(ctx context.Context, exchange string) ([]string, error) {
	e, err := s.exchangeFactory.Get(exchange)
	if err != nil {
		return nil, err
	}

	return e.Pairs(ctx)
}

func (s *Service) Price(ctx context.Context, exchange string, pair string) (float64, error) {
	e, err := s.exchangeFactory.Get(exchange)
	if err != nil {
		return 0, err
	}

	return e.Price(ctx, pair)
}

func (s *Service) Top(ctx context.Context, limit uint) ([]*domain.Arbitrage, error) {
	return s.arbitrageRepo.FindAllByFilter(ctx, filters.ArbitrageParams{
		Limit:   limit,
		SortBy:  filters.ArbitrageSortByProfit,
		SortDir: filters.Desc,
	})
}

func (s *Service) WSPrice(ctx context.Context, exchange string, pair string, ch chan<- float64) error {
	e, err := s.exchangeFactory.Get(exchange)
	if err != nil {
		return err
	}

	dataCh := make(chan *domain.Data)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-dataCh:
				ch <- data.Price
			}
		}
	}()

	e.WSPrice(ctx, pair, dataCh)
	return nil
}
