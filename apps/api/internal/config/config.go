// Package config lê a configuração do processo a partir do ambiente.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config armazena as variáveis de ambiente necessárias para inicialização da API.
type Config struct {
	DatabaseURL      string
	Port             string
	JWTSecret        string
	IGDBClientID     string
	IGDBClientSecret string
}

// Load lê e valida as variáveis de ambiente necessárias para a aplicação.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		Port:             os.Getenv("PORT"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		IGDBClientID:     os.Getenv("IGDB_CLIENT_ID"),
		IGDBClientSecret: os.Getenv("IGDB_CLIENT_SECRET"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL é obrigatória")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return Config{}, fmt.Errorf("interpretar PORT: %w", err)
	}
	if port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT deve estar entre 1 e 65535")
	}
	return cfg, nil
}

type AuthConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func LoadAuth() (AuthConfig, error) {
	cfg := AuthConfig{Secret: os.Getenv("JWT_SECRET")}
	if len(cfg.Secret) < 32 {
		return AuthConfig{}, fmt.Errorf("JWT_SECRET deve ter pelo menos 32 bytes")
	}
	var err error
	cfg.AccessTTL, err = time.ParseDuration(os.Getenv("JWT_ACCESS_TTL"))
	if err != nil {
		return AuthConfig{}, fmt.Errorf("interpretar JWT_ACCESS_TTL: %w", err)
	}
	cfg.RefreshTTL, err = time.ParseDuration(os.Getenv("JWT_REFRESH_TTL"))
	if err != nil {
		return AuthConfig{}, fmt.Errorf("interpretar JWT_REFRESH_TTL: %w", err)
	}
	if cfg.AccessTTL < time.Second || cfg.RefreshTTL <= cfg.AccessTTL {
		return AuthConfig{}, fmt.Errorf("JWT_ACCESS_TTL deve ser positivo e menor que JWT_REFRESH_TTL")
	}
	return cfg, nil
}
