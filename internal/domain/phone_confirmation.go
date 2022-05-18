package domain

import "time"

type PhoneConfirmation struct {
	ID                uint64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	AccountID         uint64
	Phone             string
	Code              string
	RemainingAttempts int
	Used              bool
}
