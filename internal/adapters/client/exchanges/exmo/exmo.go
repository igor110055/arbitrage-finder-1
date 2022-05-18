package exmo

import (
	"calc/common/config"
	"calc/foundation/id"
	"calc/internal/adapters/client"
	"calc/internal/adapters/client/exchanges/exmo/response"
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
)

const (
	tickersUri      = "/required_amount"
	pairSettingsUri = "/pair_settings"
)

var (
	promPrices = map[string]prometheus.Gauge{}

	errConsumingRateIsTooSlow = errors.New("message consuming rate is too slow")
)

type Exmo struct {
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

func NewExmo(ctx context.Context, cfg *config.ExchangeConfig, calculator calculator.CalculateService) *Exmo {
	httpClient := client.NewHTTPClient()

	exmoLogger := log.Logger.With().Str("logger", "exmo").Logger()

	exmoCtx, cancel := context.WithCancel(ctx)

	exmo := &Exmo{
		ctx:        exmoCtx,
		cancel:     cancel,
		url:        cfg.URL,
		wsURL:      cfg.WsURL,
		logger:     &exmoLogger,
		httpClient: httpClient,
		chans:      make(map[string]map[string]chan<- *domain.Data),
		calculator: calculator,
		prices:     make(map[string]float64),
		pairs:      cfg.Pairs,
	}

	go func() {
		for {
			if err := exmo.run(); err != nil {
				exmo.cancel()

				if err == errConsumingRateIsTooSlow {
					continue
				}

				return
			}

			break
		}
	}()

	return exmo
}

func (e *Exmo) run() error {
	logger := e.logger.With().Str("method", "run").Logger()

	logger.Info().Msg(strings.Join(e.pairs, ","))

	c, _, err := websocket.DefaultDialer.Dial(e.wsURL, nil)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to connect")
		return err
	}
	defer c.Close()

	topics := make([]string, len(e.pairs))
	for i, pair := range e.pairs {
		promPrices[pair] = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "calc",
			Name:        "exmo_price",
			Help:        "pair price",
			ConstLabels: prometheus.Labels{"pair": pair},
		})
		prometheus.MustRegister(promPrices[pair])
		e.chans[pair] = make(map[string]chan<- *domain.Data)
		topics[i] = fmt.Sprintf("spot/ticker:%s", pair)
	}

	init := struct {
		Id     int      `json:"id"`
		Method string   `json:"method"`
		Topics []string `json:"topics"`
	}{
		Id:     1,
		Method: "subscribe",
		Topics: topics,
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

			if ticker.Event == "error" {
				return errors.New(ticker.Message)
			} else if ticker.Event != "update" {
				continue
			}

			pair := strings.Split(ticker.Topic, ":")[1]
			if e.prices[pair] == ticker.Data.BuyPrice {
				continue
			}
			e.prices[pair] = ticker.Data.BuyPrice

			data := &domain.Data{
				Exchange: "exmo",
				Pair:     pair,
				Price:    ticker.Data.BuyPrice,
			}
			if err := e.calculator.Save(data); err != nil {
				logger.Error().Stack().Err(err).Msgf("failed to put data on calculator")
				return err
			}

			for _, ch := range e.chans[pair] {
				ch <- data
			}

			promPrices[pair].Set(data.Price)
		}
	}
}

func (e *Exmo) Pairs(ctx context.Context) ([]string, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, pairSettingsUri))
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

	var settingsResponse response.PairSettingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&settingsResponse); err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to decode ticker response")
		return nil, err
	}

	var pairs []string
	for pair := range settingsResponse {
		pairs = append(pairs, pair)
	}

	return pairs, nil
}

func (e *Exmo) Price(ctx context.Context, pair string) (float64, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, tickersUri))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to parse url")
		return 0, err
	}

	q := u.Query()
	q.Add("pair", pair)
	q.Add("quantity", "1")

	u.RawQuery = q.Encode()

	resp, err := e.httpClient.Get(ctx, u.String())
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to request ticker")
		return 0, err
	}

	defer resp.Body.Close()

	var amount response.RequiredAmount
	if err := json.NewDecoder(resp.Body).Decode(&amount); err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to decode ticker response")
		return 0, err
	}

	return amount.Amount, nil
}

func (e *Exmo) WSPrice(ctx context.Context, pair string, ch chan<- *domain.Data) {
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
