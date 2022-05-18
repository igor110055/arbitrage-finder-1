package responses

type Top struct {
	Pair         string  `json:"pair"`
	BuyExchange  string  `json:"buy_exchange"`
	SellExchange string  `json:"sell_exchange"`
	BuyPrice     float64 `json:"buy_price"`
	SellPrice    float64 `json:"sell_price"`
	Profit       float64 `json:"profit"`
}
