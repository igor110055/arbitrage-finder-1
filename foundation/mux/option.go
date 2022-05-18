package mux

import (
	"calc/internal/adapters/db"
	"context"
)

type HandleOptions struct {
	ctx      context.Context
	commit   func() error
	rollback func() error
}

func NewHandleOptions(ctx context.Context, options ...HandleOption) (*HandleOptions, error) {
	opts := &HandleOptions{
		ctx:      ctx,
		commit:   func() error { return nil },
		rollback: func() error { return nil },
	}

	for _, opt := range options {
		if err := opt(opts); err != nil {
			return nil, err
		}
	}

	return opts, nil
}

type HandleOption func(opts *HandleOptions) error

func WithTx(conn db.DB) HandleOption {
	return func(opts *HandleOptions) error {
		ctx, commit, rollback, err := conn.RunTx(opts.ctx)
		if err != nil {
			return err
		}

		opts.ctx = ctx
		opts.commit = commit
		opts.rollback = rollback

		return nil
	}
}
