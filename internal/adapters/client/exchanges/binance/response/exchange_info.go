package response

type ExchangeInfo struct {
	Timezone        string        `json:"timezone"`
	ServerTime      int64         `json:"serverTime"`
	ExchangeFilters []interface{} `json:"exchangeFilters"`
	Symbols         []*Symbol     `json:"symbols"`
}

type Symbol struct {
	Symbol                     string        `json:"symbol"`
	Status                     string        `json:"status"`
	BaseAsset                  string        `json:"baseAsset"`
	BaseAssetPrecision         int           `json:"baseAssetPrecision"`
	QuoteAsset                 string        `json:"quoteAsset"`
	QuotePrecision             int           `json:"quotePrecision"`
	QuoteAssetPrecision        int           `json:"quoteAssetPrecision"`
	OrderTypes                 []string      `json:"orderTypes"`
	IcebergAllowed             bool          `json:"icebergAllowed"`
	OcoAllowed                 bool          `json:"ocoAllowed"`
	QuoteOrderQtyMarketAllowed bool          `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop          bool          `json:"allowTrailingStop"`
	IsSpotTradingAllowed       bool          `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed     bool          `json:"isMarginTradingAllowed"`
	Filters                    []interface{} `json:"filters"`
	Permissions                []string      `json:"permissions"`
}

func (s *Symbol) HasPermission(perm string) bool {
	if s == nil {
		return false
	}

	for _, permission := range s.Permissions {
		if permission == perm {
			return true
		}
	}

	return false
}
