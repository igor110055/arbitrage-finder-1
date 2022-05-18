package sender

import "context"

type Sender interface {
	Send(ctx context.Context, phone, sms string) error
}

