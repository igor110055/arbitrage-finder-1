package config

type Logger struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
