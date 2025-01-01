package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/joho/godotenv"
)

func NewConfig() (*Config, error) {
	if env := os.Getenv("DOTENV"); env != "" {
		godotenv.Load(env)
	} else {
		godotenv.Load()
	}

	return loadConfig()
}

type Config struct {
	Port      string `env:"PORT" default:"3333"`
	DBUrl     string `env:"DATABASE_URL" required:"true"`
	CertPath  string `env:"CERT_PATH" required:"true"`
	KeyPath   string `env:"KEY_PATH" required:"true"`
	TestEmpty string `env:"TEST_EMPTY"`
}

func loadConfig() (*Config, error) {
	cfg := &Config{}
	t := reflect.TypeOf(*cfg)
	v := reflect.ValueOf(cfg).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envKey := field.Tag.Get("env")
		if envKey == "" {
			continue
		}

		val := os.Getenv(envKey)
		if val == "" {
			if field.Tag.Get("required") == "true" {
				return nil, fmt.Errorf("required environment variable %s not set", envKey)
			}
			if def := field.Tag.Get("default"); def != "" {
				val = def
			}
		}
		v.Field(i).SetString(val)
	}

	return cfg, nil
}
