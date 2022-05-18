package response

type PairSettingsResponse map[string]*PairSetting

type PairSetting struct {
	MinQuantity            string `json:"min_quantity"`
	MaxQuantity            string `json:"max_quantity"`
	MinPrice               string `json:"min_price"`
	MaxPrice               string `json:"max_price"`
	MaxAmount              string `json:"max_amount"`
	MinAmount              string `json:"min_amount"`
	PricePrecision         int    `json:"price_precision"`
	CommissionTakerPercent string `json:"commission_taker_percent"`
	CommissionMakerPercent string `json:"commission_maker_percent"`
}
