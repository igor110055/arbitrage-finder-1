package requests

import "net/http"

type SignIn struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (e *SignIn) Bind(*http.Request) error {
	return nil
}
