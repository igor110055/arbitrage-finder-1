package auth

import "calc/internal/berrors"

const baseCode = 11000

var (
	ErrInvalidInput = &berrors.BusinessError{
		ErrCode: baseCode + 1,
		Message: "got invalid input",
	}
	ErrInvalidRole = &berrors.BusinessError{
		ErrCode: baseCode + 2,
		Message: "got invalid role",
	}
	ErrAccountAlreadyExists = &berrors.BusinessError{
		ErrCode: baseCode + 3,
		Message: "this account already exists",
	}
	ErrConfirmationCodeAlreadySent = &berrors.BusinessError{
		ErrCode: baseCode + 4,
		Message: "confirmation code is already sent",
	}
	ErrConfirmationNotFound = &berrors.BusinessError{
		ErrCode: baseCode + 5,
		Message: "confirmation not found",
	}
	ErrConfirmationAlreadyUsed = &berrors.BusinessError{
		ErrCode: baseCode + 6,
		Message: "confirmation already used",
	}
	ErrAttemptsLimitReached = &berrors.BusinessError{
		ErrCode: baseCode + 7,
		Message: "confirm attempts limit is reached",
	}
	ErrExpiredConfirmationCode = &berrors.BusinessError{
		ErrCode: baseCode + 8,
		Message: "confirmation code is expired",
	}
	ErrInvalidConfirmationCode = &berrors.BusinessError{
		ErrCode:    baseCode + 9,
		Message:    "invalid confirmation code",
		NeedCommit: true,
	}
	ErrAccountNotFound = &berrors.BusinessError{
		ErrCode: baseCode + 10,
		Message: "account not found",
	}
	ErrForbidden = &berrors.BusinessError{
		ErrCode: baseCode + 11,
		Message: "forbidden",
	}
	ErrNotSuitableStatus = &berrors.BusinessError{
		ErrCode: baseCode + 12,
		Message: "not suitable status",
	}
	ErrAccountBanned = &berrors.BusinessError{
		ErrCode: baseCode + 13,
		Message: "account is banned",
	}
)

func Errors() []*berrors.BusinessError {
	return []*berrors.BusinessError{
		ErrInvalidInput,
		ErrInvalidRole,
		ErrAccountAlreadyExists,
		ErrConfirmationCodeAlreadySent,
		ErrConfirmationNotFound,
		ErrConfirmationAlreadyUsed,
		ErrAttemptsLimitReached,
		ErrExpiredConfirmationCode,
		ErrInvalidConfirmationCode,
		ErrAccountNotFound,
		ErrForbidden,
		ErrNotSuitableStatus,
		ErrAccountBanned,
	}
}
