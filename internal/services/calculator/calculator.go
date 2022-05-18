package calculator

import (
	"calc/internal/domain"
	"math"
)

type calculator struct {
	pair      string
	arbitrage domain.Arbitrage
}

func NewCalculator(pair string) *calculator {
	return &calculator{
		pair: pair,
		arbitrage: domain.Arbitrage{
			Pair:      pair,
			SellPrice: -1,
			BuyPrice:  math.MaxFloat64,
		},
	}
}

func (c *calculator) Put(data *domain.Data) *domain.Arbitrage {
	if c.pair != data.Pair {
		return nil
	}

	if (c.arbitrage.SellExchange == data.Exchange) || (c.arbitrage.SellPrice < data.Price && c.arbitrage.BuyExchange != data.Exchange) {
		c.calcSell(data)
		return &c.arbitrage
	}

	if (c.arbitrage.BuyExchange == data.Exchange) || (c.arbitrage.BuyPrice > data.Price && c.arbitrage.SellExchange != data.Exchange) {
		c.calcBuy(data)
		return &c.arbitrage
	}

	return nil
}

func (c *calculator) calcBuy(data *domain.Data) {
	c.arbitrage.BuyPrice = data.Price
	c.arbitrage.BuyExchange = data.Exchange
	c.calcProfit()
}

func (c *calculator) calcSell(data *domain.Data) {
	c.arbitrage.SellPrice = data.Price
	c.arbitrage.SellExchange = data.Exchange
	c.calcProfit()
}

func (c *calculator) calcProfit() {
	c.arbitrage.Profit = (c.arbitrage.SellPrice - c.arbitrage.BuyPrice) / c.arbitrage.SellPrice * 100
}
