package main

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Port string `yaml:"port" env:"PORT"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("config.yml", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
