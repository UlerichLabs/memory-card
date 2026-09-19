package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type resetSenhaServiceMock struct {
	solicitarFn func(context.Context, string) error
	validarFn   func(context.Context, string) error
	redefinirFn func(context.Context, string, string) error
}

func (mock resetSenhaServiceMock) Solicitar(ctx context.Context, email string) error {
	return mock.solicitarFn(ctx, email)
}

func (mock resetSenhaServiceMock) ValidarToken(ctx context.Context, token string) error {
	return mock.validarFn(ctx, token)
}

func (mock resetSenhaServiceMock) RedefinirSenha(ctx context.Context, token, senha string) error {
	return mock.redefinirFn(ctx, token, senha)
}

func TestRecuperacaoSenhaHandler_SolicitarReset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		serviceErr error
		status     int
		codigo     string
	}{
		{"email existente", `{"email":"usuario@example.com"}`, nil, http.StatusOK, ""},
		{"email inexistente", `{"email":"ausente@example.com"}`, nil, http.StatusOK, ""},
		{"corpo invalido", `{`, nil, http.StatusOK, ""},
		{"limite", `{"email":"usuario@example.com"}`, service.ErrLimiteSolicitacoesReset, http.StatusTooManyRequests, "auth.password_reset.rate_limited"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chamado := false
			handler := NewRecuperacaoSenhaHandler(resetSenhaServiceMock{solicitarFn: func(ctx context.Context, email string) error {
				chamado = true
				return tc.serviceErr
			}})
			router := gin.New()
			router.POST("/solicitar-reset", handler.SolicitarReset)
			req := httptest.NewRequest(http.MethodPost, "/solicitar-reset", strings.NewReader(tc.body))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != tc.status {
				t.Fatalf("status = %d", recorder.Code)
			}
			if tc.name == "corpo invalido" && chamado {
				t.Fatal("service nao deve ser chamado para JSON invalido")
			}
			var response struct {
				Data struct {
					Mensagem string `json:"mensagem"`
				} `json:"data"`
				Error struct {
					Codigo string `json:"codigo"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if tc.codigo != "" && response.Error.Codigo != tc.codigo {
				t.Fatal(response.Error.Codigo)
			}
			if tc.codigo == "" && response.Data.Mensagem != "Se este e-mail estiver cadastrado, você receberá as instruções em breve." {
				t.Fatal(response.Data.Mensagem)
			}
		})
	}
}

func TestRecuperacaoSenhaHandler_ValidarERedefinir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, path, body string
		serviceErr       error
		status           int
		codigo           string
	}{
		{"validar valido", "/validar-token-reset?token=valido", "", nil, http.StatusOK, ""},
		{"validar expirado", "/validar-token-reset?token=expirado", "", service.ErrTokenResetExpirado, http.StatusGone, "auth.password_reset.token_expired"},
		{"validar usado", "/validar-token-reset?token=usado", "", service.ErrTokenResetUtilizado, http.StatusConflict, "auth.password_reset.token_used"},
		{"validar inexistente", "/validar-token-reset?token=invalido", "", service.ErrTokenResetInvalido, http.StatusBadRequest, "auth.password_reset.invalid_token"},
		{"redefinir senha fraca", "/redefinir-senha", `{"token":"valido","senha":"fraca"}`, service.ErrSenhaFraca, http.StatusBadRequest, "auth.password_reset.weak_password"},
		{"redefinir sucesso", "/redefinir-senha", `{"token":"valido","senha":"SenhaNova@123"}`, nil, http.StatusOK, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewRecuperacaoSenhaHandler(resetSenhaServiceMock{
				validarFn:   func(context.Context, string) error { return tc.serviceErr },
				redefinirFn: func(context.Context, string, string) error { return tc.serviceErr },
			})
			router := gin.New()
			router.GET("/validar-token-reset", handler.ValidarTokenReset)
			router.POST("/redefinir-senha", handler.RedefinirSenha)
			method := http.MethodGet
			if tc.body != "" {
				method = http.MethodPost
			}
			req := httptest.NewRequest(method, tc.path, strings.NewReader(tc.body))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != tc.status {
				t.Fatalf("status = %d: %s", recorder.Code, recorder.Body.String())
			}
			if tc.codigo == "" {
				return
			}
			var response struct {
				Error struct {
					Codigo string `json:"codigo"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Error.Codigo != tc.codigo {
				t.Fatal(response.Error.Codigo)
			}
		})
	}
}

func TestRecuperacaoSenhaHandler_ErroInterno(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRecuperacaoSenhaHandler(resetSenhaServiceMock{validarFn: func(context.Context, string) error { return errors.New("detalhe privado") }})
	router := gin.New()
	router.GET("/validar-token-reset", handler.ValidarTokenReset)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/validar-token-reset?token=valido", nil))
	if recorder.Code != http.StatusInternalServerError || strings.Contains(recorder.Body.String(), "detalhe privado") {
		t.Fatal(recorder.Body.String())
	}
}
