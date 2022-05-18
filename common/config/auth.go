package config

import "time"

type Auth struct {
	PrivateKeyFile  string        `yaml:"private_key_file"`
	PublicKeyFile   string        `yaml:"public_key_file"`
	Algorithm       string        `yaml:"algorithm"`
	MaxAttempts     int           `yaml:"max_attempts"`
	AccessLifetime  time.Duration `yaml:"access_lifetime"`
	RefreshLifetime time.Duration `yaml:"refresh_lifetime"`
}
