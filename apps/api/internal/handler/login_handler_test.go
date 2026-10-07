package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type loginRepositoryMock struct {
	usuario      *repository.CredenciaisUsuario
	err          error
	usuarioPorID *repository.Usuario
}

func (mock loginRepositoryMock) BuscarPorEmail(context.Context, string) (*repository.CredenciaisUsuario, error) {
	return mock.usuario, mock.err
}

func (mock loginRepositoryMock) BuscarPorID(context.Context, int32) (*repository.Usuario, error) {
	if mock.err != nil {
		return nil, mock.err
	}
	if mock.usuarioPorID != nil {
		return mock.usuarioPorID, nil
	}
	if mock.usuario == nil {
		return nil, nil
	}
	return &mock.usuario.Usuario, nil
}

func TestAuthHandler_LoginRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash, err := bcrypt.GenerateFromPassword([]byte("SenhaForte@123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	usuario := &repository.CredenciaisUsuario{Usuario: repository.Usuario{ID: 1, Email: "lucas@example.com", Idioma: "pt-BR"}, SenhaHash: string(hash)}
	for _, tc := range []struct {
		name, path, body, lang string
		usuario                *repository.CredenciaisUsuario
		repoErr                error
		status                 int
		codigo, mensagem       string
	}{
		{"login sucesso", "/login", `{"email":"lucas@example.com","senha":"SenhaForte@123"}`, "", usuario, nil, 200, "", ""},
		{"email ausente", "/login", `{"email":"lucas@example.com","senha":"SenhaForte@123"}`, "", nil, nil, 401, "auth.login.invalid_credentials", "Email ou senha inválidos."},
		{"senha incorreta", "/login", `{"email":"lucas@example.com","senha":"errada"}`, "", usuario, nil, 401, "auth.login.invalid_credentials", "Email ou senha inválidos."},
		{"ingles", "/login", `{"email":"lucas@example.com","senha":"errada"}`, "en", usuario, nil, 401, "auth.login.invalid_credentials", "Invalid email or password."},
		{"json invalido", "/login", `{`, "", nil, nil, 400, "auth.login.invalid_input", "Dados de entrada inválidos."},
		{"erro banco", "/login", `{"email":"lucas@example.com","senha":"SenhaForte@123"}`, "", nil, errors.New("database unavailable"), 500, "server.internal_error", "Erro interno do servidor."},
		{"refresh invalido", "/refresh", `{"refresh_token":"invalido"}`, "", nil, nil, 401, "auth.session.expired", "Sessão expirada. Faça login novamente."},
		{"refresh json invalido", "/refresh", `{`, "en", nil, nil, 401, "auth.session.expired", "Session expired. Please log in again."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := service.NewAuthToken(uuid.NewString(), 15*time.Minute, 168*time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			svc, err := service.NewLoginService(loginRepositoryMock{usuario: tc.usuario, err: tc.repoErr}, tokens, revogacaoRepositoryMock{consultar: func(context.Context, string) (bool, error) {
				return false, nil
			}}, revogacaoRepositoryMock{})
			if err != nil {
				t.Fatal(err)
			}
			h := NewLoginHandler(svc)
			router := gin.New()
			router.POST("/login", h.Login)
			router.POST("/refresh", h.Refresh)
			req := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
			req.Header.Set("Accept-Language", tc.lang)
			result := httptest.NewRecorder()
			router.ServeHTTP(result, req)
			if result.Code != tc.status {
				t.Fatalf("status %d: %s", result.Code, result.Body.String())
			}
			var body struct {
				Data  service.LoginResult `json:"data"`
				Error struct {
					Codigo   string `json:"codigo"`
					Mensagem string `json:"mensagem"`
				} `json:"error"`
			}
			if err := json.Unmarshal(result.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tc.status != 200 {
				if body.Error.Codigo != tc.codigo || body.Error.Mensagem != tc.mensagem {
					t.Fatal(result.Body.String())
				}
				return
			}
			if strings.Contains(result.Body.String(), "senha") || body.Data.RefreshToken == "" || body.Data.Usuario.ID != 1 {
				t.Fatal("resposta incorreta")
			}
			refreshReq := httptest.NewRequest("POST", "/refresh", strings.NewReader(`{"refresh_token":"`+body.Data.RefreshToken+`"}`))
			refreshResult := httptest.NewRecorder()
			router.ServeHTTP(refreshResult, refreshReq)
			var refreshed struct {
				Data service.RefreshResult `json:"data"`
			}
			if err := json.Unmarshal(refreshResult.Body.Bytes(), &refreshed); err != nil {
				t.Fatal(err)
			}
			if refreshResult.Code != 200 || refreshed.Data.AccessToken == body.Data.AccessToken {
				t.Fatal("refresh incorreto")
			}
			if _, err := tokens.ValidarAccess(refreshed.Data.AccessToken); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAuthHandler_Refresh_IdiomaAtualizadoEUsuarioRemovido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash, err := bcrypt.GenerateFromPassword([]byte("SenhaForte@123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	usuarioBase := &repository.CredenciaisUsuario{
		Usuario:   repository.Usuario{ID: 1, Nome: "Lucas", Email: "lucas@example.com", Idioma: "pt-BR"},
		SenhaHash: string(hash),
	}
	tokens, err := service.NewAuthToken(uuid.NewString(), 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	repoInicial := loginRepositoryMock{usuario: usuarioBase}
	svcInicial, err := service.NewLoginService(repoInicial, tokens, revogacaoRepositoryMock{consultar: func(context.Context, string) (bool, error) {
		return false, nil
	}}, revogacaoRepositoryMock{})
	if err != nil {
		t.Fatal(err)
	}
	loginRes, err := svcInicial.Login(context.Background(), "lucas@example.com", "SenhaForte@123")
	if err != nil {
		t.Fatal(err)
	}
	refresh := loginRes.RefreshToken

	t.Run("200 com idioma atualizado", func(t *testing.T) {
		repo := loginRepositoryMock{
			usuarioPorID: &repository.Usuario{ID: 1, Nome: "Lucas", Email: "lucas@example.com", Idioma: "en"},
		}
		svc, err := service.NewLoginService(repo, tokens, revogacaoRepositoryMock{consultar: func(context.Context, string) (bool, error) {
			return false, nil
		}}, revogacaoRepositoryMock{})
		if err != nil {
			t.Fatal(err)
		}
		h := NewLoginHandler(svc)
		router := gin.New()
		router.POST("/refresh", h.Refresh)

		req := httptest.NewRequest("POST", "/refresh", strings.NewReader(`{"refresh_token":"`+refresh+`"}`))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Fatalf("status %d: %s", res.Code, res.Body.String())
		}
		var body struct {
			Data service.RefreshResult `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		claims, err := tokens.ValidarAccess(body.Data.AccessToken)
		if err != nil {
			t.Fatal(err)
		}
		if claims.Idioma != "en" {
			t.Fatalf("idioma esperado 'en', obtido '%s'", claims.Idioma)
		}
	})

	t.Run("401 para usuario removido", func(t *testing.T) {
		repo := loginRepositoryMock{
			usuarioPorID: nil,
		}
		svc, err := service.NewLoginService(repo, tokens, revogacaoRepositoryMock{consultar: func(context.Context, string) (bool, error) {
			return false, nil
		}}, revogacaoRepositoryMock{})
		if err != nil {
			t.Fatal(err)
		}
		h := NewLoginHandler(svc)
		router := gin.New()
		router.POST("/refresh", h.Refresh)

		req := httptest.NewRequest("POST", "/refresh", strings.NewReader(`{"refresh_token":"`+refresh+`"}`))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		if res.Code != 401 {
			t.Fatalf("status esperado 401, obtido %d: %s", res.Code, res.Body.String())
		}
		var errBody struct {
			Error struct {
				Codigo   string `json:"codigo"`
				Mensagem string `json:"mensagem"`
			} `json:"error"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &errBody); err != nil {
			t.Fatal(err)
		}
		if errBody.Error.Codigo != "auth.session.expired" {
			t.Fatalf("codigo esperado 'auth.session.expired', obtido '%s'", errBody.Error.Codigo)
		}
	})
}
