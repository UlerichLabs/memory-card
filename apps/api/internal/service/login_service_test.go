package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type loginRepoMock struct {
	buscar      func(context.Context, string) (*repository.CredenciaisUsuario, error)
	buscarPorID func(context.Context, int32) (*repository.Usuario, error)
}

func (mock loginRepoMock) BuscarPorEmail(ctx context.Context, email string) (*repository.CredenciaisUsuario, error) {
	if mock.buscar == nil {
		return nil, nil
	}
	return mock.buscar(ctx, email)
}

func (mock loginRepoMock) BuscarPorID(ctx context.Context, id int32) (*repository.Usuario, error) {
	if mock.buscarPorID == nil {
		return nil, nil
	}
	return mock.buscarPorID(ctx, id)
}

func setupLogin(t *testing.T, repo LoginRepository) (*LoginService, *AuthToken) {
	t.Helper()
	tokens, err := NewAuthToken(uuid.NewString(), 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewLoginService(repo, tokens, tokenRevogadoRepoMock{consultar: func(context.Context, string) (bool, error) {
		return false, nil
	}}, tokenRevogadoRepoMock{})
	if err != nil {
		t.Fatal(err)
	}
	return svc, tokens
}

func TestLogin_Cenarios(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("SenhaForte@123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	usuario := &repository.CredenciaisUsuario{Usuario: repository.Usuario{ID: 1, Nome: "Lucas", Email: "lucas@example.com", Idioma: "pt-BR"}, SenhaHash: string(hash)}
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name    string
		usuario *repository.CredenciaisUsuario
		senha   string
		repoErr error
		wantErr error
	}{
		{"sucesso", usuario, "SenhaForte@123", nil, nil},
		{"email ausente", nil, "SenhaForte@123", nil, ErrCredenciaisInvalidas},
		{"senha incorreta", usuario, "Incorreta@123", nil, ErrCredenciaisInvalidas},
		{"banco indisponivel", nil, "SenhaForte@123", dbErr, dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc, tokens := setupLogin(t, loginRepoMock{buscar: func(got context.Context, email string) (*repository.CredenciaisUsuario, error) {
				if got != ctx || email != usuario.Email {
					t.Fatal("contexto/email incorretos")
				}
				return tc.usuario, tc.repoErr
			}})
			result, err := svc.Login(ctx, " lucas@example.com ", tc.senha)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("erro = %v; esperado %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}
			if result.Usuario != usuario.Usuario {
				t.Fatal("usuario incorreto")
			}
			access, err := tokens.ValidarAccess(result.AccessToken)
			if err != nil {
				t.Fatal(err)
			}
			refresh, err := tokens.validar(result.RefreshToken, "refresh")
			if err != nil {
				t.Fatal(err)
			}
			if access.Subject != "1" || access.Idioma != "pt-BR" || access.ExpiresAt.Sub(access.IssuedAt.Time) != 15*time.Minute || refresh.ExpiresAt.Sub(refresh.IssuedAt.Time) != 168*time.Hour {
				t.Fatal("claims incorretas")
			}
		})
	}
}

func TestRefresh_Cenarios(t *testing.T) {
	repo := loginRepoMock{buscarPorID: func(ctx context.Context, id int32) (*repository.Usuario, error) {
		return &repository.Usuario{ID: id, Idioma: "en"}, nil
	}}
	svc, tokens := setupLogin(t, repo)
	now := time.Now().Truncate(time.Second)
	tokens.now = func() time.Time { return now }
	refresh, err := tokens.emitir("1", "en", "refresh", tokens.refreshTTL)
	if err != nil {
		t.Fatal(err)
	}
	access, err := tokens.emitir("1", "en", "access", tokens.accessTTL)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, raw string
		advance   time.Duration
		valid     bool
	}{
		{"sucesso", refresh, 0, true},
		{"expirado", refresh, 169 * time.Hour, false},
		{"malformado", "invalido", 0, false},
		{"ausente", "", 0, false},
		{"access como refresh", access, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tokens.now = func() time.Time { return now.Add(tc.advance) }
			result, err := svc.Refresh(context.Background(), tc.raw)
			if !tc.valid {
				if !errors.Is(err, ErrSessaoExpirada) {
					t.Fatalf("erro = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			claims, err := tokens.ValidarAccess(result.AccessToken)
			if err != nil || claims.Subject != "1" || claims.Idioma != "en" || result.AccessToken == access {
				t.Fatalf("refresh incorreto: %v", err)
			}
		})
	}
}

func TestRefresh_IdiomaDoBanco(t *testing.T) {
	repo := loginRepoMock{buscarPorID: func(ctx context.Context, id int32) (*repository.Usuario, error) {
		return &repository.Usuario{ID: 1, Idioma: "en"}, nil
	}}
	svc, tokens := setupLogin(t, repo)
	refresh, err := tokens.emitir("1", "pt-BR", "refresh", tokens.refreshTTL)
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.ValidarAccess(result.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Idioma != "en" {
		t.Fatalf("idioma esperado en, obtido %s", claims.Idioma)
	}
}

func TestRefresh_UsuarioInexistente(t *testing.T) {
	repo := loginRepoMock{buscarPorID: func(ctx context.Context, id int32) (*repository.Usuario, error) {
		return nil, nil
	}}
	svc, tokens := setupLogin(t, repo)
	refresh, err := tokens.emitir("1", "pt-BR", "refresh", tokens.refreshTTL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(context.Background(), refresh)
	if !errors.Is(err, ErrSessaoExpirada) {
		t.Fatalf("esperado ErrSessaoExpirada, obtido %v", err)
	}
}

func TestRefresh_SubjectInvalido(t *testing.T) {
	repo := loginRepoMock{buscarPorID: func(ctx context.Context, id int32) (*repository.Usuario, error) {
		t.Fatal("repository nao deveria ser chamado para subject invalido")
		return nil, nil
	}}
	svc, tokens := setupLogin(t, repo)
	for _, subject := range []string{"invalido", "0", "-1", "99999999999999999"} {
		token, err := tokens.emitir(subject, "pt-BR", "refresh", tokens.refreshTTL)
		if err != nil {
			t.Fatal(err)
		}
		_, err = svc.Refresh(context.Background(), token)
		if !errors.Is(err, ErrSessaoExpirada) {
			t.Fatalf("subject %s: esperado ErrSessaoExpirada, obtido %v", subject, err)
		}
	}
}

func TestRefresh_ErroBanco(t *testing.T) {
	dbErr := errors.New("database connection failed")
	repo := loginRepoMock{buscarPorID: func(ctx context.Context, id int32) (*repository.Usuario, error) {
		return nil, dbErr
	}}
	svc, tokens := setupLogin(t, repo)
	refresh, err := tokens.emitir("1", "pt-BR", "refresh", tokens.refreshTTL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(context.Background(), refresh)
	if errors.Is(err, ErrSessaoExpirada) {
		t.Fatal("erro de banco nao deve se transformar em ErrSessaoExpirada")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("erro de banco nao propagado: %v", err)
	}
}

func TestRefresh_RevogadoEExpiradoSemChamarBanco(t *testing.T) {
	tokens, err := NewAuthToken(uuid.NewString(), 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	repo := loginRepoMock{buscarPorID: func(ctx context.Context, id int32) (*repository.Usuario, error) {
		t.Fatal("repository nao deveria ser chamado")
		return nil, nil
	}}
	revogados := tokenRevogadoRepoMock{consultar: func(ctx context.Context, jti string) (bool, error) {
		return true, nil
	}}
	svc, err := NewLoginService(repo, tokens, revogados, revogados)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := tokens.emitir("1", "pt-BR", "refresh", tokens.refreshTTL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(context.Background(), refresh)
	if !errors.Is(err, ErrSessaoExpirada) {
		t.Fatalf("esperado ErrSessaoExpirada para token revogado, obtido %v", err)
	}

	expirado, err := tokens.emitir("1", "pt-BR", "refresh", -time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Refresh(context.Background(), expirado)
	if !errors.Is(err, ErrSessaoExpirada) {
		t.Fatalf("esperado ErrSessaoExpirada para token expirado, obtido %v", err)
	}
}

func TestLogin_InputInvalido(t *testing.T) {
	svc, _ := setupLogin(t, loginRepoMock{buscar: func(context.Context, string) (*repository.CredenciaisUsuario, error) {
		t.Fatal("nao deveria consultar repositorio")
		return nil, nil
	}})
	for _, tc := range []struct{ email, senha string }{
		{"invalido", "Senha@123"}, {"lucas@example.com", ""}, {"lucas@example.com", strings.Repeat("a", 73)},
	} {
		_, err := svc.Login(context.Background(), tc.email, tc.senha)
		if !errors.Is(err, ErrCredenciaisInvalidas) {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.Login(ctx, "lucas@example.com", "Senha@123"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(ctx, "token"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
