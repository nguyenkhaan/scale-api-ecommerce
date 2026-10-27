package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseUrl string
	Port        string
	Env         string
}

func loadDotEnv() error {
	err := godotenv.Load()
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("load .env: %w", err)
}

// IsLocalEnvironment reports whether the API runs on plain-HTTP developer or test machines.
func IsLocalEnvironment(environment string) bool {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "dev", "development", "test":
		return true
	}
	return false
}

func Load() (Config, error) {
	if err := loadDotEnv(); err != nil {
		return Config{}, err
	}
	databaseUrl := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseUrl == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "4000"
	}
	env := strings.TrimSpace(os.Getenv("ENV"))
	return Config{
		DatabaseUrl: databaseUrl,
		Port:        port,
		Env:         env,
	}, nil
}
