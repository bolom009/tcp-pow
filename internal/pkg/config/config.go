package config

import "github.com/caarlos0/env/v6"

type Config struct {
	Server struct {
		Host string `env:"SERVER_HOST"`
		Port int    `env:"SERVER_PORT"`
	}
	Redis struct {
		Host string `env:"REDIS_HOST"`
		Port int    `env:"REDIS_PORT"`
	}
	Hashcash struct {
		ZeroCount     int   `env:"HASHCASH_ZERO_COUNT"`
		Duration      int64 `env:"HASHCASH_DURATION"`
		MaxIterations int   `env:"HASHCASH_MAX_ITERATIONS"`
	}
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
