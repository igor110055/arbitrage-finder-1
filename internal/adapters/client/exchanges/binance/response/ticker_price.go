package response

type TickerPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}
