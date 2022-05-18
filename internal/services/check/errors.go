package check

import "calc/internal/berrors"

const baseCode = 10000

var (
	ErrDBNotReady = &berrors.BusinessError{
		ErrCode: baseCode + 1,
		Message: "db not ready",
	}
)

func Errors() []*berrors.BusinessError {
	return []*berrors.BusinessError{
		ErrDBNotReady,
	}
}
