package config

import (
	"fmt"
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
		ListenAddr:   ":8888",
		LoggingLevel: "info",
		Database: Database{
			Host: "postgres.monay.local",
			Port: "5432",
			Name: "monay",
			User: "monay_admin",
		},
		Plaid: Plaid{
			Env: "sandbox",
		},
	}
}

type Config struct {
	ListenAddr   string   `env:"LISTEN_ADDR"`
	LoggingLevel string   `env:"LOGGING_LEVEL"`
	Database     Database `envPrefix:"POSTGRES_"`
	Plaid        Plaid    `envPrefix:"PLAID_"`
}

type Database struct {
	Host   string `env:"HOST"`
	Port   string `env:"PORT"`
	Name   string `env:"DATABASE"`
	User   string `env:"USER"`
	Region string `env:"REGION"`
	// If set, the password to use to connect to the database.
	// If nil, will use RDS IAM authentication in production.
	Password *string `env:"PASSWORD"`
}

func (d Database) ConnStringWithoutPassword() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Name,
	)
}

func (d Database) ConnString() string {
	connString := d.ConnStringWithoutPassword()
	if d.Password != nil {
		connString = fmt.Sprintf("%s password=%s", connString, *d.Password)
	}
	return connString
}

type Plaid struct {
	ClientID    string `env:"CLIENT_ID"`
	Secret      string `env:"SECRET"`
	Env         string `env:"ENV"`
	RedirectURI string `env:"REDIRECT_URI"` // Required for OAuth institutions in production (must be HTTPS)
}
