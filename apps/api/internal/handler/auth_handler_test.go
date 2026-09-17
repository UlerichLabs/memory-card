package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type mockCadastroService struct {
	cadastrarFn func(ctx context.Context, nome, email, senha string) (*repository.Usuario, error)
}

func (m *mockCadastroService) Cadastrar(ctx context.Context, nome, email, senha string) (*repository.Usuario, error) {
	if m.cadastrarFn != nil {
		return m.cadastrarFn(ctx, nome, email, senha)
	}
	return nil, nil
}

func TestAuthHandler_Register_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockCadastroService{
		cadastrarFn: func(ctx context.Context, nome, email, senha string) (*repository.Usuario, error) {
			return &repository.Usuario{
				ID:        1,
				Nome:      nome,
				Email:     email,
				Idioma:    "pt-BR",
				CreatedAt: time.Now(),
			}, nil
		},
	}

	h := NewAuthHandler(mockSvc)
	router := gin.New()
	router.POST("/api/v1/auth/register", h.Register)

	body := []byte(`{"nome":"Lucas","email":"lucas@example.com","senha":"SenhaForte@123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusCreated)
	}

	var resp struct {
		Data repository.Usuario `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("falha ao decodificar JSON: %v", err)
	}
	if resp.Data.ID != 1 || resp.Data.Email != "lucas@example.com" {
		t.Fatalf("resposta inesperada: %+v", resp.Data)
	}
}

func TestAuthHandler_Register_ErrosDeValidacao(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serviceErr     error
		acceptLanguage string
		wantStatus     int
		wantCodigo     string
		wantMensagem   string
	}{
		{
			name:           "email ja cadastrado pt-BR",
			serviceErr:     service.ErrEmailJaCadastrado,
			acceptLanguage: "pt-BR",
			wantStatus:     http.StatusConflict,
			wantCodigo:     "auth.register.email_taken",
			wantMensagem:   "Este email já está em uso.",
		},
		{
			name:           "email ja cadastrado en",
			serviceErr:     service.ErrEmailJaCadastrado,
			acceptLanguage: "en-US",
			wantStatus:     http.StatusConflict,
			wantCodigo:     "auth.register.email_taken",
			wantMensagem:   "This email is already in use.",
		},
		{
			name:           "senha fraca pt-BR",
			serviceErr:     service.ErrSenhaFraca,
			acceptLanguage: "",
			wantStatus:     http.StatusBadRequest,
			wantCodigo:     "auth.register.weak_password",
			wantMensagem:   "A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.",
		},
		{
			name:           "email invalido pt-BR",
			serviceErr:     service.ErrEmailInvalido,
			acceptLanguage: "pt-BR",
			wantStatus:     http.StatusBadRequest,
			wantCodigo:     "auth.register.invalid_email",
			wantMensagem:   "Informe um email válido.",
		},
		{
			name:           "erro interno inesperado",
			serviceErr:     errors.New("db timeout"),
			acceptLanguage: "pt-BR",
			wantStatus:     http.StatusInternalServerError,
			wantCodigo:     "server.internal_error",
			wantMensagem:   "Erro interno do servidor.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := &mockCadastroService{
				cadastrarFn: func(ctx context.Context, nome, email, senha string) (*repository.Usuario, error) {
					return nil, tc.serviceErr
				},
			}

			h := NewAuthHandler(mockSvc)
			router := gin.New()
			router.POST("/api/v1/auth/register", h.Register)

			body := []byte(`{"nome":"Lucas","email":"lucas@example.com","senha":"SenhaForte@123"}`)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tc.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tc.acceptLanguage)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, esperado %d", w.Code, tc.wantStatus)
			}

			var errResp struct {
				Error struct {
					Codigo   string `json:"codigo"`
					Mensagem string `json:"mensagem"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
				t.Fatalf("falha ao decodificar JSON: %v", err)
			}
			if errResp.Error.Codigo != tc.wantCodigo {
				t.Errorf("codigo = %q, esperado %q", errResp.Error.Codigo, tc.wantCodigo)
			}
			if errResp.Error.Mensagem != tc.wantMensagem {
				t.Errorf("mensagem = %q, esperado %q", errResp.Error.Mensagem, tc.wantMensagem)
			}
		})
	}
}

func TestAuthHandler_Register_CorpoInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := &mockCadastroService{}
	h := NewAuthHandler(mockSvc)
	router := gin.New()
	router.POST("/api/v1/auth/register", h.Register)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("{invalido"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusBadRequest)
	}
}
