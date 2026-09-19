package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type recuperacaoRepoMock struct {
	registrarSolicitacaoFn func(context.Context, string) (bool, error)
	criarTokenFn           func(context.Context, int32, string, time.Time) error
	buscarTokenFn          func(context.Context, string) (*repository.TokenResetSenha, error)
	consumirFn             func(context.Context, int32, string) (bool, error)
	listarRefreshFn        func(context.Context, int32) ([]repository.RefreshTokenAtivo, error)
	removerRefreshFn       func(context.Context, int32) error
}

func (mock *recuperacaoRepoMock) RegistrarSolicitacao(ctx context.Context, email string) (bool, error) {
	return mock.registrarSolicitacaoFn(ctx, email)
}

func (mock *recuperacaoRepoMock) CriarToken(ctx context.Context, usuarioID int32, tokenHash string, expiraEm time.Time) error {
	return mock.criarTokenFn(ctx, usuarioID, tokenHash, expiraEm)
}

func (mock *recuperacaoRepoMock) BuscarToken(ctx context.Context, tokenHash string) (*repository.TokenResetSenha, error) {
	return mock.buscarTokenFn(ctx, tokenHash)
}

func (mock *recuperacaoRepoMock) ConsumirETrocarSenha(ctx context.Context, tokenID int32, senhaHash string) (bool, error) {
	return mock.consumirFn(ctx, tokenID, senhaHash)
}

func (mock *recuperacaoRepoMock) ListarRefreshTokensAtivos(ctx context.Context, usuarioID int32) ([]repository.RefreshTokenAtivo, error) {
	return mock.listarRefreshFn(ctx, usuarioID)
}

func (mock *recuperacaoRepoMock) RemoverRefreshTokensAtivos(ctx context.Context, usuarioID int32) error {
	return mock.removerRefreshFn(ctx, usuarioID)
}

type recuperacaoUsuarioRepoMock struct {
	buscarFn func(context.Context, string) (*repository.CredenciaisUsuario, error)
}

func (mock recuperacaoUsuarioRepoMock) BuscarPorEmail(ctx context.Context, email string) (*repository.CredenciaisUsuario, error) {
	return mock.buscarFn(ctx, email)
}

type emailSenderMock struct {
	enviarFn func(context.Context, string, string) error
}

func (mock emailSenderMock) EnviarRecuperacaoSenha(ctx context.Context, destinatario, link string) error {
	return mock.enviarFn(ctx, destinatario, link)
}

func novaRecuperacaoSenhaService(now time.Time, usuarios RecuperacaoUsuarioRepository, repo RecuperacaoSenhaRepository, revogados TokenRevogadoRepository, email EmailSender) *RecuperacaoSenhaService {
	svc := NewRecuperacaoSenhaService("https://app.example.com/redefinir-senha", usuarios, repo, revogados, email)
	svc.now = func() time.Time { return now }
	return svc
}

func TestRecuperacaoSenha_Solicitar_GenericoETokenSeguro(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	usuario := &repository.CredenciaisUsuario{Usuario: repository.Usuario{ID: 42, Email: "usuario@example.com"}}
	var hashArmazenado string
	var tokenEmail string
	repo := &recuperacaoRepoMock{
		registrarSolicitacaoFn: func(ctx context.Context, email string) (bool, error) {
			if email != "usuario@example.com" {
				t.Fatal("email do limite incorreto")
			}
			return true, nil
		},
		criarTokenFn: func(ctx context.Context, usuarioID int32, tokenHash string, expiraEm time.Time) error {
			if usuarioID != usuario.ID || !expiraEm.Equal(now.Add(30*time.Minute)) || len(tokenHash) != 64 {
				t.Fatal("token persistido incorretamente")
			}
			hashArmazenado = tokenHash
			return nil
		},
	}
	svc := novaRecuperacaoSenhaService(now, recuperacaoUsuarioRepoMock{buscarFn: func(ctx context.Context, email string) (*repository.CredenciaisUsuario, error) {
		return usuario, nil
	}}, repo, tokenRevogadoRepoMock{}, emailSenderMock{enviarFn: func(ctx context.Context, destinatario, link string) error {
		if destinatario != usuario.Email {
			t.Fatal("destinatario incorreto")
		}
		parsed, err := url.Parse(link)
		if err != nil {
			t.Fatal(err)
		}
		tokenEmail = parsed.Query().Get("token")
		return nil
	}})
	if err := svc.Solicitar(context.Background(), " usuario@example.com "); err != nil {
		t.Fatal(err)
	}
	if tokenEmail == "" || tokenEmail == hashArmazenado || hashToken(tokenEmail) != hashArmazenado {
		t.Fatal("token puro foi persistido ou nao corresponde ao hash")
	}

	naoEncontradoChamouEmail := false
	svc = novaRecuperacaoSenhaService(now, recuperacaoUsuarioRepoMock{buscarFn: func(context.Context, string) (*repository.CredenciaisUsuario, error) {
		return nil, nil
	}}, &recuperacaoRepoMock{registrarSolicitacaoFn: func(context.Context, string) (bool, error) { return true, nil }, criarTokenFn: func(context.Context, int32, string, time.Time) error {
		t.Fatal("nao deve criar token para email inexistente")
		return nil
	}}, tokenRevogadoRepoMock{}, emailSenderMock{enviarFn: func(context.Context, string, string) error {
		naoEncontradoChamouEmail = true
		return nil
	}})
	if err := svc.Solicitar(context.Background(), "ausente@example.com"); err != nil || naoEncontradoChamouEmail {
		t.Fatalf("resposta para email inexistente incorreta: %v", err)
	}
}

func TestRecuperacaoSenha_Solicitar_Limite(t *testing.T) {
	svc := novaRecuperacaoSenhaService(time.Now(), recuperacaoUsuarioRepoMock{}, &recuperacaoRepoMock{registrarSolicitacaoFn: func(context.Context, string) (bool, error) {
		return false, nil
	}}, tokenRevogadoRepoMock{}, emailSenderMock{})
	if err := svc.Solicitar(context.Background(), "usuario@example.com"); !errors.Is(err, ErrLimiteSolicitacoesReset) {
		t.Fatal(err)
	}
}

func TestRecuperacaoSenha_ValidarToken_Cenarios(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	raw := "token-valido"
	for _, tc := range []struct {
		name  string
		token *repository.TokenResetSenha
		err   error
	}{
		{"valido", &repository.TokenResetSenha{ID: 1, ExpiraEm: now.Add(time.Minute)}, nil},
		{"expirado", &repository.TokenResetSenha{ID: 1, ExpiraEm: now}, ErrTokenResetExpirado},
		{"utilizado", &repository.TokenResetSenha{ID: 1, ExpiraEm: now.Add(time.Minute), UsadoEm: &now}, ErrTokenResetUtilizado},
		{"inexistente", nil, ErrTokenResetInvalido},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := novaRecuperacaoSenhaService(now, recuperacaoUsuarioRepoMock{}, &recuperacaoRepoMock{buscarTokenFn: func(ctx context.Context, hash string) (*repository.TokenResetSenha, error) {
				if hash != hashToken(raw) {
					t.Fatal("token deve ser consultado pelo hash")
				}
				return tc.token, nil
			}}, tokenRevogadoRepoMock{}, emailSenderMock{})
			if err := svc.ValidarToken(context.Background(), raw); !errors.Is(err, tc.err) {
				t.Fatalf("erro = %v, esperado %v", err, tc.err)
			}
		})
	}
}

func TestRecuperacaoSenha_RedefinirSenha_SucessoRevogaSessoes(t *testing.T) {
	now := time.Now()
	raw := "token-valido"
	token := &repository.TokenResetSenha{ID: 5, UsuarioID: 42, ExpiraEm: now.Add(time.Minute)}
	revogacoes := make(map[string]time.Time)
	removido := false
	svc := novaRecuperacaoSenhaService(now, recuperacaoUsuarioRepoMock{}, &recuperacaoRepoMock{
		buscarTokenFn: func(context.Context, string) (*repository.TokenResetSenha, error) { return token, nil },
		consumirFn: func(ctx context.Context, tokenID int32, senhaHash string) (bool, error) {
			if tokenID != token.ID || strings.Contains(senhaHash, "SenhaNova@123") || bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte("SenhaNova@123")) != nil {
				t.Fatal("senha nao foi hasheada corretamente")
			}
			return true, nil
		},
		listarRefreshFn: func(ctx context.Context, usuarioID int32) ([]repository.RefreshTokenAtivo, error) {
			return []repository.RefreshTokenAtivo{{JTI: "primeiro", UsuarioID: usuarioID, ExpiraEm: now.Add(time.Hour)}, {JTI: "segundo", UsuarioID: usuarioID, ExpiraEm: now.Add(2 * time.Hour)}}, nil
		},
		removerRefreshFn: func(ctx context.Context, usuarioID int32) error {
			removido = usuarioID == token.UsuarioID
			return nil
		},
	}, tokenRevogadoRepoMock{revogar: func(ctx context.Context, jti string, expiraEm time.Time) error {
		revogacoes[jti] = expiraEm
		return nil
	}}, emailSenderMock{})
	if err := svc.RedefinirSenha(context.Background(), raw, "SenhaNova@123"); err != nil {
		t.Fatal(err)
	}
	if len(revogacoes) != 2 || !removido {
		t.Fatal("refresh tokens ativos nao foram todos invalidados")
	}
}

func TestRecuperacaoSenha_RedefinirSenha_Erros(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name, senha string
		token       *repository.TokenResetSenha
		consumido   bool
		wantErr     error
	}{
		{"expirado", "SenhaNova@123", &repository.TokenResetSenha{ID: 1, ExpiraEm: now.Add(-time.Minute)}, false, ErrTokenResetExpirado},
		{"utilizado", "SenhaNova@123", &repository.TokenResetSenha{ID: 1, ExpiraEm: now.Add(time.Minute), UsadoEm: &now}, false, ErrTokenResetUtilizado},
		{"senha fraca", "fraca", &repository.TokenResetSenha{ID: 1, ExpiraEm: now.Add(time.Minute)}, false, ErrSenhaFraca},
		{"concorrente utilizou", "SenhaNova@123", &repository.TokenResetSenha{ID: 1, ExpiraEm: now.Add(time.Minute)}, false, ErrTokenResetUtilizado},
	} {
		t.Run(tc.name, func(t *testing.T) {
			consumirChamado := false
			svc := novaRecuperacaoSenhaService(now, recuperacaoUsuarioRepoMock{}, &recuperacaoRepoMock{
				buscarTokenFn: func(context.Context, string) (*repository.TokenResetSenha, error) { return tc.token, nil },
				consumirFn:    func(context.Context, int32, string) (bool, error) { consumirChamado = true; return tc.consumido, nil },
			}, tokenRevogadoRepoMock{}, emailSenderMock{})
			err := svc.RedefinirSenha(context.Background(), "token", tc.senha)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("erro = %v, esperado %v", err, tc.wantErr)
			}
			if tc.wantErr == ErrSenhaFraca && consumirChamado {
				t.Fatal("senha fraca nao pode consumir token")
			}
		})
	}
}
