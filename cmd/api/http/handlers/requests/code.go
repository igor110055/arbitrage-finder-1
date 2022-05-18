package requests

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Code struct {
	Phone string `json:"phone" validate:"required"`
}

func (r *Code) Bind(req *http.Request) error {
	r.Phone = mux.Vars(req)["phone"]

	return nil
}
