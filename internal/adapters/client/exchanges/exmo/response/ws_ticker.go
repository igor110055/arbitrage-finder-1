package response

type WSTicker struct {
	Ts      int64  `json:"ts"`
	Event   string `json:"event"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
	Data    struct {
		BuyPrice  float64 `json:"buy_price,string"`
		SellPrice float64 `json:"sell_price,string"`
		LastTrade float64 `json:"last_trade,string"`
		High      float64 `json:"high,string"`
		Low       float64 `json:"low,string"`
		Avg       float64 `json:"avg,string"`
		Vol       float64 `json:"vol,string"`
		VolCurr   float64 `json:"vol_curr,string"`
		Updated   int     `json:"updated"`
	} `json:"data"`
}
