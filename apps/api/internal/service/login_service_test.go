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
	buscar func(context.Context, string) (*repository.CredenciaisUsuario, error)
}

func (mock loginRepoMock) BuscarPorEmail(ctx context.Context, email string) (*repository.CredenciaisUsuario, error) {
	return mock.buscar(ctx, email)
}

func setupLogin(t *testing.T, repo LoginRepository) (*LoginService, *AuthToken) {
	t.Helper()
	tokens, err := NewAuthToken(uuid.NewString(), 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewLoginService(repo, tokens, tokenRevogadoRepoMock{consultar: func(context.Context, string) (bool, error) {
		return false, nil
	}})
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
	svc, tokens := setupLogin(t, nil)
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
