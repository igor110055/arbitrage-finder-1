package middlewares

import (
	"calc/internal/berrors"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

func respond(w http.ResponseWriter, data interface{}, statusCode int) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	_, err = w.Write(jsonData)
	return err
}

func respondError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	bError := &berrors.BusinessError{}
	if errors.As(err, &bError) {
		log.Info().
			Str("ip", r.RemoteAddr).
			Str("uri", r.RequestURI).
			Err(bError).
			Int("errorCode", bError.ErrCode).
			Msg(err.Error())

		if err := respond(w, struct {
			Code int    `json:"code"`
			Msg  string `json:"message,omitempty"`
		}{
			Code: bError.Code(),
			Msg:  err.Error(),
		}, statusCode); err != nil {
			panic(err)
		}

		return
	}

	log.Error().
		Stack().
		Err(err).
		Str("uri", r.RequestURI).
		Msg("Internal server error")

	w.WriteHeader(statusCode)
}
