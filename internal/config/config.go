package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type HTTP struct {
	Port              string        `yaml:"port"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
}

type Mongo struct {
	URI string `env:"MONGO_URI" env-required:"true"`
}

type Service struct {
	DefaultTimeout time.Duration `yaml:"default_timeout"`
}

type Config struct {
	HTTP    HTTP    `yaml:"http"`
	Service Service `yaml:"service"`
	Mongo   Mongo   `yaml:"mongo"`
}

func New(path string) (*Config, error) {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	return cfg, nil
}

func LoadEnv(filenames ...string) error {
	if len(filenames) == 0 {
		return godotenv.Load()
	}

	for _, filename := range filenames {
		if err := godotenv.Load(filename); err != nil {
			return fmt.Errorf("loading env file %s: %w", filename, err)
		}
	}
	return nil
}
