package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	SeedDemo    bool // cria os dados de demonstração quando o banco não tem usuários
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	seedDemo := false
	if valor := os.Getenv("SEED_DEMO"); valor != "" {
		ligado, err := strconv.ParseBool(valor)
		if err != nil {
			return nil, fmt.Errorf("SEED_DEMO must be true or false, got %q", valor)
		}
		seedDemo = ligado
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		SeedDemo:    seedDemo,
	}, nil
}
