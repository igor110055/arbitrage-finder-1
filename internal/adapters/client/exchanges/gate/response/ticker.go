package response

type Ticker struct {
	CurrencyPair     string  `json:"currency_pair"`
	Last             float64 `json:"last,string"`
	LowestAsk        float64 `json:"lowest_ask,string"`
	HighestBid       float64 `json:"highest_bid,string"`
	ChangePercentage float64 `json:"change_percentage,string"`
	BaseVolume       float64 `json:"base_volume,string"`
	QuoteVolume      float64 `json:"quote_volume,string"`
	High24H          float64 `json:"high_24h,string"`
	Low24H           float64 `json:"low_24h,string"`
	EtfNetValue      float64 `json:"etf_net_value,string"`
	EtfPreNetValue   float64 `json:"etf_pre_net_value,string"`
	EtfPreTimestamp  int     `json:"etf_pre_timestamp"`
	EtfLeverage      float64 `json:"etf_leverage,string"`
}
