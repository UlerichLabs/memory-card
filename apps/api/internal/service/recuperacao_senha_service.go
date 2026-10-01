package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrLimiteSolicitacoesReset = errors.New("auth.password_reset.rate_limited")
	ErrTokenResetExpirado      = errors.New("auth.password_reset.token_expired")
	ErrTokenResetUtilizado     = errors.New("auth.password_reset.token_used")
	ErrTokenResetInvalido      = errors.New("auth.password_reset.invalid_token")
)

type RecuperacaoSenhaRepository interface {
	RegistrarSolicitacao(ctx context.Context, email string) (bool, error)
	CriarToken(ctx context.Context, usuarioID int32, tokenHash string, expiraEm time.Time) error
	BuscarToken(ctx context.Context, tokenHash string) (*repository.TokenResetSenha, error)
	ConsumirETrocarSenha(ctx context.Context, tokenID int32, senhaHash string) (bool, error)
	ListarRefreshTokensAtivos(ctx context.Context, usuarioID int32) ([]repository.RefreshTokenAtivo, error)
	RemoverRefreshTokensAtivos(ctx context.Context, usuarioID int32) error
}

type RecuperacaoUsuarioRepository interface {
	BuscarPorEmail(ctx context.Context, email string) (*repository.CredenciaisUsuario, error)
}

type RefreshTokenRegistry interface {
	RegistrarRefreshToken(ctx context.Context, jti string, usuarioID int32, expiraEm time.Time) error
}

type RecuperacaoSenhaService struct {
	resetURL  string
	usuarios  RecuperacaoUsuarioRepository
	repo      RecuperacaoSenhaRepository
	revogados TokenRevogadoRepository
	email     EmailSender
	now       func() time.Time
}

func NewRecuperacaoSenhaService(resetURL string, usuarios RecuperacaoUsuarioRepository, repo RecuperacaoSenhaRepository, revogados TokenRevogadoRepository, email EmailSender) *RecuperacaoSenhaService {
	return &RecuperacaoSenhaService{resetURL: resetURL, usuarios: usuarios, repo: repo, revogados: revogados, email: email, now: time.Now}
}

func (svc *RecuperacaoSenhaService) Solicitar(ctx context.Context, email string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("solicitar reset cancelado: %w", err)
	}
	email = strings.TrimSpace(email)
	if !validarEmail(email) {
		return nil
	}
	permitido, err := svc.repo.RegistrarSolicitacao(ctx, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("verificar limite de reset: %w", err)
	}
	if !permitido {
		return ErrLimiteSolicitacoesReset
	}
	usuario, err := svc.usuarios.BuscarPorEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("buscar usuario para reset: %w", err)
	}
	if usuario == nil {
		return nil
	}
	raw, hash, err := gerarTokenReset()
	if err != nil {
		return err
	}
	expiraEm := svc.now().Add(30 * time.Minute)
	if err := svc.repo.CriarToken(ctx, usuario.ID, hash, expiraEm); err != nil {
		return fmt.Errorf("persistir token de reset: %w", err)
	}
	link, err := criarLinkReset(svc.resetURL, raw)
	if err != nil {
		return err
	}
	if err := svc.email.EnviarRecuperacaoSenha(ctx, usuario.Email, link); err != nil {
		return fmt.Errorf("enviar email de reset: %w", err)
	}
	return nil
}

func (svc *RecuperacaoSenhaService) ValidarToken(ctx context.Context, raw string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("validar reset cancelado: %w", err)
	}
	_, err := svc.tokenValido(ctx, raw)
	return err
}

func (svc *RecuperacaoSenhaService) RedefinirSenha(ctx context.Context, raw, senha string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("redefinir senha cancelado: %w", err)
	}
	token, err := svc.tokenValido(ctx, raw)
	if err != nil {
		return err
	}
	if !validarSenha(senha) {
		return ErrSenhaFraca
	}
	senhaHash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gerar hash da nova senha: %w", err)
	}
	consumido, err := svc.repo.ConsumirETrocarSenha(ctx, token.ID, string(senhaHash))
	if err != nil {
		return fmt.Errorf("trocar senha com token: %w", err)
	}
	if !consumido {
		return ErrTokenResetUtilizado
	}
	refreshTokens, err := svc.repo.ListarRefreshTokensAtivos(ctx, token.UsuarioID)
	if err != nil {
		return fmt.Errorf("listar sessoes ativas: %w", err)
	}
	for _, refreshToken := range refreshTokens {
		if err := svc.revogados.Revogar(ctx, refreshToken.JTI, refreshToken.ExpiraEm); err != nil {
			return fmt.Errorf("revogar sessao ativa: %w", err)
		}
	}
	if err := svc.repo.RemoverRefreshTokensAtivos(ctx, token.UsuarioID); err != nil {
		return fmt.Errorf("remover sessoes ativas: %w", err)
	}
	return nil
}

func (svc *RecuperacaoSenhaService) tokenValido(ctx context.Context, raw string) (*repository.TokenResetSenha, error) {
	if raw == "" {
		return nil, ErrTokenResetInvalido
	}
	token, err := svc.repo.BuscarToken(ctx, hashToken(raw))
	if err != nil {
		return nil, fmt.Errorf("consultar token de reset: %w", err)
	}
	if token == nil {
		return nil, ErrTokenResetInvalido
	}
	if token.UsadoEm != nil {
		return nil, ErrTokenResetUtilizado
	}
	if !token.ExpiraEm.After(svc.now()) {
		return nil, ErrTokenResetExpirado
	}
	return token, nil
}

func gerarTokenReset() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("gerar token de reset: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(bytes)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func criarLinkReset(baseURL, token string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("URL de reset invalida")
	}
	query := parsed.Query()
	query.Set("token", token)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
