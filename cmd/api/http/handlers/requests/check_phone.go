package requests

import (
	"github.com/gorilla/mux"
	"net/http"
)

type CheckPhone struct {
	Phone string `json:"phone" validate:"required"`
}

func (r *CheckPhone) Bind(req *http.Request) error {
	r.Phone = mux.Vars(req)["phone"]

	return nil
}
