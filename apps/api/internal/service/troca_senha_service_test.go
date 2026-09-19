package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type trocaSenhaRepoMock struct {
	buscarFn    func(context.Context, int32) (*repository.CredenciaisUsuario, error)
	atualizarFn func(context.Context, int32, string) error
	listarFn    func(context.Context, int32) ([]repository.RefreshTokenAtivo, error)
	removerFn   func(context.Context, int32) error
}

func (mock trocaSenhaRepoMock) BuscarCredenciaisPorID(ctx context.Context, id int32) (*repository.CredenciaisUsuario, error) {
	return mock.buscarFn(ctx, id)
}

func (mock trocaSenhaRepoMock) AtualizarSenha(ctx context.Context, id int32, senhaHash string) error {
	return mock.atualizarFn(ctx, id, senhaHash)
}

func (mock trocaSenhaRepoMock) ListarRefreshTokensAtivos(ctx context.Context, id int32) ([]repository.RefreshTokenAtivo, error) {
	return mock.listarFn(ctx, id)
}

func (mock trocaSenhaRepoMock) RemoverRefreshTokensAtivos(ctx context.Context, id int32) error {
	return mock.removerFn(ctx, id)
}

func TestTrocaSenha_CenariosDeValidacao(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("SenhaAtual@123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, senhaAtual, novaSenha string
		wantErr                     error
	}{
		{"senha atual incorreta", "Incorreta@123", "SenhaNova@123", ErrSenhaAtualIncorreta},
		{"senha nova fraca", "SenhaAtual@123", "fraca", ErrSenhaFraca},
		{"senha nova igual", "SenhaAtual@123", "SenhaAtual@123", ErrNovaSenhaIgualAtual},
	} {
		t.Run(tc.name, func(t *testing.T) {
			atualizarChamado := false
			svc := NewTrocaSenhaService(trocaSenhaRepoMock{
				buscarFn: func(context.Context, int32) (*repository.CredenciaisUsuario, error) {
					return &repository.CredenciaisUsuario{Usuario: repository.Usuario{ID: 42}, SenhaHash: string(hash)}, nil
				},
				atualizarFn: func(context.Context, int32, string) error { atualizarChamado = true; return nil },
				listarFn:    func(context.Context, int32) ([]repository.RefreshTokenAtivo, error) { return nil, nil },
				removerFn:   func(context.Context, int32) error { return nil },
			}, trocaSenhaRepoMock{
				listarFn:  func(context.Context, int32) ([]repository.RefreshTokenAtivo, error) { return nil, nil },
				removerFn: func(context.Context, int32) error { return nil },
			}, tokenRevogadoRepoMock{})
			err := svc.Trocar(context.Background(), "42", tc.senhaAtual, tc.novaSenha)
			if !errors.Is(err, tc.wantErr) || atualizarChamado {
				t.Fatalf("erro = %v, atualizacao chamada = %v", err, atualizarChamado)
			}
		})
	}
}

func TestTrocaSenha_SucessoAtualizaHashERevogaSessoes(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("SenhaAtual@123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	var hashNovo string
	var revogacoes []string
	removido := false
	repo := trocaSenhaRepoMock{
		buscarFn: func(context.Context, int32) (*repository.CredenciaisUsuario, error) {
			return &repository.CredenciaisUsuario{Usuario: repository.Usuario{ID: 42}, SenhaHash: string(hash)}, nil
		},
		atualizarFn: func(_ context.Context, id int32, senhaHash string) error {
			if id != 42 {
				t.Fatal("usuario incorreto")
			}
			hashNovo = senhaHash
			return nil
		},
		listarFn: func(context.Context, int32) ([]repository.RefreshTokenAtivo, error) {
			return []repository.RefreshTokenAtivo{{JTI: "um", ExpiraEm: time.Now().Add(time.Hour)}, {JTI: "dois", ExpiraEm: time.Now().Add(2 * time.Hour)}}, nil
		},
		removerFn: func(_ context.Context, id int32) error { removido = id == 42; return nil },
	}
	svc := NewTrocaSenhaService(repo, repo, tokenRevogadoRepoMock{revogar: func(_ context.Context, jti string, _ time.Time) error {
		revogacoes = append(revogacoes, jti)
		return nil
	}})
	if err := svc.Trocar(context.Background(), "42", "SenhaAtual@123", "SenhaNova@123"); err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hashNovo), []byte("SenhaNova@123")) != nil || len(revogacoes) != 2 || !removido {
		t.Fatal("senha ou sessoes nao foram atualizadas corretamente")
	}
}
