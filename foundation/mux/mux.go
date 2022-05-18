package mux

import (
	"calc/internal/berrors"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

var upgrader = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler func(r *http.Request) (interface{}, error)

type WSHandler func(ctx context.Context, c *websocket.Conn, vars map[string]string) error

type Router struct {
	*mux.Router
}

func NewRouter() *Router {
	return &Router{
		mux.NewRouter(),
	}
}

func (r *Router) SubRouter() *Router {
	return &Router{
		r.NewRoute().Subrouter(),
	}
}

func (r *Router) PathPrefix(path string) *Router {
	return &Router{
		r.Router.PathPrefix(path).Subrouter(),
	}
}

func (r *Router) Group(fn func(r *Router)) {
	router := r.SubRouter()

	if fn != nil {
		fn(router)
	}
}

func (r *Router) Route(path string, fn func(r *Router)) {
	router := r.PathPrefix(path)

	if fn != nil {
		fn(router)
	}
}

func (r *Router) HTTPHandle(path string, handler http.Handler, options ...HandleOption) *mux.Route {
	return r.Router.Handle(path, handler)
}

func (r *Router) Handle(path string, handler Handler, options ...HandleOption) *mux.Route {
	return r.Router.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		opts, err := NewHandleOptions(ctx, options...)
		if err != nil {
			respondError(w, r, err)
			return
		}

		defer opts.rollback()

		resp, err := handler(r.WithContext(opts.ctx))
		if err != nil {
			bError := &berrors.BusinessError{}
			if errors.As(err, &bError) && bError.NeedCommit {
				if err := opts.commit(); err != nil {
					respondError(w, r, err)
					return
				}
			}
			respondError(w, r, err)
			return
		}

		if err := respond(w, resp, http.StatusOK); err != nil {
			respondError(w, r, err)
			return
		}

		if err := opts.commit(); err != nil {
			respondError(w, r, err)
			return
		}
	}))
}

func (r *Router) WSHandle(path string, handler WSHandler, options ...HandleOption) *mux.Route {
	return r.Router.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error on connect ", err)
			return
		}

		defer func() {
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			c.Close()
		}()

		ctx, cancel := context.WithCancel(r.Context())
		go func() {
			_, _, err := c.ReadMessage()
			if err != nil {
				cancel()
			}
		}()

		vars := mux.Vars(r)
		if err := handler(ctx, c, vars); err != nil {
			wsRespondError(r, c, err)
			return
		}

		log.Info().
			Str("ip", r.RemoteAddr).
			Str("uri", r.RequestURI).
			Msg("disconnected")
	}))
}

func respond(w http.ResponseWriter, data interface{}, statusCode int) error {
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	_, err = w.Write(jsonData)
	return err
}

func respondError(w http.ResponseWriter, r *http.Request, err error) {
	bError := &berrors.BusinessError{}
	if errors.As(err, &bError) {
		log.Info().
			Str("ip", r.RemoteAddr).
			Str("uri", r.RequestURI).
			Err(bError).
			Int("errorCode", bError.ErrCode).
			Msg(err.Error())

		if err := respond(w, &berrors.BusinessError{
			ErrCode: bError.ErrCode,
			Message: err.Error(),
		}, http.StatusBadRequest); err != nil {
			panic(err)
		}

		return
	}

	log.Error().
		Str("ip", r.RemoteAddr).
		Str("uri", r.RequestURI).
		Stack().
		Err(err).
		Msg("Internal server error")

	w.WriteHeader(http.StatusInternalServerError)
}

func wsRespondError(r *http.Request, c *websocket.Conn, err error) {
	bError := &berrors.BusinessError{}
	if errors.As(err, &bError) {
		log.Info().
			Str("ip", r.RemoteAddr).
			Str("uri", r.RequestURI).
			Err(bError).
			Int("errorCode", bError.ErrCode).
			Msg(err.Error())

		if err := c.WriteJSON(&berrors.BusinessError{
			ErrCode: bError.ErrCode,
			Message: err.Error(),
		}); err != nil {
			panic(err)
		}

		return
	}

	log.Error().
		Str("ip", r.RemoteAddr).
		Str("uri", r.RequestURI).
		Stack().
		Err(err).
		Msg("Internal server error")

	if err := c.WriteMessage(websocket.TextMessage, []byte("internal server error")); err != nil {
		panic(err)
	}
}
