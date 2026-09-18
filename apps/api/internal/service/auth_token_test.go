package service

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestAuthToken_ClaimsInvalidas(t *testing.T) {
	secret := uuid.NewString()
	tokens, err := NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*AuthClaims)
		method jwt.SigningMethod
	}{
		{"sem expiracao", func(c *AuthClaims) { c.ExpiresAt = nil }, jwt.SigningMethodHS256},
		{"sem emissao", func(c *AuthClaims) { c.IssuedAt = nil }, jwt.SigningMethodHS256},
		{"emissao futura", func(c *AuthClaims) { c.IssuedAt = jwt.NewNumericDate(time.Now().Add(time.Hour)) }, jwt.SigningMethodHS256},
		{"emissor incorreto", func(c *AuthClaims) { c.Issuer = "outro" }, jwt.SigningMethodHS256},
		{"audiencia incorreta", func(c *AuthClaims) { c.Audience = jwt.ClaimStrings{"refresh"} }, jwt.SigningMethodHS256},
		{"tipo incorreto", func(c *AuthClaims) { c.Tipo = "refresh" }, jwt.SigningMethodHS256},
		{"subject invalido", func(c *AuthClaims) { c.Subject = "abc" }, jwt.SigningMethodHS256},
		{"subject zero", func(c *AuthClaims) { c.Subject = "0" }, jwt.SigningMethodHS256},
		{"sem jti", func(c *AuthClaims) { c.ID = "" }, jwt.SigningMethodHS256},
		{"algoritmo incorreto", func(c *AuthClaims) {}, jwt.SigningMethodHS384},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := AuthClaims{Tipo: "access", Idioma: "pt-BR", RegisteredClaims: jwt.RegisteredClaims{
				Subject: "1", Issuer: "memory-card", Audience: jwt.ClaimStrings{"access"}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)), ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			}}
			tc.mutate(&claims)
			raw, err := jwt.NewWithClaims(tc.method, claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tokens.ValidarAccess(raw); !errors.Is(err, ErrTokenInvalido) {
				t.Fatalf("esperado token invalido: %v", err)
			}
		})
	}
}

func TestNewAuthToken_ConfigInvalida(t *testing.T) {
	for _, tc := range []struct {
		secret          string
		access, refresh time.Duration
	}{
		{"", time.Minute, time.Hour}, {uuid.NewString(), 0, time.Hour}, {uuid.NewString(), time.Hour, time.Minute},
	} {
		if _, err := NewAuthToken(tc.secret, tc.access, tc.refresh); err == nil {
			t.Fatal("config invalida aceita")
		}
	}
}
