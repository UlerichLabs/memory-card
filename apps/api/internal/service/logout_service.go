package service

import (
	"context"
	"fmt"
	"time"
)

type TokenRevogadoRepository interface {
	Revogar(ctx context.Context, jti string, expiraEm time.Time) error
	EstaRevogado(ctx context.Context, jti string) (bool, error)
}

type LogoutService struct {
	tokens    *AuthToken
	revogados TokenRevogadoRepository
}

func NewLogoutService(tokens *AuthToken, revogados TokenRevogadoRepository) *LogoutService {
	return &LogoutService{tokens: tokens, revogados: revogados}
}

func (svc *LogoutService) Logout(ctx context.Context, subject, raw string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("logout cancelado: %w", err)
	}
	claims, err := svc.tokens.validar(raw, "refresh")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSessaoExpirada, err)
	}
	if claims.Subject != subject {
		return ErrTokenInvalido
	}
	if err := svc.revogados.Revogar(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
		return fmt.Errorf("revogar refresh token: %w", err)
	}
	return nil
}
