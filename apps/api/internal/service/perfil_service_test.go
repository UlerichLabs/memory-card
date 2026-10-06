package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type perfilRepoMock struct {
	buscarPorID     func(context.Context, int32) (*repository.Usuario, error)
	buscarPerfil    func(context.Context, int32) (*repository.PerfilUsuario, error)
	atualizarPerfil func(context.Context, repository.AtualizarPerfilParams) (*repository.PerfilUsuario, error)
}

func (mock perfilRepoMock) BuscarPorID(ctx context.Context, id int32) (*repository.Usuario, error) {
	if mock.buscarPorID != nil {
		return mock.buscarPorID(ctx, id)
	}
	return nil, nil
}

func (mock perfilRepoMock) BuscarPerfil(ctx context.Context, id int32) (*repository.PerfilUsuario, error) {
	if mock.buscarPerfil != nil {
		return mock.buscarPerfil(ctx, id)
	}
	return nil, nil
}

func (mock perfilRepoMock) AtualizarPerfil(ctx context.Context, params repository.AtualizarPerfilParams) (*repository.PerfilUsuario, error) {
	if mock.atualizarPerfil != nil {
		return mock.atualizarPerfil(ctx, params)
	}
	return nil, nil
}

func TestPerfil_Cenarios(t *testing.T) {
	usuario := &repository.Usuario{ID: 42, Nome: "Lucas", Email: "lucas@example.com", Idioma: "en"}
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name             string
		usuario          *repository.Usuario
		repoErr, wantErr error
	}{
		{"sucesso", usuario, nil, nil},
		{"usuario removido", nil, nil, ErrTokenInvalido},
		{"falha no banco", nil, dbErr, dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			called := false
			svc := NewPerfilService(perfilRepoMock{buscarPorID: func(got context.Context, id int32) (*repository.Usuario, error) {
				called = true
				if got != ctx || id != usuario.ID {
					t.Fatal("contexto ou identidade incorretos")
				}
				return tc.usuario, tc.repoErr
			}})
			result, err := svc.Perfil(ctx, "42")
			if !called || !errors.Is(err, tc.wantErr) || result != tc.usuario {
				t.Fatalf("perfil = %v, erro = %v", result, err)
			}
		})
	}
}

func TestPerfil_IdentidadeInvalida(t *testing.T) {
	svc := NewPerfilService(perfilRepoMock{buscarPorID: func(context.Context, int32) (*repository.Usuario, error) {
		t.Fatal("nao deveria consultar repositorio")
		return nil, nil
	}})
	for _, subject := range []string{"", "abc", "0", "-1", "2147483648"} {
		t.Run(subject, func(t *testing.T) {
			result, err := svc.Perfil(context.Background(), subject)
			if result != nil || !errors.Is(err, ErrTokenInvalido) {
				t.Fatalf("erro = %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.Perfil(ctx, "42"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestPerfilService_AtualizarPerfil_SucessoEParametros(t *testing.T) {
	ctx := context.Background()
	usuarioID := int32(42)
	username := "lucas_gamer"
	bio := "Apaixonado por RPGs clássicos"
	jogoFavoritoID := int32(101)
	consoleFavorito := "SNES"
	jogandoDesde := int32(1995)
	jogandoDesdeInt16 := int16(1995)

	esperadoPerfil := &repository.PerfilUsuario{
		Nome:                "Lucas Silva",
		Email:               "lucas@example.com",
		Username:            &username,
		Bio:                 &bio,
		JogoFavorito:        &repository.JogoFavoritoResumo{ID: 101, Nome: "Chrono Trigger"},
		ConsoleFavorito:     &consoleFavorito,
		JogandoDesde:        &jogandoDesdeInt16,
		JogandoDesdeEfetivo: &jogandoDesdeInt16,
	}

	chamado := false
	mock := perfilRepoMock{
		atualizarPerfil: func(reqCtx context.Context, params repository.AtualizarPerfilParams) (*repository.PerfilUsuario, error) {
			chamado = true
			if reqCtx != ctx {
				t.Fatal("contexto incorreto")
			}
			if params.ID != usuarioID {
				t.Fatalf("id esperado %d, obteve %d", usuarioID, params.ID)
			}
			if params.Nome != "Lucas Silva" {
				t.Fatalf("nome esperado Lucas Silva, obteve %s", params.Nome)
			}
			if params.Username == nil || *params.Username != username {
				t.Fatalf("username incorreto: %+v", params.Username)
			}
			if params.Bio == nil || *params.Bio != bio {
				t.Fatalf("bio incorreta: %+v", params.Bio)
			}
			if params.JogoFavoritoID == nil || *params.JogoFavoritoID != jogoFavoritoID {
				t.Fatalf("jogoFavoritoID incorreto: %+v", params.JogoFavoritoID)
			}
			if params.ConsoleFavorito == nil || *params.ConsoleFavorito != consoleFavorito {
				t.Fatalf("consoleFavorito incorreto: %+v", params.ConsoleFavorito)
			}
			if params.JogandoDesde == nil || *params.JogandoDesde != jogandoDesdeInt16 {
				t.Fatalf("jogandoDesde incorreto: %+v", params.JogandoDesde)
			}
			return esperadoPerfil, nil
		},
	}

	svc := NewPerfilService(mock)
	svc.SetNow(func() time.Time { return time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC) })

	input := AtualizarPerfilInput{
		Nome:            "  Lucas Silva  ",
		Username:        &username,
		Bio:             &bio,
		JogoFavoritoID:  &jogoFavoritoID,
		ConsoleFavorito: &consoleFavorito,
		JogandoDesde:    &jogandoDesde,
	}

	perfil, err := svc.AtualizarPerfil(ctx, "42", input)
	if err != nil {
		t.Fatalf("falha inesperada: %v", err)
	}
	if !chamado {
		t.Fatal("repository AtualizarPerfil nao foi chamado")
	}
	if perfil != esperadoPerfil {
		t.Fatalf("perfil retornado divergente: %+v", perfil)
	}
}

func TestPerfilService_AtualizarPerfil_Validacoes(t *testing.T) {
	svc := NewPerfilService(perfilRepoMock{})
	svc.SetNow(func() time.Time { return time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC) })

	nomeLongo := strings.Repeat("a", 101)
	usernameCurto := "ab"
	usernameLongo := strings.Repeat("u", 31)
	usernameInvalidoHifen := "usuario-gamer"
	usernameInvalidoEspaco := "usuario gamer"
	usernameInvalidoArroba := "usuario@gamer"
	bioLonga := strings.Repeat("b", 281)
	consoleLongo := strings.Repeat("c", 101)
	anoAntesDe1970 := int32(1969)
	anoFuturo := int32(2027)
	jogoIdInvalido := int32(0)
	jogoIdNegativo := int32(-5)

	casos := []struct {
		nome    string
		input   AtualizarPerfilInput
		wantErr error
	}{
		{
			nome:    "nome vazio",
			input:   AtualizarPerfilInput{Nome: ""},
			wantErr: ErrPerfilNomeObrigatorio,
		},
		{
			nome:    "nome apenas espacos",
			input:   AtualizarPerfilInput{Nome: "   "},
			wantErr: ErrPerfilNomeObrigatorio,
		},
		{
			nome:    "nome muito longo",
			input:   AtualizarPerfilInput{Nome: nomeLongo},
			wantErr: ErrPerfilNomeInvalido,
		},
		{
			nome:    "username curto",
			input:   AtualizarPerfilInput{Nome: "Valido", Username: &usernameCurto},
			wantErr: ErrPerfilUsernameInvalido,
		},
		{
			nome:    "username longo",
			input:   AtualizarPerfilInput{Nome: "Valido", Username: &usernameLongo},
			wantErr: ErrPerfilUsernameInvalido,
		},
		{
			nome:    "username com hifen",
			input:   AtualizarPerfilInput{Nome: "Valido", Username: &usernameInvalidoHifen},
			wantErr: ErrPerfilUsernameInvalido,
		},
		{
			nome:    "username com espaco",
			input:   AtualizarPerfilInput{Nome: "Valido", Username: &usernameInvalidoEspaco},
			wantErr: ErrPerfilUsernameInvalido,
		},
		{
			nome:    "username com arroba",
			input:   AtualizarPerfilInput{Nome: "Valido", Username: &usernameInvalidoArroba},
			wantErr: ErrPerfilUsernameInvalido,
		},
		{
			nome:    "bio muito longa",
			input:   AtualizarPerfilInput{Nome: "Valido", Bio: &bioLonga},
			wantErr: ErrPerfilBioMuitoLonga,
		},
		{
			nome:    "console muito longo",
			input:   AtualizarPerfilInput{Nome: "Valido", ConsoleFavorito: &consoleLongo},
			wantErr: ErrPerfilConsoleInvalido,
		},
		{
			nome:    "jogando desde anterior a 1970",
			input:   AtualizarPerfilInput{Nome: "Valido", JogandoDesde: &anoAntesDe1970},
			wantErr: ErrPerfilJogandoDesdeInvalido,
		},
		{
			nome:    "jogando desde no futuro",
			input:   AtualizarPerfilInput{Nome: "Valido", JogandoDesde: &anoFuturo},
			wantErr: ErrPerfilJogandoDesdeInvalido,
		},
		{
			nome:    "jogo favorito id zero",
			input:   AtualizarPerfilInput{Nome: "Valido", JogoFavoritoID: &jogoIdInvalido},
			wantErr: ErrPerfilJogoFavoritoNaoEncontrado,
		},
		{
			nome:    "jogo favorito id negativo",
			input:   AtualizarPerfilInput{Nome: "Valido", JogoFavoritoID: &jogoIdNegativo},
			wantErr: ErrPerfilJogoFavoritoNaoEncontrado,
		},
	}

	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			_, err := svc.AtualizarPerfil(context.Background(), "42", tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("esperava erro %v, obteve %v", tc.wantErr, err)
			}
		})
	}
}

func TestPerfilService_AtualizarPerfil_ErrosDoRepositorio(t *testing.T) {
	ctx := context.Background()

	t.Run("conflito de username", func(t *testing.T) {
		mock := perfilRepoMock{
			atualizarPerfil: func(context.Context, repository.AtualizarPerfilParams) (*repository.PerfilUsuario, error) {
				return nil, repository.ErrUsernameEmUso
			},
		}
		svc := NewPerfilService(mock)
		uname := "usuario_duplicado"
		_, err := svc.AtualizarPerfil(ctx, "42", AtualizarPerfilInput{Nome: "Valido", Username: &uname})
		if !errors.Is(err, ErrPerfilUsernameEmUso) {
			t.Fatalf("esperava ErrPerfilUsernameEmUso, obteve: %v", err)
		}
	})

	t.Run("jogo favorito nao encontrado ou de outro usuario", func(t *testing.T) {
		mock := perfilRepoMock{
			atualizarPerfil: func(context.Context, repository.AtualizarPerfilParams) (*repository.PerfilUsuario, error) {
				return nil, repository.ErrJogoFavoritoNaoEncontrado
			},
		}
		svc := NewPerfilService(mock)
		jogoID := int32(999)
		_, err := svc.AtualizarPerfil(ctx, "42", AtualizarPerfilInput{Nome: "Valido", JogoFavoritoID: &jogoID})
		if !errors.Is(err, ErrPerfilJogoFavoritoNaoEncontrado) {
			t.Fatalf("esperava ErrPerfilJogoFavoritoNaoEncontrado, obteve: %v", err)
		}
	})
}

func TestPerfilService_ObterPerfil(t *testing.T) {
	ctx := context.Background()
	perfilExemplo := &repository.PerfilUsuario{
		Nome:  "Lucas",
		Email: "lucas@example.com",
	}

	t.Run("sucesso", func(t *testing.T) {
		mock := perfilRepoMock{
			buscarPerfil: func(reqCtx context.Context, id int32) (*repository.PerfilUsuario, error) {
				if id != 42 {
					t.Fatalf("id inesperado %d", id)
				}
				return perfilExemplo, nil
			},
		}
		svc := NewPerfilService(mock)
		res, err := svc.ObterPerfil(ctx, "42")
		if err != nil || res != perfilExemplo {
			t.Fatalf("resultado inesperado: res=%+v err=%v", res, err)
		}
	})

	t.Run("usuario nao encontrado", func(t *testing.T) {
		mock := perfilRepoMock{
			buscarPerfil: func(context.Context, int32) (*repository.PerfilUsuario, error) {
				return nil, nil
			},
		}
		svc := NewPerfilService(mock)
		_, err := svc.ObterPerfil(ctx, "42")
		if !errors.Is(err, ErrTokenInvalido) {
			t.Fatalf("esperava ErrTokenInvalido, obteve %v", err)
		}
	})

	t.Run("identidade invalida", func(t *testing.T) {
		svc := NewPerfilService(perfilRepoMock{})
		_, err := svc.ObterPerfil(ctx, "invalido")
		if !errors.Is(err, ErrTokenInvalido) {
			t.Fatalf("esperava ErrTokenInvalido, obteve %v", err)
		}
	})
}
