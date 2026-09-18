package service

import (
	"context"
	"errors"
	"testing"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type perfilRepoMock struct {
	buscar func(context.Context, int32) (*repository.Usuario, error)
}

func (mock perfilRepoMock) BuscarPorID(ctx context.Context, id int32) (*repository.Usuario, error) {
	return mock.buscar(ctx, id)
}

func TestPerfil_Cenarios(t *testing.T) {
	usuario := &repository.Usuario{ID: 42, Nome: "Lucas", Email: "lucas@example.com", Idioma: "en"}
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name             string
		usuario          *repository.Usuario
		repoErr, wantErr error
	}{
		{"sucesso", usuario, nil, nil},
		{"usuario removido", nil, nil, ErrTokenInvalido},
		{"falha no banco", nil, dbErr, dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			called := false
			svc := NewPerfilService(perfilRepoMock{buscar: func(got context.Context, id int32) (*repository.Usuario, error) {
				called = true
				if got != ctx || id != usuario.ID {
					t.Fatal("contexto ou identidade incorretos")
				}
				return tc.usuario, tc.repoErr
			}})
			result, err := svc.Perfil(ctx, "42")
			if !called || !errors.Is(err, tc.wantErr) || result != tc.usuario {
				t.Fatalf("perfil = %v, erro = %v", result, err)
			}
		})
	}
}

func TestPerfil_IdentidadeInvalida(t *testing.T) {
	svc := NewPerfilService(perfilRepoMock{buscar: func(context.Context, int32) (*repository.Usuario, error) {
		t.Fatal("nao deveria consultar repositorio")
		return nil, nil
	}})
	for _, subject := range []string{"", "abc", "0", "-1", "2147483648"} {
		t.Run(subject, func(t *testing.T) {
			result, err := svc.Perfil(context.Background(), subject)
			if result != nil || !errors.Is(err, ErrTokenInvalido) {
				t.Fatalf("erro = %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.Perfil(ctx, "42"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
