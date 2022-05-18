package response

type WSTicker struct {
	ID          int     `json:"u"`
	Symbol      string  `json:"s"`
	Bid         float64 `json:"b,string"`
	BidQuantity float64 `json:"B,string"`
	Ask         float64 `json:"a,string"`
	AskQuantity float64 `json:"A,string"`
	Result      *struct {
		ErrorMessage string `json:"msg"`
	} `json:"result"`
}
