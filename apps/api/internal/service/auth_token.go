package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrTokenInvalido = errors.New("auth.session.unauthorized")

type AuthClaims struct {
	Idioma string `json:"idioma"`
	Tipo   string `json:"tipo"`
	jwt.RegisteredClaims
}

type AuthToken struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func NewAuthToken(secret string, accessTTL, refreshTTL time.Duration) (*AuthToken, error) {
	if len(secret) < 32 || accessTTL < time.Second || refreshTTL <= accessTTL {
		return nil, fmt.Errorf("configuracao JWT invalida")
	}
	return &AuthToken{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL, now: time.Now}, nil
}

func (tokens *AuthToken) emitir(subject, idioma, tipo string, ttl time.Duration) (string, error) {
	now := tokens.now()
	claims := AuthClaims{Idioma: idioma, Tipo: tipo, RegisteredClaims: jwt.RegisteredClaims{
		Subject: subject, Issuer: "memory-card", Audience: jwt.ClaimStrings{tipo},
		ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tokens.secret)
	if err != nil {
		return "", fmt.Errorf("assinar token: %w", err)
	}
	return token, nil
}

func (tokens *AuthToken) validar(raw, tipo string) (*AuthClaims, error) {
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return tokens.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithIssuer("memory-card"), jwt.WithAudience(tipo), jwt.WithTimeFunc(tokens.now))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenInvalido, err)
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 32)
	if err != nil || id <= 0 || !token.Valid || claims.Tipo != tipo || claims.ID == "" || claims.IssuedAt == nil {
		return nil, ErrTokenInvalido
	}
	return claims, nil
}

func (tokens *AuthToken) ValidarAccess(raw string) (*AuthClaims, error) {
	return tokens.validar(raw, "access")
}
