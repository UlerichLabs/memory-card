package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCredenciaisInvalidas = errors.New("auth.login.invalid_credentials")
	ErrSessaoExpirada       = errors.New("auth.session.expired")
)

type LoginRepository interface {
	BuscarPorEmail(ctx context.Context, email string) (*repository.CredenciaisUsuario, error)
}

type LoginService struct {
	repo      LoginRepository
	tokens    *AuthToken
	dummyHash []byte
	revogados TokenRevogadoRepository
}

type LoginResult struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	Usuario      repository.Usuario `json:"usuario"`
}

type RefreshResult struct {
	AccessToken string `json:"access_token"`
}

func NewLoginService(repo LoginRepository, tokens *AuthToken, revogados TokenRevogadoRepository) (*LoginService, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("gerar hash auxiliar: %w", err)
	}
	return &LoginService{repo: repo, tokens: tokens, dummyHash: hash, revogados: revogados}, nil
}

func (svc *LoginService) Login(ctx context.Context, email, senha string) (*LoginResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("login cancelado: %w", err)
	}
	email = strings.TrimSpace(email)
	if !validarEmail(email) || len(email) > 254 || len(senha) == 0 || len(senha) > 72 {
		return nil, ErrCredenciaisInvalidas
	}
	usuario, err := svc.repo.BuscarPorEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("buscar credenciais: %w", err)
	}
	hash := svc.dummyHash
	if usuario != nil {
		hash = []byte(usuario.SenhaHash)
	}
	senhaErr := bcrypt.CompareHashAndPassword(hash, []byte(senha))
	if senhaErr != nil || usuario == nil {
		return nil, ErrCredenciaisInvalidas
	}
	subject := strconv.FormatInt(int64(usuario.ID), 10)
	access, err := svc.tokens.emitir(subject, usuario.Idioma, "access", svc.tokens.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("gerar access token: %w", err)
	}
	refresh, err := svc.tokens.emitir(subject, usuario.Idioma, "refresh", svc.tokens.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("gerar refresh token: %w", err)
	}
	return &LoginResult{AccessToken: access, RefreshToken: refresh, Usuario: usuario.Usuario}, nil
}

func (svc *LoginService) Refresh(ctx context.Context, raw string) (*RefreshResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("refresh cancelado: %w", err)
	}
	claims, err := svc.tokens.validar(raw, "refresh")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSessaoExpirada, err)
	}
	revogado, err := svc.revogados.EstaRevogado(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("verificar refresh token revogado: %w", err)
	}
	if revogado {
		return nil, ErrSessaoExpirada
	}
	access, err := svc.tokens.emitir(claims.Subject, claims.Idioma, "access", svc.tokens.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("renovar access token: %w", err)
	}
	return &RefreshResult{AccessToken: access}, nil
}
