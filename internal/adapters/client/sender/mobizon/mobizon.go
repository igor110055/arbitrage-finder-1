package mobizon

import (
	"bytes"
	"calc/common/config"
	"calc/internal/adapters/client"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"strings"
)

const (
	messageUri = "/Message/SendSmsMessage"

	apiKeyParam = "apiKey"
)

type Mobizon struct {
	url        string
	apiKey     string
	logger     *zerolog.Logger
	httpClient client.HTTPClient
}

func NewMobizon(cfg *config.Sender) *Mobizon {
	httpClient := client.NewHTTPClient()

	mobizonLogger := log.Logger.With().Str("logger", "mobizon").Logger()

	mobizon := &Mobizon{
		url:        cfg.URL,
		apiKey:     cfg.ApiKey,
		logger:     &mobizonLogger,
		httpClient: httpClient,
	}

	return mobizon
}

func (e *Mobizon) Send(ctx context.Context, phone, sms string) error {
	u, err := url.Parse(fmt.Sprintf("%s%s", e.url, messageUri))
	if err != nil {
		e.logger.Error().Stack().Err(err).Msg("failed to parse url")
		return err
	}

	queryParams := u.Query()
	queryParams.Add(apiKeyParam, e.apiKey)
	u.RawQuery = queryParams.Encode()

	phone = strings.TrimSpace(strings.ReplaceAll(phone, "+", ""))
	b, err := json.Marshal(struct {
		Recipient string `json:"recipient"`
		Text      string `json:"text"`
	}{
		Recipient: phone,
		Text:      sms,
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

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("failed to send request with code %d", resp.StatusCode))
	}

	return nil
}
