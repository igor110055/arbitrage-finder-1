package mocks

import (
	"context"
	"github.com/rs/zerolog/log"
)

type DummySMSSender struct{}

func NewDummySMSSender() *DummySMSSender {
	return &DummySMSSender{}
}

func (s *DummySMSSender) Send(ctx context.Context, phone, sms string) error {
	log.Debug().Msgf("DummySMSSender [%s]: %s", phone, sms)

	return nil
}
