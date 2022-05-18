package exchanges

import (
	"calc/internal/domain"
	"context"
)

type Exchange interface {
	Pairs(ctx context.Context) ([]string, error)
	Price(ctx context.Context, pair string) (float64, error)
	WSPrice(ctx context.Context, pair string, ch chan<- *domain.Data)
}
