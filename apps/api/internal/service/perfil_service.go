package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrPerfilNomeObrigatorio           = errors.New("perfil.nome_obrigatorio")
	ErrPerfilNomeInvalido              = errors.New("perfil.nome_invalido")
	ErrPerfilUsernameInvalido          = errors.New("perfil.username_invalido")
	ErrPerfilUsernameEmUso             = errors.New("perfil.username_em_uso")
	ErrPerfilBioMuitoLonga             = errors.New("perfil.bio_muito_longa")
	ErrPerfilJogoFavoritoNaoEncontrado = errors.New("perfil.jogo_favorito_nao_encontrado")
	ErrPerfilConsoleInvalido           = errors.New("perfil.console_invalido")
	ErrPerfilJogandoDesdeInvalido      = errors.New("perfil.jogando_desde_invalido")
	ErrPerfilEntradaInvalida           = errors.New("perfil.entrada_invalida")
)

var regexUsernameValido = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)

type AtualizarPerfilInput struct {
	Nome            string  `json:"nome"`
	Username        *string `json:"username"`
	Bio             *string `json:"bio"`
	JogoFavoritoID  *int32  `json:"jogo_favorito_id"`
	ConsoleFavorito *string `json:"console_favorito"`
	JogandoDesde    *int32  `json:"jogando_desde"`
}

type PerfilRepository interface {
	BuscarPorID(ctx context.Context, id int32) (*repository.Usuario, error)
	BuscarPerfil(ctx context.Context, id int32) (*repository.PerfilUsuario, error)
	AtualizarPerfil(ctx context.Context, params repository.AtualizarPerfilParams) (*repository.PerfilUsuario, error)
}

type PerfilService struct {
	repo PerfilRepository
	now  func() time.Time
}

func NewPerfilService(repo PerfilRepository) *PerfilService {
	return &PerfilService{repo: repo, now: time.Now}
}

func (svc *PerfilService) SetNow(nowFn func() time.Time) {
	svc.now = nowFn
}

func (svc *PerfilService) Perfil(ctx context.Context, subject string) (*repository.Usuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("perfil cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}
	usuario, err := svc.repo.BuscarPorID(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("consultar perfil: %w", err)
	}
	if usuario == nil {
		return nil, ErrTokenInvalido
	}
	return usuario, nil
}

func (svc *PerfilService) ObterPerfil(ctx context.Context, subject string) (*repository.PerfilUsuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("obter perfil cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}
	perfil, err := svc.repo.BuscarPerfil(ctx, int32(id))
	if err != nil {
		return nil, fmt.Errorf("consultar perfil: %w", err)
	}
	if perfil == nil {
		return nil, ErrTokenInvalido
	}
	return perfil, nil
}

func (svc *PerfilService) AtualizarPerfil(
	ctx context.Context,
	subject string,
	input AtualizarPerfilInput,
) (*repository.PerfilUsuario, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("atualizar perfil cancelado: %w", err)
	}
	id, err := strconv.ParseInt(subject, 10, 32)
	if err != nil || id <= 0 {
		return nil, ErrTokenInvalido
	}

	nome := strings.TrimSpace(input.Nome)
	if nome == "" {
		return nil, ErrPerfilNomeObrigatorio
	}
	if len([]rune(nome)) > 100 {
		return nil, ErrPerfilNomeInvalido
	}

	var username *string
	if input.Username != nil {
		uname := strings.TrimSpace(*input.Username)
		if uname != "" {
			if len([]rune(uname)) < 3 || len([]rune(uname)) > 30 || !regexUsernameValido.MatchString(uname) {
				return nil, ErrPerfilUsernameInvalido
			}
			username = &uname
		}
	}

	var bio *string
	if input.Bio != nil {
		b := strings.TrimSpace(*input.Bio)
		if b != "" {
			if len([]rune(b)) > 280 {
				return nil, ErrPerfilBioMuitoLonga
			}
			bio = &b
		}
	}

	var jogoFavoritoID *int32
	if input.JogoFavoritoID != nil {
		if *input.JogoFavoritoID <= 0 {
			return nil, ErrPerfilJogoFavoritoNaoEncontrado
		}
		jogoFavoritoID = input.JogoFavoritoID
	}

	var consoleFavorito *string
	if input.ConsoleFavorito != nil {
		c := strings.TrimSpace(*input.ConsoleFavorito)
		if c != "" {
			if len([]rune(c)) > 100 {
				return nil, ErrPerfilConsoleInvalido
			}
			consoleFavorito = &c
		}
	}

	var jogandoDesde *int16
	if input.JogandoDesde != nil {
		anoAtual := svc.now().Year()
		if *input.JogandoDesde < 1970 || int(*input.JogandoDesde) > anoAtual {
			return nil, ErrPerfilJogandoDesdeInvalido
		}
		ano := int16(*input.JogandoDesde)
		jogandoDesde = &ano
	}

	perfil, err := svc.repo.AtualizarPerfil(ctx, repository.AtualizarPerfilParams{
		ID:              int32(id),
		Nome:            nome,
		Username:        username,
		Bio:             bio,
		JogoFavoritoID:  jogoFavoritoID,
		ConsoleFavorito: consoleFavorito,
		JogandoDesde:    jogandoDesde,
	})
	if err != nil {
		if errors.Is(err, repository.ErrUsernameEmUso) {
			return nil, ErrPerfilUsernameEmUso
		}
		if errors.Is(err, repository.ErrJogoFavoritoNaoEncontrado) {
			return nil, ErrPerfilJogoFavoritoNaoEncontrado
		}
		return nil, fmt.Errorf("atualizar perfil no repository: %w", err)
	}

	return perfil, nil
}
