package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/joho/godotenv"
)

type Config struct {
	Domain    string `env:"DOMAIN" default:"localhost"`
	Port      string `env:"PORT" default:"3333"`
	DBUrl     string `env:"DATABASE_URL" required:"true"`
	CertPath  string `env:"CERT_PATH" required:"true"`
	KeyPath   string `env:"KEY_PATH" required:"true"`
	TestEmpty string `env:"TEST_EMPTY"`
	Env       string `env:"ENV" default:"dev"`
}

var config *Config

func Get() *Config {
	return config
}

func findEnvFile(envFile string) (string, error) {
	// If it's an absolute path, check it directly
	if filepath.IsAbs(envFile) {
		if _, err := os.Stat(envFile); err != nil {
			return "", fmt.Errorf("env file not found at absolute path: %s", envFile)
		}
		return envFile, nil
	}

	// For relative paths, search up the directory tree
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		envPath := filepath.Join(dir, envFile)
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}

		// Get parent directory
		parent := filepath.Dir(dir)
		// If we're already at the root, stop searching
		if parent == dir {
			return "", fmt.Errorf("env file %s not found in directory tree", envFile)
		}
		dir = parent
	}
}

func init() {
	envFile := ".env"

	if env := os.Getenv("DOTENV"); env != "" {
		envFile = env
	}

	envPath, err := findEnvFile(envFile)
	if err != nil {
		panic(fmt.Sprintf("Error: %v\n", err))
	} else {
		fmt.Printf("Loading env from: %s\n", envPath)
		if err := godotenv.Load(envPath); err != nil {
			panic(fmt.Sprintf("Error loading env file: %v\n", err))
		}
	}

	var configErr error
	config, configErr = loadConfig()
	if configErr != nil {
		panic(configErr)
	}
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
