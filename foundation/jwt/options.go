package jwt

import "time"

type AuthenticatorOption func(*Authenticator)

// WithAccessTokenLifetime sets access token lifetime
func WithAccessTokenLifetime(lifetime time.Duration) AuthenticatorOption {
	return func(auth *Authenticator) {
		auth.accessLifetime = lifetime
	}
}

// WithRefreshTokenLifetime sets refresh token lifetime
func WithRefreshTokenLifetime(lifetime time.Duration) AuthenticatorOption {
	return func(auth *Authenticator) {
		auth.refreshLifetime = lifetime
	}
}

// WithRefreshTokenKeeper adds RefreshTokenKeeper to Authenticator
func WithRefreshTokenKeeper(rtKeeper RefreshTokenKeeper) AuthenticatorOption {
	return func(auth *Authenticator) {
		auth.rtKeeper = rtKeeper
	}
}
