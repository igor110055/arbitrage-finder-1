package config

import "time"

type Server struct {
	Host            string        `yaml:"host"`
	HttpPort        int           `yaml:"http_port"`
	DebugPort       int           `yaml:"debug_port"`
	UseTLS          bool          `yaml:"use_tls"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}
