package filters

type ArbitrageSortBy string

const (
	ArbitrageSortByProfit ArbitrageSortBy = "profit"
)

type ArbitrageParams struct {
	Limit   uint
	SortBy  ArbitrageSortBy
	SortDir SortDirection
}
