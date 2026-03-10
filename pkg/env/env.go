package env

import (
	"github.com/caarlos0/env/v10"
	"go.uber.org/fx"
)

type Config struct {
	Port        string `env:"PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

var Module = fx.Provide(NewConfig)
