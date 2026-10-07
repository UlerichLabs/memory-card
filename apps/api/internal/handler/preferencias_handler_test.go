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

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type preferenciasServiceMock struct {
	obterPreferencias     func(context.Context, string) (*repository.PreferenciasUsuario, error)
	atualizarPreferencias func(context.Context, string, service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error)
}

func (m preferenciasServiceMock) ObterPreferencias(ctx context.Context, subject string) (*repository.PreferenciasUsuario, error) {
	if m.obterPreferencias != nil {
		return m.obterPreferencias(ctx, subject)
	}
	return nil, nil
}

func (m preferenciasServiceMock) AtualizarPreferencias(ctx context.Context, subject string, input service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error) {
	if m.atualizarPreferencias != nil {
		return m.atualizarPreferencias(ctx, subject, input)
	}
	return nil, nil
}

func gerarTokenTestePreferencias(t *testing.T, secret string, usuarioID int32, idioma string) string {
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

func payloadJSONValido() map[string]any {
	return map[string]any{
		"idioma":             "pt-BR",
		"formato_data":       "dmy",
		"fuso_horario":       "America/Sao_Paulo",
		"visual_biblioteca":  "grade",
		"itens_por_pagina":   24,
		"reduzir_animacoes":  false,
		"modo_tema":          "escuro",
		"estilo_tema":        "padrao",
	}
}

func TestPreferenciasHandler_ObterPreferencias_200EDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	esperado := &repository.PreferenciasUsuario{
		Idioma:           "pt-BR",
		FormatoData:      "dmy",
		FusoHorario:      "America/Sao_Paulo",
		VisualBiblioteca: "grade",
		ItensPorPagina:   24,
		ReduzirAnimacoes: false,
		ModoTema:         "escuro",
		EstiloTema:       "padrao",
	}

	mock := preferenciasServiceMock{
		obterPreferencias: func(ctx context.Context, subject string) (*repository.PreferenciasUsuario, error) {
			if subject != "42" {
				t.Fatalf("subject esperado 42, recebido %s", subject)
			}
			return esperado, nil
		},
	}

	handler := NewPreferenciasHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/me/preferencias", handler.ObterPreferencias)

	token := gerarTokenTestePreferencias(t, secret, 42, "pt-BR")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/preferencias", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Data repository.PreferenciasUsuario `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}
	if resp.Data != *esperado {
		t.Fatalf("resposta = %+v, want %+v", resp.Data, *esperado)
	}
}

func TestPreferenciasHandler_AtualizarPreferencias_200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	retorno := &repository.PreferenciasUsuario{
		Idioma:           "en",
		FormatoData:      "ymd",
		FusoHorario:      "Europe/Lisbon",
		VisualBiblioteca: "lista",
		ItensPorPagina:   50,
		ReduzirAnimacoes: true,
		ModoTema:         "claro",
		EstiloTema:       "playstation",
	}

	mock := preferenciasServiceMock{
		atualizarPreferencias: func(ctx context.Context, subject string, input service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error) {
			if subject != "42" {
				t.Fatalf("subject esperado 42, recebido %s", subject)
			}
			return retorno, nil
		},
	}

	handler := NewPreferenciasHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.PUT("/me/preferencias", handler.AtualizarPreferencias)

	payload := map[string]any{
		"idioma":             "en",
		"formato_data":       "ymd",
		"fuso_horario":       "Europe/Lisbon",
		"visual_biblioteca":  "lista",
		"itens_por_pagina":   50,
		"reduzir_animacoes":  true,
		"modo_tema":          "claro",
		"estilo_tema":        "playstation",
	}
	body, _ := json.Marshal(payload)

	token := gerarTokenTestePreferencias(t, secret, 42, "en")
	req := httptest.NewRequest(http.MethodPut, "/api/v1/me/preferencias", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Data repository.PreferenciasUsuario `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}
	if resp.Data != *retorno {
		t.Fatalf("resposta = %+v, want %+v", resp.Data, *retorno)
	}
}

func TestPreferenciasHandler_401SemToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	mock := preferenciasServiceMock{}
	handler := NewPreferenciasHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/me/preferencias", handler.ObterPreferencias)
	privadas.PUT("/me/preferencias", handler.AtualizarPreferencias)

	t.Run("GET sem token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/me/preferencias", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("PUT sem token", func(t *testing.T) {
		body, _ := json.Marshal(payloadJSONValido())
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/preferencias", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreferenciasHandler_400CamposInvalidos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		nome       string
		serviceErr error
		wantCodigo string
	}{
		{"idioma", service.ErrPreferenciasIdiomaInvalido, "preferencias.idioma_invalido"},
		{"formato_data", service.ErrPreferenciasFormatoDataInvalido, "preferencias.formato_data_invalido"},
		{"fuso_horario", service.ErrPreferenciasFusoHorarioInvalido, "preferencias.fuso_horario_invalido"},
		{"visual_biblioteca", service.ErrPreferenciasVisualBibliotecaInvalido, "preferencias.visual_biblioteca_invalido"},
		{"itens_por_pagina", service.ErrPreferenciasItensPorPaginaInvalido, "preferencias.itens_por_pagina_invalido"},
		{"modo_tema", service.ErrPreferenciasModoTemaInvalido, "preferencias.modo_tema_invalido"},
		{"estilo_tema", service.ErrPreferenciasEstiloTemaInvalido, "preferencias.estilo_tema_invalido"},
	}

	for _, tc := range testCases {
		t.Run(tc.nome, func(t *testing.T) {
			mock := preferenciasServiceMock{
				atualizarPreferencias: func(context.Context, string, service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error) {
					return nil, tc.serviceErr
				},
			}

			handler := NewPreferenciasHandler(mock)
			router := gin.New()
			privadas := middleware.GrupoPrivado(router, tokens)
			privadas.PUT("/me/preferencias", handler.AtualizarPreferencias)

			body, _ := json.Marshal(payloadJSONValido())
			token := gerarTokenTestePreferencias(t, secret, 42, "pt-BR")
			req := httptest.NewRequest(http.MethodPut, "/api/v1/me/preferencias", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}

			var resp struct {
				Error struct {
					Codigo   string `json:"codigo"`
					Mensagem string `json:"mensagem"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("erro ao ler resposta: %v", err)
			}
			if resp.Error.Codigo != tc.wantCodigo {
				t.Fatalf("codigo = %s, want %s", resp.Error.Codigo, tc.wantCodigo)
			}
			if resp.Error.Mensagem != i18n.T("pt-BR", tc.wantCodigo) {
				t.Fatalf("mensagem = %s, want %s", resp.Error.Mensagem, i18n.T("pt-BR", tc.wantCodigo))
			}
		})
	}
}

func TestPreferenciasHandler_400CampoAusente(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	campos := []string{
		"idioma",
		"formato_data",
		"fuso_horario",
		"visual_biblioteca",
		"itens_por_pagina",
		"reduzir_animacoes",
		"modo_tema",
		"estilo_tema",
	}

	mock := preferenciasServiceMock{
		atualizarPreferencias: func(ctx context.Context, subject string, input service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error) {
			return nil, service.ErrPreferenciasEntradaInvalida
		},
	}
	handler := NewPreferenciasHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.PUT("/me/preferencias", handler.AtualizarPreferencias)

	for _, campo := range campos {
		t.Run("sem "+campo, func(t *testing.T) {
			payload := payloadJSONValido()
			delete(payload, campo)

			body, _ := json.Marshal(payload)
			token := gerarTokenTestePreferencias(t, secret, 42, "pt-BR")
			req := httptest.NewRequest(http.MethodPut, "/api/v1/me/preferencias", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			var resp struct {
				Error struct {
					Codigo string `json:"codigo"`
				} `json:"error"`
			}
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Error.Codigo != "preferencias.entrada_invalida" {
				t.Fatalf("codigo = %s, want preferencias.entrada_invalida", resp.Error.Codigo)
			}
		})
	}
}

func TestPreferenciasHandler_400CampoDesconhecidoEJSONMalformado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	mock := preferenciasServiceMock{
		atualizarPreferencias: func(context.Context, string, service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error) {
			t.Fatal("servico nao deveria ser chamado")
			return nil, nil
		},
	}
	handler := NewPreferenciasHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.PUT("/me/preferencias", handler.AtualizarPreferencias)

	token := gerarTokenTestePreferencias(t, secret, 42, "pt-BR")

	t.Run("campo desconhecido", func(t *testing.T) {
		payload := payloadJSONValido()
		payload["campo_extra"] = "invalido"
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/preferencias", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
		var resp struct {
			Error struct {
				Codigo string `json:"codigo"`
			} `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Error.Codigo != "preferencias.entrada_invalida" {
			t.Fatalf("codigo = %s, want preferencias.entrada_invalida", resp.Error.Codigo)
		}
	})

	t.Run("json malformado", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/preferencias", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
		var resp struct {
			Error struct {
				Codigo string `json:"codigo"`
			} `json:"error"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Error.Codigo != "preferencias.entrada_invalida" {
			t.Fatalf("codigo = %s, want preferencias.entrada_invalida", resp.Error.Codigo)
		}
	})
}

func TestPreferenciasHandler_MensagensConformeIdioma(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	mock := preferenciasServiceMock{
		atualizarPreferencias: func(context.Context, string, service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error) {
			return nil, service.ErrPreferenciasFusoHorarioInvalido
		},
	}
	handler := NewPreferenciasHandler(mock)
	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.PUT("/me/preferencias", handler.AtualizarPreferencias)

	for _, lang := range []string{"pt-BR", "en"} {
		t.Run("idioma "+lang, func(t *testing.T) {
			token := gerarTokenTestePreferencias(t, secret, 42, lang)
			body, _ := json.Marshal(payloadJSONValido())
			req := httptest.NewRequest(http.MethodPut, "/api/v1/me/preferencias", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			var resp struct {
				Error struct {
					Codigo   string `json:"codigo"`
					Mensagem string `json:"mensagem"`
				} `json:"error"`
			}
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Error.Codigo != "preferencias.fuso_horario_invalido" {
				t.Fatalf("codigo = %s, want preferencias.fuso_horario_invalido", resp.Error.Codigo)
			}
			wantMsg := i18n.T(lang, "preferencias.fuso_horario_invalido")
			if resp.Error.Mensagem != wantMsg {
				t.Fatalf("mensagem = %s, want %s", resp.Error.Mensagem, wantMsg)
			}
		})
	}
}
