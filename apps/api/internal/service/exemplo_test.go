package service_test

import (
	"context"
	"errors"
	"testing"
	"time"
)

/*
PADRÃO DE TESTE DE SERVICE COM REPOSITORY MOCKADO (MEMOR-19 / MEMOR-27)
========================================================================

Este arquivo define o padrão canônico para testes unitários da camada service no Memory Card.
Quando implementar novos services de domínio (ex: auth, usuarios, coleções, jogos):

1. INTERFACE DE REPOSITORY DEFINIDA PELO CONSUMIDOR:
   - O service define a interface mínima necessária para suas operações ("Go: aceite interfaces, retorne structs").
   - A interface deve ser pequena e coesa (Interface Segregation Principle).
   - Exemplo:
       type UsuarioRepository interface {
           BuscarPorEmail(ctx context.Context, email string) (*Usuario, error)
           Criar(ctx context.Context, u *Usuario) error
       }

2. PADRÃO DE MOCK ESCOLHIDO: MOCK MANUAL (STRUCT COM CAMPOS DE FUNÇÃO)
   - Decisão: mocks manuais são a convenção preferencial do projeto.
   - Motivos:
     a) Zero dependências externas de binários no CI/desenvolvimento (sem necessidade de mockgen/moq).
     b) Código 100% idiomático Go, legível e rastreável em code reviews de portfólio.
     c) Facilidade de configurar comportamentos diferentes por teste usando closures simples.
     d) Se futuramente uma interface se tornar muito extensa, ferramentas de geração como `moq`
        podem ser avaliadas, mantendo a mesma assinatura.

3. TESTES ISOLADOS (SEM BANCO REAL):
   - A camada service NUNCA instancia conexão de banco ou importa driver SQL diretamente.
   - Testes unitários validam lógica de negócio, transformações e mapeamento de erros de domínio.
   - Testes que necessitam de banco de dados real pertencem à camada de integração/e2e
     (testcontainers-go, previsto para a história MEMOR-20).

4. NOMENCLATURA E ESTRUTURA:
   - Nome das funções de teste: Test<Funcao>_<Cenario> (ex: TestExemploUsuarioService_BuscarPorEmail_Sucesso)
   - Table-driven tests quando houver mais de 2 variações do mesmo cenário.
*/

// Erros de domínio tipados para o exemplo.
var (
	errUsuarioNaoEncontrado = errors.New("usuário não encontrado")
	errRepositorioFalhou    = errors.New("falha no repositório")
)

// ExemploUsuario representa a entidade de domínio do exemplo.
type ExemploUsuario struct {
	ID        int64
	Nome      string
	Email     string
	CreatedAt time.Time
}

// ExemploUsuarioRepository é a interface que o service consome (inversão de dependência).
type ExemploUsuarioRepository interface {
	BuscarPorEmail(ctx context.Context, email string) (*ExemploUsuario, error)
}

// ExemploUsuarioService contém a regra de negócio para operações com usuário.
type ExemploUsuarioService struct {
	repo ExemploUsuarioRepository
}

// NewExemploUsuarioService instancia o service com sua dependência injetada.
func NewExemploUsuarioService(repo ExemploUsuarioRepository) *ExemploUsuarioService {
	return &ExemploUsuarioService{repo: repo}
}

// ObterPerfilPorEmail busca um usuário e aplica validações de domínio.
func (s *ExemploUsuarioService) ObterPerfilPorEmail(ctx context.Context, email string) (*ExemploUsuario, error) {
	if email == "" {
		return nil, errors.New("email não pode ser vazio")
	}

	usuario, err := s.repo.BuscarPorEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return usuario, nil
}

// mockExemploUsuarioRepository é o mock manual da interface ExemploUsuarioRepository.
type mockExemploUsuarioRepository struct {
	buscarPorEmailFn func(ctx context.Context, email string) (*ExemploUsuario, error)
}

func (m *mockExemploUsuarioRepository) BuscarPorEmail(ctx context.Context, email string) (*ExemploUsuario, error) {
	if m.buscarPorEmailFn != nil {
		return m.buscarPorEmailFn(ctx, email)
	}
	return nil, errors.New("BuscarPorEmail não implementado no mock")
}

// --- Testes Unitários Demonstrando o Padrão ---

func TestExemploUsuarioService_ObterPerfilPorEmail(t *testing.T) {
	usuarioPadrao := &ExemploUsuario{
		ID:        1,
		Nome:      "Jogador Exemplo",
		Email:     "jogador@memorycard.app",
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name        string
		email       string
		mockFn      func(ctx context.Context, email string) (*ExemploUsuario, error)
		wantUsuario *ExemploUsuario
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "busca com sucesso retorna usuario",
			email: "jogador@memorycard.app",
			mockFn: func(ctx context.Context, email string) (*ExemploUsuario, error) {
				return usuarioPadrao, nil
			},
			wantUsuario: usuarioPadrao,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:  "email vazio retorna erro de validacao sem chamar repository",
			email: "",
			mockFn: func(ctx context.Context, email string) (*ExemploUsuario, error) {
				t.Fatal("repository não deveria ter sido invocado para email vazio")
				return nil, nil
			},
			wantUsuario: nil,
			wantErr:     true,
			expectedErr: nil,
		},
		{
			name:  "usuario nao encontrado no repository propaga erro",
			email: "inexistente@memorycard.app",
			mockFn: func(ctx context.Context, email string) (*ExemploUsuario, error) {
				return nil, errUsuarioNaoEncontrado
			},
			wantUsuario: nil,
			wantErr:     true,
			expectedErr: errUsuarioNaoEncontrado,
		},
		{
			name:  "falha de infraestrutura no repository propaga erro",
			email: "qualquer@memorycard.app",
			mockFn: func(ctx context.Context, email string) (*ExemploUsuario, error) {
				return nil, errRepositorioFalhou
			},
			wantUsuario: nil,
			wantErr:     true,
			expectedErr: errRepositorioFalhou,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockExemploUsuarioRepository{
				buscarPorEmailFn: tc.mockFn,
			}
			service := NewExemploUsuarioService(mockRepo)

			usuario, err := service.ObterPerfilPorEmail(context.Background(), tc.email)

			if (err != nil) != tc.wantErr {
				t.Fatalf("ObterPerfilPorEmail() erro = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
				t.Fatalf("ObterPerfilPorEmail() erro = %v, esperava %v", err, tc.expectedErr)
			}
			if tc.wantUsuario != nil && usuario != tc.wantUsuario {
				t.Errorf("ObterPerfilPorEmail() usuario = %+v, esperava %+v", usuario, tc.wantUsuario)
			}
		})
	}
}
