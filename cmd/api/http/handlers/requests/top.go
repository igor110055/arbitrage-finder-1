package requests

import (
	"net/http"
	"strconv"
)

type Top struct {
	Limit uint `json:"limit"`
}

func (e *Top) Bind(req *http.Request) error {
	q := req.URL.Query()

	e.Limit = 20

	limitString := q.Get("limit")
	if limitString != "" {
		limit, err := strconv.ParseUint(limitString, 10, 32)
		if err != nil {
			return err
		}

		e.Limit = uint(limit)
	}

	return nil
}
