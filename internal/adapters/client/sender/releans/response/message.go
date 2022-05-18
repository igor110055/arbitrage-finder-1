package response

type Message struct {
	ApiVersion        string      `json:"api_version"`
	Message           string      `json:"message"`
	Id                string      `json:"id"`
	From              string      `json:"from"`
	To                string      `json:"to"`
	Region            string      `json:"region"`
	Operator          string      `json:"operator"`
	DateCreated       string      `json:"date_created"`
	DateSent          string      `json:"date_sent"`
	DlrStatus         string      `json:"dlr_status"`
	StatusDescription interface{} `json:"status_description"`
	Timezone          string      `json:"timezone"`
	Price             float64     `json:"price"`
	PriceUnit         string      `json:"price_unit"`
	Code              int         `json:"code"`
	Status            int         `json:"status"`
}
