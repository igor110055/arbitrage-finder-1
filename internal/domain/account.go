package domain

import (
	"database/sql/driver"
	"github.com/pkg/errors"
)

type Account struct {
	ID       uint64
	Phone    string
	Password string
	Status   AccountStatus
}

type AccountStatus struct{ string }

var (
	AccountStatusCreated = AccountStatus{"created"}
	AccountStatusActive  = AccountStatus{"active"}
	AccountStatusBanned  = AccountStatus{"banned"}
)

func GetAccountStatus(s string) (AccountStatus, error) {
	switch s {
	case AccountStatusCreated.string, AccountStatusActive.string, AccountStatusBanned.string:
		return AccountStatus{s}, nil
	default:
		return AccountStatus{}, errors.New("invalid enum param")
	}
}

func (e AccountStatus) String() string {
	return e.string
}

func (e AccountStatus) Value() (driver.Value, error) {
	return e.string, nil
}

func (e *AccountStatus) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.Errorf("type assertion to string failed")
	}

	en, err := GetAccountStatus(s)
	if err != nil {
		return err
	}

	*e = en

	return nil
}
