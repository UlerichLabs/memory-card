package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type tokenRevogadoRepoMock struct {
	revogar   func(context.Context, string, time.Time) error
	consultar func(context.Context, string) (bool, error)
	registrar func(context.Context, string, int32, time.Time) error
}

func (mock tokenRevogadoRepoMock) Revogar(ctx context.Context, jti string, expiraEm time.Time) error {
	return mock.revogar(ctx, jti, expiraEm)
}

func (mock tokenRevogadoRepoMock) EstaRevogado(ctx context.Context, jti string) (bool, error) {
	return mock.consultar(ctx, jti)
}

func (mock tokenRevogadoRepoMock) RegistrarRefreshToken(ctx context.Context, jti string, usuarioID int32, expiraEm time.Time) error {
	if mock.registrar == nil {
		return nil
	}
	return mock.registrar(ctx, jti, usuarioID, expiraEm)
}

func TestLogout_Cenarios(t *testing.T) {
	tokens, err := NewAuthToken(uuid.NewString(), time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := tokens.emitir("42", "en", "refresh", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.validar(refresh, "refresh")
	if err != nil {
		t.Fatal(err)
	}
	expirado, err := tokens.emitir("42", "en", "refresh", -time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	access, err := tokens.emitir("42", "en", "access", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	outrosTokens, err := NewAuthToken(uuid.NewString(), time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	assinaturaInvalida, err := outrosTokens.emitir("42", "en", "refresh", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name, subject, raw string
		repoErr, wantErr   error
		wantCall           bool
	}{
		{"sucesso", "42", refresh, nil, nil, true},
		{"outro usuario", "43", refresh, nil, ErrTokenInvalido, false},
		{"sem identidade", "", refresh, nil, ErrTokenInvalido, false},
		{"expirado", "42", expirado, nil, ErrSessaoExpirada, false},
		{"malformado", "42", "invalido", nil, ErrSessaoExpirada, false},
		{"ausente", "42", "", nil, ErrSessaoExpirada, false},
		{"access como refresh", "42", access, nil, ErrSessaoExpirada, false},
		{"assinatura invalida", "42", assinaturaInvalida, nil, ErrSessaoExpirada, false},
		{"falha no banco", "42", refresh, dbErr, dbErr, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			called := false
			repo := tokenRevogadoRepoMock{revogar: func(got context.Context, jti string, expiraEm time.Time) error {
				called = true
				if got != ctx || jti != claims.ID || !expiraEm.Equal(claims.ExpiresAt.Time) {
					t.Fatal("revogacao nao recebeu contexto, jti e expiracao do refresh")
				}
				return tc.repoErr
			}}
			err := NewLogoutService(tokens, repo).Logout(ctx, tc.subject, tc.raw)
			if !errors.Is(err, tc.wantErr) || called != tc.wantCall {
				t.Fatalf("erro = %v, repo chamado = %v", err, called)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := NewLogoutService(tokens, tokenRevogadoRepoMock{}).Logout(ctx, "42", refresh); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRefresh_Revogacao(t *testing.T) {
	tokens, err := NewAuthToken(uuid.NewString(), time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := tokens.emitir("42", "en", "refresh", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.validar(refresh, "refresh")
	if err != nil {
		t.Fatal(err)
	}
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name             string
		revogado         bool
		repoErr, wantErr error
	}{
		{"revogado com assinatura e TTL validos", true, nil, ErrSessaoExpirada},
		{"nao revogado", false, nil, nil},
		{"falha na consulta", false, dbErr, dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			called := false
			repo := tokenRevogadoRepoMock{consultar: func(got context.Context, jti string) (bool, error) {
				called = true
				if got != ctx || jti != claims.ID {
					t.Fatal("contexto/jti incorreto")
				}
				return tc.revogado, tc.repoErr
			}}
			svc, err := NewLoginService(nil, tokens, repo, repo)
			if err != nil {
				t.Fatal(err)
			}
			result, err := svc.Refresh(ctx, refresh)
			if !called || !errors.Is(err, tc.wantErr) {
				t.Fatalf("erro = %v, repo chamado = %v", err, called)
			}
			if tc.wantErr != nil {
				if result != nil {
					t.Fatal("refresh rejeitado nao pode emitir token")
				}
				return
			}
			access, err := tokens.ValidarAccess(result.AccessToken)
			if err != nil || access.Subject != claims.Subject || access.Idioma != claims.Idioma {
				t.Fatalf("access token incorreto: %v", err)
			}
		})
	}
}
