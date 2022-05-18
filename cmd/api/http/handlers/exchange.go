package handlers

import (
	"calc/cmd/api/http/handlers/requests"
	"calc/cmd/api/http/handlers/responses"
	"calc/internal/berrors"
	"calc/internal/services/auth"
	"calc/internal/services/exchange"
	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type exchangeGroup struct {
	exchangeService *exchange.Service
}

func newExchangeGroup(exchangeService *exchange.Service) *exchangeGroup {
	return &exchangeGroup{
		exchangeService: exchangeService,
	}
}

// Exchanges godoc
// @Tags Exchange
// @Router /exchange [get]
// @Summary returns available exchanges
// @Produce json
// @Success 200 {array} string
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (eg *exchangeGroup) Exchanges(r *http.Request) (interface{}, error) {
	return eg.exchangeService.Exchanges(r.Context()), nil
}

// Pairs godoc
// @Tags Exchange
// @Router /exchange/{exchange}/pairs [get]
// @Summary returns available pairs
// @Produce json
// @Param exchange exchange string true "Exchange"
// @Success 200 {array} string
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (eg *exchangeGroup) Pairs(r *http.Request) (interface{}, error) {
	return eg.exchangeService.Pairs(r.Context(), mux.Vars(r)["exchange"])
}

// Price godoc
// @Tags Exchange
// @Router /exchange/{exchange}/price/{pair} [get]
// @Summary returns price of pair
// @Produce json
// @Param exchange exchange string true "Exchange"
// @Param pair pair string true "Pair"
// @Success 200 {object} float64
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (eg *exchangeGroup) Price(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	return eg.exchangeService.Price(r.Context(), vars["exchange"], vars["pair"])
}

// Top godoc
// @Tags Exchange
// @Router /exchange/top [get]
// @Summary returns the top most profitable pairs for arbitrage
// @Produce json
// @Success 200 {object} responses.Top
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (eg *exchangeGroup) Top(r *http.Request) (interface{}, error) {
	var req requests.Top
	if err := requests.Bind(r, &req); err != nil {
		return nil, berrors.WrapWithError(auth.ErrInvalidInput, err)
	}

	top, err := eg.exchangeService.Top(r.Context(), req.Limit)
	if err != nil {
		return nil, err
	}

	resp := make([]*responses.Top, 0)
	for _, t := range top {
		resp = append(resp, &responses.Top{
			Pair:         t.Pair,
			BuyExchange:  t.BuyExchange,
			SellExchange: t.SellExchange,
			BuyPrice:     t.BuyPrice,
			SellPrice:    t.SellPrice,
			Profit:       t.Profit,
		})
	}

	return resp, nil
}

// WSPrice godoc
// @Tags Exchange
// @Router /exchange/ws/{exchange}/price/{pair} [get]
// @Summary pair price subscription
// @Produce json
// @Param exchange exchange string true "Exchange"
// @Param pair pair string true "Pair"
// @Success 200 {object} адщфе64
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (eg *exchangeGroup) WSPrice(ctx context.Context, c *websocket.Conn, vars map[string]string) error {
	ch := make(chan float64)
	errCh := make(chan error)

	pair, exch := vars["pair"], vars["exchange"]
	go func() {
		err := eg.exchangeService.WSPrice(ctx, exch, pair, ch)
		if err != nil {
			errCh <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			return err
		case price := <-ch:
			err := c.WriteJSON(struct {
				Pair     string    `json:"pair"`
				Exchange string    `json:"exchange"`
				Price    float64   `json:"price"`
				Time     time.Time `json:"time"`
			}{
				Pair:     pair,
				Exchange: exch,
				Price:    price,
				Time:     time.Now(),
			})
			if err != nil {
				return err
			}
		}
	}
}
