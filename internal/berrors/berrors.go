package berrors

import (
	"github.com/pkg/errors"
	"strconv"
)

type BusinessError struct {
	ErrCode    int    `json:"err_code"`
	Message    string `json:"message"`
	NeedCommit bool   `swaggerignore:"true" json:"-"`
}

func (e *BusinessError) Code() int {
	return e.ErrCode
}

func (e *BusinessError) Error() string {
	return e.Message + ": error code " + strconv.Itoa(e.ErrCode)
}

func WrapWithError(bErr *BusinessError, err error) error {
	return errors.Wrap(bErr, err.Error())
}
