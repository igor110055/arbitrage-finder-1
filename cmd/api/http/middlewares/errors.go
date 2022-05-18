package middlewares

import (
	"calc/foundation/jwt"
	"calc/internal/berrors"
)

const baseCode = 11000

var (
	ErrNoToken = &berrors.BusinessError{
		ErrCode: baseCode + 1,
		Message: "no token provided in request",
	}
	ErrInvalidToken = &berrors.BusinessError{
		ErrCode: baseCode + 2,
		Message: jwt.ErrInvalidToken.Error(),
	}
)
