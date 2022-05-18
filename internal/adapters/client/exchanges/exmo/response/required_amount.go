package response

type RequiredAmount struct {
	Quantity int     `json:"quantity,string"`
	Amount   float64 `json:"amount,string"`
	AvgPrice float64 `json:"avg_price,string"`
}
