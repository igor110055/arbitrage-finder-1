package jwt

import (
	"context"
	"github.com/oklog/ulid/v2"
)

// RefreshTokenKeeper saves/deletes refresh token attached to user in storage
type RefreshTokenKeeper interface {
	Save(ctx context.Context, accountID uint64, tokenID ulid.ULID) error
	Delete(ctx context.Context, tokenID ulid.ULID) error
}
