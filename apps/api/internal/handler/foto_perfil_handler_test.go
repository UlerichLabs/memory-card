// Package handler contem os handlers HTTP da API.
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type fotoPerfilServiceMock struct {
	upload        func(ctx context.Context, subject string, r io.Reader) (*repository.PerfilUsuario, error)
	definirCapa   func(ctx context.Context, subject string, jogoID int32) (*repository.PerfilUsuario, error)
	removerFoto   func(ctx context.Context, subject string) (*repository.PerfilUsuario, error)
	buscarArquivo func(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error)
}

func (m fotoPerfilServiceMock) Upload(ctx context.Context, subject string, r io.Reader) (*repository.PerfilUsuario, error) {
	if m.upload != nil {
		return m.upload(ctx, subject, r)
	}
	return nil, nil
}

func (m fotoPerfilServiceMock) DefinirCapa(ctx context.Context, subject string, jogoID int32) (*repository.PerfilUsuario, error) {
	if m.definirCapa != nil {
		return m.definirCapa(ctx, subject, jogoID)
	}
	return nil, nil
}

func (m fotoPerfilServiceMock) RemoverFoto(ctx context.Context, subject string) (*repository.PerfilUsuario, error) {
	if m.removerFoto != nil {
		return m.removerFoto(ctx, subject)
	}
	return nil, nil
}

func (m fotoPerfilServiceMock) BuscarArquivo(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error) {
	if m.buscarArquivo != nil {
		return m.buscarArquivo(ctx, nome)
	}
	return nil, time.Time{}, service.ErrFotoArquivoNaoEncontrado
}

func setupFotoHandlerTest(t *testing.T, svc FotoPerfilServicer) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()

	const secret = "foto-perfil-jwt-test-secret-at-least-32-bytes"
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("criar tokens: %v", err)
	}

	h := NewFotoPerfilHandler(svc)
	router.GET("/api/v1/avatares/:arquivo", h.ServirArquivo)

	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.PUT("/me/foto", h.Upload)
	privadas.PUT("/me/foto/capa", h.DefinirCapa)
	privadas.DELETE("/me/foto", h.RemoverFoto)

	rawToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: "pt-BR",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(42),
			Issuer:    "memory-card",
			Audience:  jwt.ClaimStrings{"access"},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("gerar token: %v", err)
	}

	return router, rawToken
}

func criarMultipartRequest(t *testing.T, campo, filename string, dados []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if campo != "" {
		part, err := writer.CreateFormFile(campo, filename)
		if err != nil {
			t.Fatalf("criar form file: %v", err)
		}
		if _, err := part.Write(dados); err != nil {
			t.Fatalf("escrever form file: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("fechar writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func TestFotoPerfilHandler_Upload(t *testing.T) {
	t.Run("200 sucesso", func(t *testing.T) {
		url := "/api/v1/avatares/11111111111111111111111111111111.jpg"
		tipo := "upload"
		svc := fotoPerfilServiceMock{
			upload: func(ctx context.Context, subject string, r io.Reader) (*repository.PerfilUsuario, error) {
				return &repository.PerfilUsuario{
					Nome:       "Lucas",
					AvatarURL:  &url,
					AvatarTipo: &tipo,
				}, nil
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		body, ct := criarMultipartRequest(t, "arquivo", "avatar.jpg", []byte("jpeg_fake_content"))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", ct)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status esperado 200, obteve %d; body: %s", rec.Code, rec.Body.String())
		}
		var resp map[string]repository.PerfilUsuario
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decodificar resposta: %v", err)
		}
		if resp["data"].AvatarURL == nil || *resp["data"].AvatarURL != url {
			t.Fatalf("avatar_url inesperado: %+v", resp["data"])
		}
	})

	t.Run("400 arquivo ausente", func(t *testing.T) {
		router, token := setupFotoHandlerTest(t, fotoPerfilServiceMock{})

		body, ct := criarMultipartRequest(t, "outro_campo", "avatar.jpg", []byte("conteudo"))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", ct)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status esperado 400, obteve %d; body: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "foto.arquivo_obrigatorio") {
			t.Fatalf("esperado erro foto.arquivo_obrigatorio, obteve %s", rec.Body.String())
		}
	})

	t.Run("400 imagem corrompida", func(t *testing.T) {
		svc := fotoPerfilServiceMock{
			upload: func(ctx context.Context, subject string, r io.Reader) (*repository.PerfilUsuario, error) {
				return nil, service.ErrFotoImagemInvalida
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		body, ct := criarMultipartRequest(t, "arquivo", "corrupt.jpg", []byte("dados"))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", ct)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status esperado 400, obteve %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "foto.imagem_invalida") {
			t.Fatalf("esperado foto.imagem_invalida, obteve %s", rec.Body.String())
		}
	})

	t.Run("401 sem token", func(t *testing.T) {
		router, _ := setupFotoHandlerTest(t, fotoPerfilServiceMock{})

		body, ct := criarMultipartRequest(t, "arquivo", "avatar.jpg", []byte("dados"))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto", body)
		req.Header.Set("Content-Type", ct)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status esperado 401, obteve %d", rec.Code)
		}
	})

	t.Run("413 acima de 2 MB", func(t *testing.T) {
		svc := fotoPerfilServiceMock{
			upload: func(ctx context.Context, subject string, r io.Reader) (*repository.PerfilUsuario, error) {
				return nil, service.ErrFotoArquivoMuitoGrande
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		body, ct := criarMultipartRequest(t, "arquivo", "enorme.jpg", []byte("grande"))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", ct)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status esperado 413, obteve %d; body: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "foto.arquivo_muito_grande") {
			t.Fatalf("esperado erro foto.arquivo_muito_grande, obteve %s", rec.Body.String())
		}
	})

	t.Run("415 tipo invalido", func(t *testing.T) {
		svc := fotoPerfilServiceMock{
			upload: func(ctx context.Context, subject string, r io.Reader) (*repository.PerfilUsuario, error) {
				return nil, service.ErrFotoTipoNaoSuportado
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		body, ct := criarMultipartRequest(t, "arquivo", "doc.txt", []byte("texto"))
		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", ct)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("status esperado 415, obteve %d; body: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "foto.tipo_nao_suportado") {
			t.Fatalf("esperado erro foto.tipo_nao_suportado, obteve %s", rec.Body.String())
		}
	})
}

func TestFotoPerfilHandler_DefinirCapa(t *testing.T) {
	t.Run("200 sucesso", func(t *testing.T) {
		capa := "https://images.igdb.com/capa.jpg"
		tipo := "jogo"
		svc := fotoPerfilServiceMock{
			definirCapa: func(ctx context.Context, subject string, jogoID int32) (*repository.PerfilUsuario, error) {
				return &repository.PerfilUsuario{
					Nome:       "Lucas",
					AvatarURL:  &capa,
					AvatarTipo: &tipo,
				}, nil
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto/capa", strings.NewReader(`{"jogo_id": 15}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status esperado 200, obteve %d; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("400 jogo_id invalido", func(t *testing.T) {
		router, token := setupFotoHandlerTest(t, fotoPerfilServiceMock{})

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto/capa", strings.NewReader(`{"jogo_id": -5}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status esperado 400, obteve %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "foto.jogo_id_invalido") {
			t.Fatalf("esperado foto.jogo_id_invalido, obteve %s", rec.Body.String())
		}
	})

	t.Run("400 jogo sem capa", func(t *testing.T) {
		svc := fotoPerfilServiceMock{
			definirCapa: func(ctx context.Context, subject string, jogoID int32) (*repository.PerfilUsuario, error) {
				return nil, service.ErrFotoJogoSemCapa
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto/capa", strings.NewReader(`{"jogo_id": 15}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status esperado 400, obteve %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "foto.jogo_sem_capa") {
			t.Fatalf("esperado foto.jogo_sem_capa, obteve %s", rec.Body.String())
		}
	})

	t.Run("401 sem token", func(t *testing.T) {
		router, _ := setupFotoHandlerTest(t, fotoPerfilServiceMock{})

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto/capa", strings.NewReader(`{"jogo_id": 15}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status esperado 401, obteve %d", rec.Code)
		}
	})

	t.Run("404 capa de jogo de outro usuario ou inexistente", func(t *testing.T) {
		svc := fotoPerfilServiceMock{
			definirCapa: func(ctx context.Context, subject string, jogoID int32) (*repository.PerfilUsuario, error) {
				return nil, service.ErrFotoJogoNaoEncontrado
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto/capa", strings.NewReader(`{"jogo_id": 999}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status esperado 404, obteve %d; body: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "foto.jogo_nao_encontrado") {
			t.Fatalf("esperado foto.jogo_nao_encontrado, obteve %s", rec.Body.String())
		}
	})
}

func TestFotoPerfilHandler_RemoverFoto(t *testing.T) {
	t.Run("200 sucesso", func(t *testing.T) {
		svc := fotoPerfilServiceMock{
			removerFoto: func(ctx context.Context, subject string) (*repository.PerfilUsuario, error) {
				return &repository.PerfilUsuario{
					Nome:       "Lucas",
					AvatarURL:  nil,
					AvatarTipo: nil,
				}, nil
			},
		}
		router, token := setupFotoHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/foto", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status esperado 200, obteve %d", rec.Code)
		}
	})

	t.Run("401 sem token", func(t *testing.T) {
		router, _ := setupFotoHandlerTest(t, fotoPerfilServiceMock{})

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/foto", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status esperado 401, obteve %d", rec.Code)
		}
	})
}

type nopCloserReadSeeker struct {
	*bytes.Reader
}

func (n *nopCloserReadSeeker) Close() error {
	return nil
}

func TestFotoPerfilHandler_ServirArquivo(t *testing.T) {
	nomeValido := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg"

	t.Run("200 publico com headers obrigatorios", func(t *testing.T) {
		dados := []byte("jpeg_bytes")
		svc := fotoPerfilServiceMock{
			buscarArquivo: func(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error) {
				if nome != nomeValido {
					return nil, time.Time{}, service.ErrFotoArquivoNaoEncontrado
				}
				return &nopCloserReadSeeker{Reader: bytes.NewReader(dados)}, time.Unix(1700000000, 0), nil
			},
		}
		router, _ := setupFotoHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/avatares/"+nomeValido, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status esperado 200, obteve %d; body: %s", rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("Cache-Control esperado 'public, max-age=31536000, immutable', obteve %q", got)
		}
		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("X-Content-Type-Options esperado 'nosniff', obteve %q", got)
		}
		if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "image/jpeg") {
			t.Fatalf("Content-Type esperado 'image/jpeg', obteve %q", got)
		}
		if !bytes.Equal(rec.Body.Bytes(), dados) {
			t.Fatalf("conteudo servido incorreto")
		}
	})

	t.Run("404 arquivo inexistente", func(t *testing.T) {
		svc := fotoPerfilServiceMock{
			buscarArquivo: func(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error) {
				return nil, time.Time{}, service.ErrFotoArquivoNaoEncontrado
			},
		}
		router, _ := setupFotoHandlerTest(t, svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/avatares/"+nomeValido, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status esperado 404, obteve %d", rec.Code)
		}
	})

	t.Run("404 nome fora do padrao regex", func(t *testing.T) {
		router, _ := setupFotoHandlerTest(t, fotoPerfilServiceMock{})

		casos := []string{
			"curto.jpg",
			"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.jpg",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg.exe",
		}

		for _, nome := range casos {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/avatares/"+nome, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("para %q esperado 404, obteve %d", nome, rec.Code)
			}
		}
	})

	t.Run("404 directory traversal com pontos e encoded", func(t *testing.T) {
		router, _ := setupFotoHandlerTest(t, fotoPerfilServiceMock{})

		casos := []string{
			"/api/v1/avatares/..",
			"/api/v1/avatares/../etc/passwd",
			"/api/v1/avatares/%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		}

		for _, url := range casos {
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("para %q esperado 404, obteve %d", url, rec.Code)
			}
		}
	})
}
