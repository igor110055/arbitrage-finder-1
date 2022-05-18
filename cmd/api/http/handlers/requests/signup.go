package requests

import "net/http"

type SignUp struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

func (r *SignUp) Bind(_ *http.Request) error {
	return nil
}
