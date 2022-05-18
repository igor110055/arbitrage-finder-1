package releans

import (
	"bytes"
	"calc/common/config"
	"calc/internal/adapters/client"
	"calc/internal/adapters/client/sender/releans/response"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
)

const (
	messageUri = "/message"
)

type Releans struct {
	url        string
	senderID   string
	logger     *zerolog.Logger
	httpClient client.HTTPClient
}

func NewReleans(cfg *config.Sender) *Releans {
	httpClient := client.NewHTTPClient(client.WithApiKey(cfg.ApiKey))

	releansLogger := log.Logger.With().Str("logger", "releans").Logger()

	releans := &Releans{
		url: cfg.URL,
		//senderID:   cfg.SenderID,
		logger:     &releansLogger,
		httpClient: httpClient,
	}

	return releans
}

func (e *Releans) Send(ctx context.Context, phone, sms string) error {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, messageUri))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to parse url")
		return err
	}

	b, err := json.Marshal(struct {
		Sender  string `json:"sender"`
		Mobile  string `json:"mobile"`
		Content string `json:"content"`
	}{
		Sender:  e.senderID,
		Mobile:  phone,
		Content: sms,
	})
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to marshal request")
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(b))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to make request")
		return err
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to send message")
		return err
	}

	defer resp.Body.Close()

	var msg response.Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to decode ticker response")
		return err
	}

	if msg.Code != 201 {
		return errors.New(fmt.Sprintf("%d: %s", msg.Status, msg.Message))
	}

	return nil
}
