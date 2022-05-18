package requests

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

type Binder interface {
	Bind(r *http.Request) error
}

func Bind(r *http.Request, v Binder) error {
	if err := Decode(r.Body, v); err != nil && err != io.EOF {
		return err
	}

	if err := v.Bind(r); err != nil {
		return err
	}

	return ValidateStruct(v)
}

func Decode(r io.Reader, v interface{}) error {
	defer io.Copy(ioutil.Discard, r)
	return json.NewDecoder(r).Decode(v)
}
