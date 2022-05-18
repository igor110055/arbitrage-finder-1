package handlers

import (
	"calc/cmd/api/http/middlewares"
	"calc/foundation/jwt"
	"calc/foundation/mux"
	"calc/internal/adapters/db"
	"calc/internal/services/auth"
	"calc/internal/services/exchange"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func API(
	build string,
	db db.DB,
	jwtAuth *jwt.Authenticator,
	authService *auth.Service,
	exchangeService *exchange.Service,
) http.Handler {
	r := mux.NewRouter()
	r.HTTPHandle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r *mux.Router) {
		r.Use(middlewares.Logger)

		cg := newCheckGroup(build, db)
		r.Route("/check", func(r *mux.Router) {
			r.Handle("/readiness", cg.Readiness).Methods(http.MethodGet)
			r.Handle("/liveness", cg.Liveness).Methods(http.MethodGet)
		})

		ag := newAuthGroup(authService)
		r.Route("/auth", func(r *mux.Router) {
			r.Handle("/sign-up", ag.SignUp, mux.WithTx(db)).Methods(http.MethodPost)
			r.Handle("/confirm", ag.Confirm, mux.WithTx(db)).Methods(http.MethodPost)
			r.Handle("/code/{phone}", ag.Code, mux.WithTx(db)).Methods(http.MethodPost)
			r.Handle("/sign-in", ag.SignIn).Methods(http.MethodPost)
			r.Handle("/check/{phone}", ag.CheckPhone).Methods(http.MethodGet)

			r.Group(func(r *mux.Router) {
				r.Use(middlewares.Verify(jwtAuth, jwt.Refresh))
				r.Handle("/refresh", ag.Refresh).Methods(http.MethodPost)
			})
		})

		eg := newExchangeGroup(exchangeService)
		r.Route("/exchange", func(r *mux.Router) {
			//r.Use(middlewares.Verify(jwtAuth, jwt.Access))
			r.Handle("", eg.Exchanges).Methods(http.MethodGet)
			r.Handle("/{exchange}/pairs", eg.Pairs).Methods(http.MethodGet)
			r.Handle("/{exchange}/price/{pair}", eg.Price).Methods(http.MethodGet)
			r.Handle("/top", eg.Top).Methods(http.MethodGet)
			r.Route("/ws", func(r *mux.Router) {
				r.WSHandle("/{exchange}/price/{pair}", eg.WSPrice).Methods(http.MethodGet)
			})
		})
	})

	return handlers.CORS(
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut}),
		handlers.AllowedHeaders([]string{
			"Authorization",
			"Content-Type",
		}),
	)(r)
}
