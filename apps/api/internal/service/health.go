// Package service concentra as regras de negócio da aplicação.
package service

import (
	"context"
	"errors"
	"fmt"
)

// ErrDatabaseUnavailable indica falha na verificação de conectividade com o banco de dados.
var ErrDatabaseUnavailable = errors.New("banco de dados indisponível")

// DatabasePinger define o contrato para verificação de conectividade com o banco de dados.
type DatabasePinger interface {
	Ping(ctx context.Context) error
}

// HealthService coordena as verificações de saúde dos componentes da aplicação.
type HealthService struct {
	db DatabasePinger
}

// NewHealthService cria uma nova instância de HealthService com o pinger configurado.
func NewHealthService(db DatabasePinger) *HealthService {
	return &HealthService{db: db}
}

// CheckDatabase verifica se o banco de dados está respondendo a pings.
func (s *HealthService) CheckDatabase(ctx context.Context) error {
	if err := s.db.Ping(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrDatabaseUnavailable, err)
	}
	return nil
}
