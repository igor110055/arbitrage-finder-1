package exchanges

import (
	"calc/common/config"
	"calc/internal/adapters/client/exchanges/binance"
	"calc/internal/adapters/client/exchanges/exmo"
	"calc/internal/adapters/client/exchanges/gate"
	"calc/internal/services/calculator"
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ExchangeFactory struct {
	cfg       *config.Config
	logger    *zerolog.Logger
	exchanges map[string]Exchange
}

func NewExchangeFactory(ctx context.Context, cfg *config.Config, calculateService calculator.CalculateService) *ExchangeFactory {
	factoryLogger := log.With().Str("logger", "exchange_factory").Logger()

	exchanges := make(map[string]Exchange)
	for exchange, exchangeCfg := range cfg.Exchanges.Configs {
		var exch Exchange
		switch exchange {
		case "exmo":
			exch = exmo.NewExmo(ctx, exchangeCfg, calculateService)
		case "binance":
			exch = binance.NewBinance(ctx, exchangeCfg, calculateService)
		case "gate":
			exch = gate.NewGate(ctx, exchangeCfg, calculateService)
		}

		exchanges[exchange] = exch
	}

	return &ExchangeFactory{
		cfg:       cfg,
		logger:    &factoryLogger,
		exchanges: exchanges,
	}
}

func (f *ExchangeFactory) List() []string {
	var exchanges []string
	for exch := range f.exchanges {
		exchanges = append(exchanges, exch)
	}

	return exchanges
}

func (f *ExchangeFactory) Get(exchange string) (Exchange, error) {
	if exch, ok := f.exchanges[exchange]; ok {
		return exch, nil
	}

	f.logger.Error().Stack().Err(ErrExchangeNotFound).Msgf("%s exchange not found", exchange)
	return nil, ErrExchangeNotFound
}
