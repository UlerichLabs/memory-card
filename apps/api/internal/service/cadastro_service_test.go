package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type mockUsuarioRepo struct {
	existePorEmailFn func(ctx context.Context, email string) (bool, error)
	criarFn          func(ctx context.Context, nome, email, senhaHash string) (*repository.Usuario, error)
	criarChamado     bool
}

func (m *mockUsuarioRepo) ExistePorEmail(ctx context.Context, email string) (bool, error) {
	if m.existePorEmailFn != nil {
		return m.existePorEmailFn(ctx, email)
	}
	return false, nil
}

func (m *mockUsuarioRepo) Criar(ctx context.Context, nome, email, senhaHash string) (*repository.Usuario, error) {
	m.criarChamado = true
	if m.criarFn != nil {
		return m.criarFn(ctx, nome, email, senhaHash)
	}
	return nil, errors.New("nao implementado")
}

func TestCadastroService_Cadastrar_Sucesso(t *testing.T) {
	mockRepo := &mockUsuarioRepo{
		existePorEmailFn: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		criarFn: func(ctx context.Context, nome, email, senhaHash string) (*repository.Usuario, error) {
			if senhaHash == "SenhaForte@123" {
				t.Fatal("senha nao pode ser salva em texto puro")
			}
			return &repository.Usuario{
				ID:        1,
				Nome:      nome,
				Email:     email,
				Idioma:    "pt-BR",
				CreatedAt: time.Now(),
			}, nil
		},
	}

	svc := NewCadastroService(mockRepo)
	u, err := svc.Cadastrar(context.Background(), "Lucas", "lucas@example.com", "SenhaForte@123")
	if err != nil {
		t.Fatalf("esperava sucesso, obteve erro: %v", err)
	}
	if u == nil || u.ID != 1 || u.Email != "lucas@example.com" {
		t.Fatalf("usuario retornado invalido: %+v", u)
	}
	if !mockRepo.criarChamado {
		t.Fatal("esperava que repo.Criar tivesse sido chamado")
	}
}

func TestCadastroService_Cadastrar_EmailJaCadastrado(t *testing.T) {
	mockRepo := &mockUsuarioRepo{
		existePorEmailFn: func(ctx context.Context, email string) (bool, error) {
			return true, nil
		},
	}

	svc := NewCadastroService(mockRepo)
	_, err := svc.Cadastrar(context.Background(), "Lucas", "existente@example.com", "SenhaForte@123")
	if !errors.Is(err, ErrEmailJaCadastrado) {
		t.Fatalf("esperava ErrEmailJaCadastrado, obteve: %v", err)
	}
	if mockRepo.criarChamado {
		t.Fatal("repo.Criar nao deveria ter sido chamado quando email ja existe")
	}
}

func TestCadastroService_Cadastrar_SenhaFraca(t *testing.T) {
	tests := []struct {
		name  string
		senha string
	}{
		{
			name:  "menos de 8 caracteres",
			senha: "Ab1!xyz",
		},
		{
			name:  "sem maiuscula",
			senha: "senhaforte@123",
		},
		{
			name:  "sem numero",
			senha: "SenhaForte@abc",
		},
		{
			name:  "sem caractere especial",
			senha: "SenhaForte1234",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockUsuarioRepo{}
			svc := NewCadastroService(mockRepo)

			_, err := svc.Cadastrar(context.Background(), "Lucas", "lucas@example.com", tc.senha)
			if !errors.Is(err, ErrSenhaFraca) {
				t.Fatalf("esperava ErrSenhaFraca para senha %q, obteve: %v", tc.senha, err)
			}
			if mockRepo.criarChamado {
				t.Fatal("repo.Criar nao deveria ter sido chamado para senha fraca")
			}
		})
	}
}

func TestCadastroService_Cadastrar_EmailInvalido(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "vazio",
			email: "",
		},
		{
			name:  "sem arroba",
			email: "lucasexample.com",
		},
		{
			name:  "sem dominio",
			email: "lucas@",
		},
		{
			name:  "sem tld",
			email: "lucas@example",
		},
		{
			name:  "com espaco",
			email: "lucas @example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockUsuarioRepo{}
			svc := NewCadastroService(mockRepo)

			_, err := svc.Cadastrar(context.Background(), "Lucas", tc.email, "SenhaForte@123")
			if !errors.Is(err, ErrEmailInvalido) {
				t.Fatalf("esperava ErrEmailInvalido para email %q, obteve: %v", tc.email, err)
			}
			if mockRepo.criarChamado {
				t.Fatal("repo.Criar nao deveria ter sido chamado para email invalido")
			}
		})
	}
}
