//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/UlerichLabs/memory-card/apps/api/internal/testutil"
)

type apiErrorResponse struct {
	Error struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	} `json:"error"`
}

func TestE2E_Cadastro(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pg := testutil.SetupPostgres(t)

	queries := db.New(pg.Pool)
	usuarioRepo := repository.NewUsuarioRepository(queries)
	cadastroService := service.NewCadastroService(usuarioRepo)
	authHandler := handler.NewAuthHandler(cadastroService)

	router := gin.New()
	router.POST("/api/v1/auth/register", authHandler.Register)

	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
	})

	t.Run("fluxo completo com persistencia real no banco", func(t *testing.T) {
		payload := map[string]string{
			"nome":  "Lucas",
			"email": "lucas@example.com",
			"senha": "SenhaForte@123",
		}
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("falha ao serializar json: %v", err)
		}

		resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao enviar POST /api/v1/auth/register: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusCreated)
		}

		var successResp struct {
			Data repository.Usuario `json:"data"`
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("falha ao ler corpo da resposta: %v", err)
		}
		if err := json.Unmarshal(respBytes, &successResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de sucesso: %v", err)
		}

		if successResp.Data.ID <= 0 || successResp.Data.Email != "lucas@example.com" {
			t.Fatalf("dados do usuario incorretos na resposta: %+v", successResp.Data)
		}

		var dbID int32
		var dbSenhaHash string
		var dbIdioma string
		query := "SELECT id, senha_hash, idioma FROM usuarios WHERE email = $1"
		err = pg.Pool.QueryRow(context.Background(), query, "lucas@example.com").Scan(&dbID, &dbSenhaHash, &dbIdioma)
		if err != nil {
			t.Fatalf("falha ao consultar usuario persistido no banco: %v", err)
		}

		if dbID != successResp.Data.ID {
			t.Errorf("dbID = %d, responseID = %d", dbID, successResp.Data.ID)
		}
		if dbIdioma != "pt-BR" {
			t.Errorf("dbIdioma = %q, esperado 'pt-BR'", dbIdioma)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(dbSenhaHash), []byte("SenhaForte@123")); err != nil {
			t.Fatalf("hash da senha no banco invalido: %v", err)
		}
	})

	t.Run("email duplicado retorna 409", func(t *testing.T) {
		payload := map[string]string{
			"nome":  "Lucas Duplicado",
			"email": "lucas@example.com",
			"senha": "SenhaForte@123",
		}
		bodyBytes, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao enviar POST com email duplicado: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusConflict)
		}

		var errResp apiErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de erro: %v", err)
		}
		if errResp.Error.Codigo != "auth.register.email_taken" {
			t.Errorf("codigo = %q, esperado 'auth.register.email_taken'", errResp.Error.Codigo)
		}
		if errResp.Error.Mensagem != "Este email já está em uso." {
			t.Errorf("mensagem = %q, esperado 'Este email já está em uso.'", errResp.Error.Mensagem)
		}
	})

	t.Run("senha fraca retorna 400 sem persistir no banco", func(t *testing.T) {
		emailFraco := "senhafraca@example.com"
		payload := map[string]string{
			"nome":  "Usuario Fraco",
			"email": emailFraco,
			"senha": "123",
		}
		bodyBytes, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao enviar POST com senha fraca: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusBadRequest)
		}

		var errResp apiErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de erro: %v", err)
		}
		if errResp.Error.Codigo != "auth.register.weak_password" {
			t.Errorf("codigo = %q, esperado 'auth.register.weak_password'", errResp.Error.Codigo)
		}
		if errResp.Error.Mensagem != "A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial." {
			t.Errorf("mensagem inesperada: %q", errResp.Error.Mensagem)
		}

		var count int
		err = pg.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM usuarios WHERE email = $1", emailFraco).Scan(&count)
		if err != nil {
			t.Fatalf("falha ao verificar persistencia no banco: %v", err)
		}
		if count != 0 {
			t.Fatalf("usuario com senha fraca nao deveria ter sido persistido no banco")
		}
	})

	t.Run("accept language en retorna mensagem em ingles", func(t *testing.T) {
		payload := map[string]string{
			"nome":  "English User",
			"email": "en-weak@example.com",
			"senha": "weak",
		}
		bodyBytes, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/register", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao criar requisicao: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Language", "en")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("falha ao enviar requisicao com Accept-Language: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusBadRequest)
		}

		var errResp apiErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de erro: %v", err)
		}
		if errResp.Error.Codigo != "auth.register.weak_password" {
			t.Errorf("codigo = %q, esperado 'auth.register.weak_password'", errResp.Error.Codigo)
		}
		if errResp.Error.Mensagem != "Password must be at least 8 characters long, including an uppercase letter, a number, and a special character." {
			t.Errorf("mensagem em ingles = %q, inesperada", errResp.Error.Mensagem)
		}

		dupPayload := map[string]string{
			"nome":  "English User 2",
			"email": "lucas@example.com",
			"senha": "SenhaForte@123",
		}
		dupBytes, _ := json.Marshal(dupPayload)
		dupReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/register", bytes.NewReader(dupBytes))
		dupReq.Header.Set("Content-Type", "application/json")
		dupReq.Header.Set("Accept-Language", "en-US")

		dupResp, err := http.DefaultClient.Do(dupReq)
		if err != nil {
			t.Fatalf("falha ao enviar requisicao de duplicidade com Accept-Language: %v", err)
		}
		defer dupResp.Body.Close()

		if dupResp.StatusCode != http.StatusConflict {
			t.Fatalf("status code = %d, esperado %d", dupResp.StatusCode, http.StatusConflict)
		}

		var dupErrResp apiErrorResponse
		dupBody, _ := io.ReadAll(dupResp.Body)
		_ = json.Unmarshal(dupBody, &dupErrResp)
		if dupErrResp.Error.Codigo != "auth.register.email_taken" {
			t.Errorf("codigo = %q, esperado 'auth.register.email_taken'", dupErrResp.Error.Codigo)
		}
		if dupErrResp.Error.Mensagem != "This email is already in use." {
			t.Errorf("mensagem em ingles duplicado = %q, inesperada", dupErrResp.Error.Mensagem)
		}
	})
}
