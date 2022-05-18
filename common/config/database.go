package config

type DB struct {
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Host         string `yaml:"host"`
	Name         string `yaml:"name"`
	DisableTLS   bool   `yaml:"disable_tls"`
	CertPath     string `yaml:"cert_path"`
	Log          bool   `yaml:"log"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}
