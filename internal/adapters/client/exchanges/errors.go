package exchanges

import (
	"github.com/pkg/errors"
)

var (
	ErrExchangeNotFound     = errors.New("exchange not found")
	ErrExchangeNotImplement = errors.New("exchange not implement")
)
