// Package handler contem os handlers HTTP da API.
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type perfilCompletoServiceMock struct {
	obterPerfil     func(context.Context, string) (*repository.PerfilUsuario, error)
	atualizarPerfil func(context.Context, string, service.AtualizarPerfilInput) (*repository.PerfilUsuario, error)
}

func (mock perfilCompletoServiceMock) ObterPerfil(ctx context.Context, subject string) (*repository.PerfilUsuario, error) {
	if mock.obterPerfil != nil {
		return mock.obterPerfil(ctx, subject)
	}
	return nil, nil
}

func (mock perfilCompletoServiceMock) AtualizarPerfil(
	ctx context.Context,
	subject string,
	input service.AtualizarPerfilInput,
) (*repository.PerfilUsuario, error) {
	if mock.atualizarPerfil != nil {
		return mock.atualizarPerfil(ctx, subject, input)
	}
	return nil, nil
}

func gerarTokenTestePerfil(t *testing.T, secret string, usuarioID int32, idioma string) string {
	t.Helper()
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: idioma,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(usuarioID)),
			Issuer:    "memory-card",
			Audience:  jwt.ClaimStrings{"access"},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("falha ao gerar token: %v", err)
	}
	return raw
}

func TestPerfilHandler_ObterPerfil_SucessoE401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	username := "lucas_gamer"
	jogandoDesde := int16(1995)
	perfilEsperado := &repository.PerfilUsuario{
		Nome:                "Lucas",
		Email:               "lucas@example.com",
		Username:            &username,
		JogoFavorito:        &repository.JogoFavoritoResumo{ID: 10, Nome: "Chrono Trigger"},
		JogandoDesde:        &jogandoDesde,
		JogandoDesdeEfetivo: &jogandoDesde,
	}

	mock := perfilCompletoServiceMock{
		obterPerfil: func(ctx context.Context, subject string) (*repository.PerfilUsuario, error) {
			if subject != "42" {
				t.Fatalf("subject esperado 42, obteve %s", subject)
			}
			return perfilEsperado, nil
		},
	}

	handler := NewPerfilHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/me/perfil", handler.ObterPerfil)

	t.Run("200 com token valido", func(t *testing.T) {
		token := gerarTokenTestePerfil(t, secret, 42, "pt-BR")
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me/perfil", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("esperava status 200, obteve %d body=%s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Data.Nome != "Lucas" || resp.Data.Email != "lucas@example.com" ||
			resp.Data.Username == nil || *resp.Data.Username != username {
			t.Fatalf("dados do perfil inesperados: %+v", resp.Data)
		}
	})

	t.Run("401 sem token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me/perfil", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("esperava 401 sem token, obteve %d", rec.Code)
		}
	})
}

func TestPerfilHandler_AtualizarPerfil_SucessoE401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	username := "novo_user"
	jogandoDesde := int16(2000)
	perfilAtualizado := &repository.PerfilUsuario{
		Nome:                "Nome Atualizado",
		Email:               "lucas@example.com",
		Username:            &username,
		JogandoDesde:        &jogandoDesde,
		JogandoDesdeEfetivo: &jogandoDesde,
	}

	mock := perfilCompletoServiceMock{
		atualizarPerfil: func(ctx context.Context, subject string, input service.AtualizarPerfilInput) (*repository.PerfilUsuario, error) {
			if subject != "42" {
				t.Fatalf("subject esperado 42, obteve %s", subject)
			}
			if input.Nome != "Nome Atualizado" {
				t.Fatalf("nome inesperado %s", input.Nome)
			}
			return perfilAtualizado, nil
		},
	}

	handler := NewPerfilHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.PUT("/me/perfil", handler.AtualizarPerfil)

	t.Run("200 PUT com dados validos", func(t *testing.T) {
		token := gerarTokenTestePerfil(t, secret, 42, "pt-BR")
		corpo := []byte(`{"nome":"Nome Atualizado","username":"novo_user","jogando_desde":2000}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/perfil", bytes.NewReader(corpo))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("esperava status 200, obteve %d body=%s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Data.Nome != "Nome Atualizado" || *resp.Data.Username != "novo_user" {
			t.Fatalf("resposta inesperada: %+v", resp.Data)
		}
	})

	t.Run("401 sem token no PUT", func(t *testing.T) {
		corpo := []byte(`{"nome":"Nome Atualizado"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/perfil", bytes.NewReader(corpo))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("esperava 401 sem token, obteve %d", rec.Code)
		}
	})
}

func TestPerfilHandler_ErrosDeValidacao_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token := gerarTokenTestePerfil(t, secret, 42, "pt-BR")

	casos := []struct {
		nome           string
		serviceErr     error
		corpoJson      string
		esperadoStatus int
		esperadoCodigo string
	}{
		{
			nome:           "nome obrigatorio",
			serviceErr:     service.ErrPerfilNomeObrigatorio,
			corpoJson:      `{"nome":""}`,
			esperadoStatus: http.StatusBadRequest,
			esperadoCodigo: "perfil.nome_obrigatorio",
		},
		{
			nome:           "nome invalido longo",
			serviceErr:     service.ErrPerfilNomeInvalido,
			corpoJson:      `{"nome":"nome muito longo"}`,
			esperadoStatus: http.StatusBadRequest,
			esperadoCodigo: "perfil.nome_invalido",
		},
		{
			nome:           "username invalido",
			serviceErr:     service.ErrPerfilUsernameInvalido,
			corpoJson:      `{"nome":"Lucas","username":"ab"}`,
			esperadoStatus: http.StatusBadRequest,
			esperadoCodigo: "perfil.username_invalido",
		},
		{
			nome:           "bio muito longa",
			serviceErr:     service.ErrPerfilBioMuitoLonga,
			corpoJson:      `{"nome":"Lucas","bio":"bio longa"}`,
			esperadoStatus: http.StatusBadRequest,
			esperadoCodigo: "perfil.bio_muito_longa",
		},
		{
			nome:           "console invalido",
			serviceErr:     service.ErrPerfilConsoleInvalido,
			corpoJson:      `{"nome":"Lucas","console_favorito":"console longo"}`,
			esperadoStatus: http.StatusBadRequest,
			esperadoCodigo: "perfil.console_invalido",
		},
		{
			nome:           "jogando desde invalido",
			serviceErr:     service.ErrPerfilJogandoDesdeInvalido,
			corpoJson:      `{"nome":"Lucas","jogando_desde":1950}`,
			esperadoStatus: http.StatusBadRequest,
			esperadoCodigo: "perfil.jogando_desde_invalido",
		},
		{
			nome:           "entrada json malformada",
			serviceErr:     nil,
			corpoJson:      `{"jogando_desde":"nao-e-numero"}`,
			esperadoStatus: http.StatusBadRequest,
			esperadoCodigo: "perfil.entrada_invalida",
		},
	}

	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			mock := perfilCompletoServiceMock{
				atualizarPerfil: func(context.Context, string, service.AtualizarPerfilInput) (*repository.PerfilUsuario, error) {
					return nil, tc.serviceErr
				},
			}
			handler := NewPerfilHandler(mock)
			router := gin.New()
			middleware.GrupoPrivado(router, tokens).PUT("/me/perfil", handler.AtualizarPerfil)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/me/perfil", bytes.NewBufferString(tc.corpoJson))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.esperadoStatus {
				t.Fatalf("esperava status %d, obteve %d body=%s", tc.esperadoStatus, rec.Code, rec.Body.String())
			}
			var resp struct {
				Error struct {
					Codigo   string `json:"codigo"`
					Mensagem string `json:"mensagem"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error.Codigo != tc.esperadoCodigo {
				t.Fatalf("codigo esperado %s, obteve %s", tc.esperadoCodigo, resp.Error.Codigo)
			}
		})
	}
}

func TestPerfilHandler_404JogoFavoritoE409Username(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token := gerarTokenTestePerfil(t, secret, 42, "pt-BR")

	t.Run("404 jogo favorito nao encontrado ou de outro usuario", func(t *testing.T) {
		mock := perfilCompletoServiceMock{
			atualizarPerfil: func(context.Context, string, service.AtualizarPerfilInput) (*repository.PerfilUsuario, error) {
				return nil, service.ErrPerfilJogoFavoritoNaoEncontrado
			},
		}
		handler := NewPerfilHandler(mock)
		router := gin.New()
		middleware.GrupoPrivado(router, tokens).PUT("/me/perfil", handler.AtualizarPerfil)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/perfil", bytes.NewBufferString(`{"nome":"Lucas","jogo_favorito_id":999}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("esperava 404, obteve %d body=%s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Error struct {
				Codigo string `json:"codigo"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Error.Codigo != "perfil.jogo_favorito_nao_encontrado" {
			t.Fatalf("codigo inesperado: %s", resp.Error.Codigo)
		}
	})

	t.Run("409 username em uso", func(t *testing.T) {
		mock := perfilCompletoServiceMock{
			atualizarPerfil: func(context.Context, string, service.AtualizarPerfilInput) (*repository.PerfilUsuario, error) {
				return nil, service.ErrPerfilUsernameEmUso
			},
		}
		handler := NewPerfilHandler(mock)
		router := gin.New()
		middleware.GrupoPrivado(router, tokens).PUT("/me/perfil", handler.AtualizarPerfil)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/perfil", bytes.NewBufferString(`{"nome":"Lucas","username":"usuario_existente"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("esperava 409, obteve %d body=%s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Error struct {
				Codigo string `json:"codigo"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Error.Codigo != "perfil.username_em_uso" {
			t.Fatalf("codigo inesperado: %s", resp.Error.Codigo)
		}
	})
}

func TestPerfilHandler_Internacionalizacao_PtBR_E_EN(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	mock := perfilCompletoServiceMock{
		atualizarPerfil: func(context.Context, string, service.AtualizarPerfilInput) (*repository.PerfilUsuario, error) {
			return nil, service.ErrPerfilNomeObrigatorio
		},
	}
	handler := NewPerfilHandler(mock)
	router := gin.New()
	middleware.GrupoPrivado(router, tokens).PUT("/me/perfil", handler.AtualizarPerfil)

	t.Run("mensagem em pt-BR", func(t *testing.T) {
		token := gerarTokenTestePerfil(t, secret, 42, "pt-BR")
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/perfil", bytes.NewBufferString(`{"nome":""}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "pt-BR")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var resp struct {
			Error struct {
				Codigo   string `json:"codigo"`
				Mensagem string `json:"mensagem"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Error.Codigo != "perfil.nome_obrigatorio" || resp.Error.Mensagem != "O nome é obrigatório." {
			t.Fatalf("mensagem pt-BR incorreta: %+v", resp.Error)
		}
	})

	t.Run("mensagem em en", func(t *testing.T) {
		token := gerarTokenTestePerfil(t, secret, 42, "en")
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/perfil", bytes.NewBufferString(`{"nome":""}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var resp struct {
			Error struct {
				Codigo   string `json:"codigo"`
				Mensagem string `json:"mensagem"`
			} `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Error.Codigo != "perfil.nome_obrigatorio" || resp.Error.Mensagem != "Name is required." {
			t.Fatalf("mensagem en incorreta: %+v", resp.Error)
		}
	})
}
