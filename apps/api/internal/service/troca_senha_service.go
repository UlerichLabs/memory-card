package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrSenhaAtualIncorreta = errors.New("auth.password_change.current_password_invalid")
	ErrNovaSenhaIgualAtual = errors.New("auth.password_change.same_password")
)

type TrocaSenhaRepository interface {
	BuscarCredenciaisPorID(ctx context.Context, id int32) (*repository.CredenciaisUsuario, error)
	AtualizarSenha(ctx context.Context, id int32, senhaHash string) error
}

type RefreshTokenRepository interface {
	ListarRefreshTokensAtivos(ctx context.Context, usuarioID int32) ([]repository.RefreshTokenAtivo, error)
	RemoverRefreshTokensAtivos(ctx context.Context, usuarioID int32) error
}

type TrocaSenhaService struct {
	usuarios      TrocaSenhaRepository
	refreshTokens RefreshTokenRepository
	revogados     TokenRevogadoRepository
}

func NewTrocaSenhaService(usuarios TrocaSenhaRepository, refreshTokens RefreshTokenRepository, revogados TokenRevogadoRepository) *TrocaSenhaService {
	return &TrocaSenhaService{usuarios: usuarios, refreshTokens: refreshTokens, revogados: revogados}
}

func (svc *TrocaSenhaService) Trocar(ctx context.Context, subject, senhaAtual, novaSenha string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("troca de senha cancelada: %w", err)
	}
	usuarioID, err := strconv.ParseInt(subject, 10, 32)
	if err != nil {
		return fmt.Errorf("identificador de usuario invalido: %w", err)
	}
	usuario, err := svc.usuarios.BuscarCredenciaisPorID(ctx, int32(usuarioID))
	if err != nil {
		return fmt.Errorf("buscar credenciais para troca de senha: %w", err)
	}
	if usuario == nil || bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(senhaAtual)) != nil {
		return ErrSenhaAtualIncorreta
	}
	if !validarSenha(novaSenha) {
		return ErrSenhaFraca
	}
	if bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(novaSenha)) == nil {
		return ErrNovaSenhaIgualAtual
	}
	novaSenhaHash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gerar hash da nova senha: %w", err)
	}
	if err := svc.usuarios.AtualizarSenha(ctx, usuario.ID, string(novaSenhaHash)); err != nil {
		return fmt.Errorf("persistir nova senha: %w", err)
	}
	refreshTokens, err := svc.refreshTokens.ListarRefreshTokensAtivos(ctx, usuario.ID)
	if err != nil {
		return fmt.Errorf("listar sessoes ativas: %w", err)
	}
	for _, refreshToken := range refreshTokens {
		if err := svc.revogados.Revogar(ctx, refreshToken.JTI, refreshToken.ExpiraEm); err != nil {
			return fmt.Errorf("revogar sessao ativa: %w", err)
		}
	}
	if err := svc.refreshTokens.RemoverRefreshTokensAtivos(ctx, usuario.ID); err != nil {
		return fmt.Errorf("remover sessoes ativas: %w", err)
	}
	return nil
}
