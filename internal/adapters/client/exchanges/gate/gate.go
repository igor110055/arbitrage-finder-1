package gate

import (
	"calc/common/config"
	"calc/foundation/id"
	"calc/internal/adapters/client"
	"calc/internal/adapters/client/exchanges/gate/response"
	"calc/internal/domain"
	"calc/internal/services/calculator"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/url"
	"strings"
	"time"
)

const (
	pairsUri  = "/spot/currency_pairs"
	tickerUri = "/spot/tickers"
)

var (
	promPrices = map[string]prometheus.Gauge{}

	errNotFound = errors.New("not found")
)

type Gate struct {
	ctx        context.Context
	cancel     func()
	url        string
	wsURL      string
	logger     *zerolog.Logger
	httpClient client.HTTPClient
	calculator calculator.CalculateService
	chans      map[string]map[string]chan<- *domain.Data
	prices     map[string]float64
	pairs      []string
}

func NewGate(ctx context.Context, cfg *config.ExchangeConfig, calculator calculator.CalculateService) *Gate {
	httpClient := client.NewHTTPClient()

	gateLogger := log.Logger.With().Str("logger", "gate").Logger()

	gateCtx, cancel := context.WithCancel(ctx)

	gate := &Gate{
		ctx:        gateCtx,
		cancel:     cancel,
		url:        cfg.URL,
		wsURL:      cfg.WsURL,
		logger:     &gateLogger,
		httpClient: httpClient,
		chans:      make(map[string]map[string]chan<- *domain.Data),
		calculator: calculator,
		prices:     make(map[string]float64),
		pairs:      cfg.Pairs,
	}

	go func() {
		for {
			if err := gate.run(); err != nil {
				gate.cancel()
				return
			}

			break
		}
	}()

	return gate
}

func (e *Gate) run() error {
	logger := e.logger.With().Str("method", "run").Logger()

	logger.Info().Msg(strings.Join(e.pairs, ","))

	c, _, err := websocket.DefaultDialer.Dial(e.wsURL, nil)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to connect")
		return err
	}
	defer c.Close()

	pairs := make([]string, len(e.pairs))
	for i, pair := range e.pairs {
		promPrices[pair] = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "calc",
			Name:        "gate_price",
			Help:        "pair price",
			ConstLabels: prometheus.Labels{"pair": pair},
		})
		prometheus.MustRegister(promPrices[pair])
		e.chans[pair] = make(map[string]chan<- *domain.Data)
		pairs[i] = pair
	}

	init := struct {
		Time    int64    `json:"time"`
		Channel string   `json:"channel"`
		Event   string   `json:"event"`
		Payload []string `json:"payload"`
	}{
		Time:    time.Now().Unix(),
		Channel: "spot.tickers",
		Event:   "subscribe",
		Payload: pairs,
	}

	err = c.WriteJSON(init)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to write init message")
		return err
	}
	logger.Debug().Msgf("init message %v successful sended", init)

	for {
		select {
		case <-e.ctx.Done():
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logger.Error().Stack().Err(err).Msgf("failed to close connection")
				return err
			}

			logger.Info().Msgf("connection closed %v", e.ctx.Err())
			return nil
		default:
			var ticker *response.WSTicker
			if err := c.ReadJSON(&ticker); err != nil {
				logger.Error().Stack().Err(err).Msgf("failed to read message")
				return err
			}

			if ticker.Event != "update" {
				continue
			}

			if e.prices[ticker.Result.CurrencyPair] == ticker.Result.Last {
				continue
			}
			e.prices[ticker.Result.CurrencyPair] = ticker.Result.Last

			data := &domain.Data{
				Exchange: "gate",
				Pair:     ticker.Result.CurrencyPair,
				Price:    ticker.Result.Last,
			}
			if err := e.calculator.Save(data); err != nil {
				logger.Error().Stack().Err(err).Msgf("failed to put data on calculator")
				return err
			}

			for _, ch := range e.chans[ticker.Result.CurrencyPair] {
				ch <- data
			}

			promPrices[ticker.Result.CurrencyPair].Set(data.Price)
		}
	}
}

func (e *Gate) Pairs(ctx context.Context) ([]string, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, pairsUri))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to parse url")
		return nil, err
	}

	resp, err := e.httpClient.Get(ctx, u.String())
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to request ticker")
		return nil, err
	}

	defer resp.Body.Close()

	var pairsResponse response.PairsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pairsResponse); err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to decode ticker response")
		return nil, err
	}

	var pairs []string
	for _, pair := range pairsResponse {
		if pair.TradeStatus == "tradable" {
			pairs = append(pairs, pair.Id)
		}
	}

	return pairs, nil
}

func (e *Gate) Price(ctx context.Context, pair string) (float64, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, tickerUri))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to parse url")
		return 0, err
	}

	q := u.Query()
	q.Add("currency_pair", pair)

	u.RawQuery = q.Encode()

	resp, err := e.httpClient.Get(ctx, u.String())
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to request ticker")
		return 0, err
	}

	defer resp.Body.Close()

	var tickers []*response.Ticker
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to decode ticker response")
		return 0, err
	}

	if len(tickers) == 0 {
		return 0, errNotFound
	}

	return tickers[0].Last, nil
}

func (e *Gate) WSPrice(ctx context.Context, pair string, ch chan<- *domain.Data) {
	chanID := id.ULID().String()
	e.chans[pair][chanID] = ch

	defer delete(e.chans[pair], chanID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.ctx.Done():
			return
		}
	}
}
