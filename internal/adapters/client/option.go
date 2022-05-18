package client

type Settings struct {
	ApiKey string
}

type Option interface {
	Apply(*Settings)
}

type withApiKey string

func (w withApiKey) Apply(o *Settings) {
	o.ApiKey = string(w)
}

func WithApiKey(apiKey string) Option {
	return withApiKey(apiKey)
}
