package binance

import (
	"calc/common/config"
	"calc/foundation/id"
	"calc/internal/adapters/client"
	"calc/internal/adapters/client/exchanges/binance/response"
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
	exchangeInfoUri = "/exchangeInfo"
	tickerPriceUri  = "/ticker/price"
	chunksCount     = 3
)

var (
	promPrices = map[string]prometheus.Gauge{}

	errConsumingRateIsTooSlow = errors.New("message consuming rate is too slow")
)

type Binance struct {
	ctx        context.Context
	cancel     func()
	url        string
	wsURL      string
	logger     *zerolog.Logger
	httpClient client.HTTPClient
	calculator calculator.CalculateService
	chans      map[string]map[string]chan<- *domain.Data
	prices     map[string]float64
	symbols    map[string]string
	pairs      []string
}

func NewBinance(ctx context.Context, cfg *config.ExchangeConfig, calculator calculator.CalculateService) *Binance {
	httpClient := client.NewHTTPClient()

	binanceLogger := log.Logger.With().Str("logger", "binance").Logger()

	binanceCtx, cancel := context.WithCancel(ctx)

	binance := &Binance{
		ctx:        binanceCtx,
		cancel:     cancel,
		url:        cfg.URL,
		wsURL:      cfg.WsURL,
		logger:     &binanceLogger,
		httpClient: httpClient,
		chans:      make(map[string]map[string]chan<- *domain.Data),
		calculator: calculator,
		prices:     make(map[string]float64),
		symbols:    make(map[string]string),
		pairs:      cfg.Pairs,
	}

	go func() {
		for {
			if err := binance.run(); err != nil {
				binanceLogger.Error().Stack().Err(err).Msg("failed to run binance")
				binance.cancel()

				if err == errConsumingRateIsTooSlow {
					continue
				}

				return
			}

			break
		}
	}()

	return binance
}

func (e *Binance) run() error {
	logger := e.logger.With().Str("method", "run").Logger()

	logger.Info().Msg(strings.Join(e.pairs, ","))

	c, _, err := websocket.DefaultDialer.Dial(e.wsURL, nil)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to connect")
		return err
	}
	defer c.Close()

	//c.SetPingHandler(func(message string) error {
	//	logger.Info().Msg("send pong message")
	//	err := c.WriteControl(websocket.PongMessage, nil, time.Now().Add(time.Second))
	//	if err != nil {
	//		if err == websocket.ErrCloseSent {
	//			logger.Error().Stack().Err(err).Msg("close connection on pong message")
	//			return nil
	//		}
	//		logger.Error().Stack().Err(err).Msg("failed to send pong message")
	//	}
	//
	//	return err
	//})

	for _, pair := range e.pairs {
		promPrices[pair] = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   "calc",
			Name:        "binance_price",
			Help:        "pair price",
			ConstLabels: prometheus.Labels{"pair": pair},
		})
		prometheus.MustRegister(promPrices[pair])
		e.chans[pair] = make(map[string]chan<- *domain.Data)
		e.symbols[strings.ReplaceAll(pair, "_", "")] = pair
	}

	k := 0
	streams := make([][]string, 0)
	for _, pair := range e.pairs {
		if k%chunksCount == 0 {
			k = 0
		}

		if len(streams) == k {
			streams = append(streams, []string{})
		}

		streams[k] = append(streams[k], fmt.Sprintf("%s@bookTicker", strings.ReplaceAll(strings.ToLower(pair), "_", "")))
		k++
	}

	for i, stream := range streams {
		time.Sleep(time.Millisecond * 500)

		init := struct {
			Method string   `json:"method"`
			Params []string `json:"params"`
			Id     int      `json:"id"`
		}{
			Id:     i,
			Method: "SUBSCRIBE",
			Params: stream,
		}

		err = c.WriteJSON(init)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to write init message")
			return err
		}
		logger.Debug().Msgf("init message %v successful sended", init)
	}

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

			if ticker.Result != nil {
				logger.Error().Stack().Msgf("failed on response message [%s]", ticker.Result.ErrorMessage)
				return errors.New(ticker.Result.ErrorMessage)
			}

			pair := e.symbols[ticker.Symbol]

			if e.prices[pair] == ticker.Bid {
				continue
			}
			e.prices[pair] = ticker.Bid

			data := &domain.Data{
				Exchange: "binance",
				Pair:     pair,
				Price:    ticker.Bid,
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

func (e *Binance) Pairs(ctx context.Context) ([]string, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, exchangeInfoUri))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to parse url")
		return nil, err
	}

	resp, err := e.httpClient.Get(ctx, u.String())
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to request ticker price")
		return nil, err
	}

	defer resp.Body.Close()

	var exchangeInfo response.ExchangeInfo
	if err := json.NewDecoder(resp.Body).Decode(&exchangeInfo); err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to decode pair ticker price response")
		return nil, err
	}

	var pairs []string
	for _, symbol := range exchangeInfo.Symbols {
		if symbol.Status == "TRADING" && symbol.IsSpotTradingAllowed && symbol.HasPermission("SPOT") {
			pairs = append(pairs, fmt.Sprintf("%s_%s", symbol.BaseAsset, symbol.QuoteAsset))
		}
	}

	return pairs, nil
}

func (e *Binance) Price(ctx context.Context, pair string) (float64, error) {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, tickerPriceUri))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to parse url")
		return 0, err
	}

	q := u.Query()
	q.Add("symbol", strings.ReplaceAll(pair, "_", ""))

	u.RawQuery = q.Encode()

	resp, err := e.httpClient.Get(ctx, u.String())
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to request ticker price")
		return 0, err
	}

	defer resp.Body.Close()

	var tickerPrice response.TickerPrice
	if err := json.NewDecoder(resp.Body).Decode(&tickerPrice); err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to decode pair ticker price response")
		return 0, err
	}

	return tickerPrice.Price, nil
}

func (e *Binance) WSPrice(ctx context.Context, pair string, ch chan<- *domain.Data) {
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
