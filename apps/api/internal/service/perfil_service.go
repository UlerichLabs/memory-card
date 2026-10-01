package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type PerfilRepository interface {
	BuscarPorID(ctx context.Context, id int32) (*repository.Usuario, error)
}

type PerfilService struct {
	repo PerfilRepository
}

func NewPerfilService(repo PerfilRepository) *PerfilService {
	return &PerfilService{repo: repo}
}

func (svc *PerfilService) Perfil(ctx context.Context, subject string) (*repository.Usuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("perfil cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}
	usuario, err := svc.repo.BuscarPorID(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("consultar perfil: %w", err)
	}
	if usuario == nil {
		return nil, ErrTokenInvalido
	}
	return usuario, nil
}
