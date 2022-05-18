package exchange

import "calc/internal/berrors"

const baseCode = 12000

var (
	ErrCannotConnect = &berrors.BusinessError{
		ErrCode: baseCode + 1,
		Message: "cannot connect to exchange",
	}
)
