package auth

import (
	"calc/foundation/hash"
	"calc/foundation/jwt"
	"calc/foundation/random"
	"calc/internal/adapters/client/sender"
	"calc/internal/adapters/db"
	"calc/internal/domain"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

const (
	ConfirmationCodeLen      = 4
	ConfirmationCodeLifetime = time.Hour
	RepeatTime               = 30 * time.Second
)

type Service struct {
	accountRepo           db.AccountRepo
	phoneConfirmationRepo db.PhoneConfirmationRepo
	smsSender             sender.Sender
	jwtAuth               *jwt.Authenticator
	maxAttempts           int
}

func NewService(
	accountRepo db.AccountRepo,
	phoneConfirmationRepo db.PhoneConfirmationRepo,
	smsSender sender.Sender,
	jwtAuth *jwt.Authenticator,
	maxAttempts int,
) *Service {
	return &Service{
		accountRepo:           accountRepo,
		phoneConfirmationRepo: phoneConfirmationRepo,
		smsSender:             smsSender,
		jwtAuth:               jwtAuth,
		maxAttempts:           maxAttempts,
	}
}

type SignUpArgs struct {
	Phone    string
	Password string
}

func (s *Service) SignUp(ctx context.Context, args *SignUpArgs) (uint64, error) {
	passHash := hash.GenerateHash(args.Password)

	account, err := s.accountRepo.Create(ctx, &domain.Account{
		ID:       0,
		Phone:    args.Phone,
		Password: passHash,
		Status:   domain.AccountStatusCreated,
	})
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			return 0, ErrAccountAlreadyExists
		}

		return 0, err
	}

	return s.CreatePhoneConfirmation(ctx, account.Phone)
}

func (s *Service) CreatePhoneConfirmation(ctx context.Context, phone string) (uint64, error) {
	account, err := s.accountRepo.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.Wrapf(ErrAccountNotFound, "account with phone %s doesnt exist", phone)
		}

		return 0, errors.Wrapf(err, "failed to find account by phone %s", phone)
	}

	if account.Status != domain.AccountStatusCreated {
		return 0, errors.Wrapf(ErrNotSuitableStatus, "account is %s", account.Status.String())
	}

	confirmation, err := s.phoneConfirmationRepo.FindByAccountID(ctx, account.ID)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		code, err := random.NumCode(ConfirmationCodeLen)
		if err != nil {
			return 0, err
		}

		confirmation, err = s.phoneConfirmationRepo.Create(ctx, &domain.PhoneConfirmation{
			AccountID:         account.ID,
			Phone:             phone,
			Code:              strconv.Itoa(int(code)),
			RemainingAttempts: s.maxAttempts,
			Used:              false,
		})
		if err != nil {
			return 0, errors.Wrapf(err, "failed to create phone confirmation (phone: %q)", phone)
		}

		log.Info().Msgf("code %s generated for phone %s", confirmation.Code, confirmation.Phone)
	case err != nil:
		return 0, errors.Wrap(err, "find confirmation by account id: database error")
	default:
		timePassed := time.Since(confirmation.UpdatedAt)

		if !confirmation.Used && timePassed <= RepeatTime {
			return 0, errors.Wrapf(ErrConfirmationCodeAlreadySent,
				"confirmation code was already sent, try in %s", RepeatTime-timePassed)
		}

		code, err := random.NumCode(ConfirmationCodeLen)
		if err != nil {
			return 0, err
		}

		confirmation.Code = strconv.Itoa(int(code))
		confirmation.RemainingAttempts = s.maxAttempts

		if err := s.phoneConfirmationRepo.Update(ctx, confirmation); err != nil {
			return 0, errors.Wrapf(err, "failed to reset phone confirmation")
		}

		log.Info().Msgf("code %s generated for phone %s", confirmation.Code, confirmation.Phone)
	}

	if err := s.smsSender.Send(ctx, phone, confirmation.Code); err != nil {
		return 0, errors.Wrapf(err, "failed to send releans on %s phone", phone)
	}

	return confirmation.ID, nil
}

type ConfirmArgs struct {
	ConfirmationID uint64
	Password       string
	Code           string
}

func (s *Service) Confirm(ctx context.Context, args *ConfirmArgs) (*jwt.TokenPair, error) {
	confirmation, err := s.phoneConfirmationRepo.FindByID(ctx, args.ConfirmationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrapf(ErrConfirmationNotFound, "phone confirmation with id %d doesnt exist", args.ConfirmationID)
		}

		return nil, errors.Wrapf(err, "failed to find phone confirmation by id %q", args.ConfirmationID)
	}

	account, err := s.accountRepo.FindByID(ctx, confirmation.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrapf(ErrAccountNotFound, "account with id %d doesnt exist", confirmation.AccountID)
		}

		return nil, errors.Wrapf(err, "failed to find account by id %q", confirmation.AccountID)
	}

	passHash := hash.GenerateHash(args.Password)
	if passHash != account.Password {
		return nil, errors.Wrap(ErrForbidden, "invalid password on confirm phone")
	}

	if confirmation.Used {
		return nil, errors.Wrapf(ErrConfirmationAlreadyUsed, "phone confirmation %d", args.ConfirmationID)
	}

	if confirmation.RemainingAttempts <= 0 {
		return nil, errors.Wrapf(ErrAttemptsLimitReached, "attempts limit reached for %d", args.ConfirmationID)
	}

	if time.Since(confirmation.CreatedAt) > ConfirmationCodeLifetime {
		return nil, errors.Wrapf(ErrExpiredConfirmationCode, "confirmation code is expired for %d", args.ConfirmationID)
	}

	if confirmation.Code != args.Code {
		confirmation.RemainingAttempts--
		if err := s.phoneConfirmationRepo.Update(ctx, confirmation); err != nil {
			return nil, errors.Wrapf(err, "failed to decrement phone confirmation attempts id=%d", args.ConfirmationID)
		}

		return nil, errors.Wrapf(ErrInvalidConfirmationCode, "invalid confirmation code for %d", args.ConfirmationID)
	}

	confirmation.Used = true
	if err := s.phoneConfirmationRepo.Update(ctx, confirmation); err != nil {
		return nil, errors.Wrapf(err, "failed to set used phone confirmation for %d", args.ConfirmationID)
	}

	account.Status = domain.AccountStatusActive
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, errors.Wrapf(err, "failed to set account status for %d", account.ID)
	}

	tp, err := s.jwtAuth.GenerateTokenPair(ctx, confirmation.AccountID)
	if err != nil {
		return nil, err
	}

	return tp, nil
}

func (s *Service) SignIn(ctx context.Context, phone, password string) (*jwt.TokenPair, error) {
	passHash := hash.GenerateHash(password)

	account, err := s.accountRepo.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrapf(ErrAccountNotFound, "account with phone %s doesnt exist", phone)
		}

		return nil, errors.Wrapf(err, "failed to find account by phone = %q", phone)
	}

	if account.Password != passHash {
		return nil, ErrForbidden
	}

	if account.Status == domain.AccountStatusBanned {
		return nil, ErrAccountBanned
	}

	if account.Status != domain.AccountStatusActive {
		return nil, errors.Wrapf(ErrNotSuitableStatus, "accaount is %s", account.Status.String())
	}

	pair, err := s.jwtAuth.GenerateTokenPair(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (s *Service) Refresh(ctx context.Context, accountID uint64) (*jwt.TokenPair, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find account by id = %d", accountID)
	}

	if account.Status == domain.AccountStatusBanned {
		return nil, ErrAccountBanned
	}

	pair, err := s.jwtAuth.GenerateTokenPair(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (s *Service) CheckPhone(ctx context.Context, phone string) (bool, error) {
	if _, err := s.accountRepo.FindByPhone(ctx, phone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, errors.Wrapf(err, "failed to find account by phone %q", phone)
	}

	return true, nil
}
