package config

import (
	"go.uber.org/config"
	"os"
)

type Config struct {
	Env       string    `yaml:"env"`
	AppName   string    `yaml:"app"`
	Version   string    `yaml:"version"`
	Server    *Server   `yaml:"server"`
	Database  *DB       `json:"database"`
	Logger    *Logger   `yaml:"logger"`
	Auth      *Auth     `yaml:"auth"`
	Exchanges *Exchange `yaml:"exchanges"`
	Sender    *Sender   `yaml:"sender"`
}

var cfg Config

func NewConfig(filename string) (*Config, error) {
	provider, err := config.NewYAML(append(
		[]config.YAMLOption{
			config.Expand(os.LookupEnv),
		},
		config.File(filename),
	)...)
	if err != nil {
		return nil, err
	}

	err = provider.Get("").Populate(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
