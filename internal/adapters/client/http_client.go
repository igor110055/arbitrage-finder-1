package client

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	dialContextTimeout   = 60 * time.Second
	dialContextKeepAlive = 60 * time.Second
	httpClientTimeout    = 5 * time.Minute
	tlsHandshakeTimeout  = 15 * time.Second
)

var ErrClientInternal = errors.New("api.http_client_internal_error")

type HTTPClient interface {
	Get(ctx context.Context, url string) (*http.Response, error)
	Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)
	PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error)
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

type httpClient struct {
	cli      *http.Client
	logger   *zerolog.Logger
	settings *Settings
}

func NewHTTPClient(options ...Option) HTTPClient {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialContextTimeout,
			KeepAlive: dialContextKeepAlive,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout: tlsHandshakeTimeout,
		MaxConnsPerHost:     1000,
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
	}
	cli := &http.Client{
		Transport: t,
		Timeout:   httpClientTimeout,
	}

	settings := &Settings{}
	for _, option := range options {
		option.Apply(settings)
	}

	httpLogger := log.Logger.With().Str("logger", "http_client").Logger()

	return &httpClient{
		cli:      cli,
		logger:   &httpLogger,
		settings: settings,
	}
}

func (h *httpClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return h.Do(ctx, req)
}

func (h *httpClient) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	return h.Post(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (h *httpClient) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return h.Do(ctx, req)
}

func (h *httpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if h.settings.ApiKey != "" {
		req.Header.Add("Authorization", "Bearer "+h.settings.ApiKey)
	}

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Logger.Error().Stack().Err(err).Msg("failed to dump request")
	}
	h.logger.Debug().Msgf("dump request: %s", string(dump))

	debugDuration := time.Now()
	resp, err := h.cli.Do(req.Clone(ctx))
	if resp != nil {
		dumpResp, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Logger.Error().Stack().Err(err).Msg("failed to dump response")
		}
		h.logger.Debug().Msgf("dump response: %s", string(dumpResp))
	}
	if err != nil {
		h.logger.Error().Stack().Err(err).Msg("failed to make request")
		return nil, err
	}

	h.logger.Debug().Msgf("debugDuration=%s", time.Now().Sub(debugDuration))

	return resp, err
}
