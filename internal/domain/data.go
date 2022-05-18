package domain

import "github.com/pkg/errors"

var (
	ErrNotEqualPairs = errors.New("not equal pairs")
)

type Data struct {
	Exchange string
	Pair     string
	Price    float64
}

type Arbitrage struct {
	Pair         string
	BuyExchange  string
	SellExchange string
	BuyPrice     float64
	SellPrice    float64
	Profit       float64
}
