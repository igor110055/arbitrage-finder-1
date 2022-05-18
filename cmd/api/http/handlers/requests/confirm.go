package requests

import "net/http"

type Confirm struct {
	ConfirmationID uint64 `json:"confirmation_id" validate:"required"`
	Password       string `json:"password" validate:"required"`
	Code           string `json:"code" validate:"required"`
}

func (r *Confirm) Bind(_ *http.Request) error {
	return nil
}
