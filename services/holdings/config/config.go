package config

import (
	"sync"

	"dario.cat/mergo"
	"github.com/caarlos0/env/v11"
)

var (
	once   sync.Once
	config *Config
)

func Load() (*Config, error) {
	var err error
	once.Do(func() {
		config, err = NewConfig()
	})
	return config, err
}

func NewConfig() (*Config, error) {
	conf := defaultConfig()
	envConfig := &Config{}

	err := env.ParseWithOptions(envConfig, env.Options{Prefix: "MONAY_HOLDINGS_"})
	if err != nil {
		return nil, err
	}

	err = mergo.Merge(conf, envConfig, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func defaultConfig() *Config {
	return &Config{
		ListenAddr:     ":8888",
		LoggingLevel:   "info",
		DBPath:         "./holdings.db",
		PortfolioURL:   "http://localhost:8889",
		BondServiceURL: "http://localhost:39995",
	}
}

type Config struct {
	ListenAddr      string `env:"LISTEN_ADDR"`
	LoggingLevel    string `env:"LOGGING_LEVEL"`
	DBPath          string `env:"DB_PATH"`
	PortfolioURL    string `env:"PORTFOLIO_URL"`
	BondServiceURL  string `env:"BOND_SERVICE_URL"`
	ConvictionsPath string `env:"CONVICTIONS_PATH"`
}
