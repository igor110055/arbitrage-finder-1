package config

type Exchange struct {
	Pairs   []string                   `yaml:"pairs"`
	Configs map[string]*ExchangeConfig `yaml:"configs"`
}

type ExchangeConfig struct {
	URL   string   `yaml:"url"`
	WsURL string   `yaml:"ws_url"`
	Pairs []string `yaml:"pairs"`
}
