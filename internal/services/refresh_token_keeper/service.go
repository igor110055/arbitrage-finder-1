package refresh_token_keeper

import (
	"calc/foundation/jwt"
	"calc/internal/adapters/db"
	"context"
	"github.com/oklog/ulid/v2"
	"github.com/pkg/errors"
)

const maxRefreshTokensPerUser = 1

type Service struct {
	refreshTokenRepo db.RefreshTokenRepo
}

func NewService(refreshTokenRepo db.RefreshTokenRepo) *Service {
	return &Service{
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (s *Service) Save(ctx context.Context, accountID uint64, tokenID ulid.ULID) error {
	tokensCount, err := s.refreshTokenRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return errors.Wrapf(err, "failed to count refresh tokens by account ID=%d", accountID)
	}

	if tokensCount >= maxRefreshTokensPerUser {
		tokensToDelete := int(tokensCount) - maxRefreshTokensPerUser + 1
		for i := 0; i < tokensToDelete; i++ {
			if err := s.refreshTokenRepo.DeleteOldestByAccountID(ctx, accountID); err != nil {
				return errors.Wrapf(err, "failed to delete oldest refresh token by account ID=%d", accountID)
			}
		}
	}

	err = s.refreshTokenRepo.Create(ctx, accountID, tokenID)
	return errors.Wrapf(err, "failed to create refresh token, accountID: %d, token: %s", accountID, tokenID)
}

func (s *Service) Delete(ctx context.Context, tokenID ulid.ULID) error {
	n, err := s.refreshTokenRepo.Delete(ctx, tokenID)
	if err != nil {
		return errors.Wrapf(err, "failed to delete refresh tokenID ID=%s", tokenID)
	}

	if n == 0 {
		return errors.Wrapf(jwt.ErrInvalidToken, "refresh token=%s doesn't exist in SQL-database", tokenID)
	}

	return nil
}
