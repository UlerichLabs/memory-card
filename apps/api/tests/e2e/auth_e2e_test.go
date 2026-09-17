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

func TestE2E_Cadastro_FluxoCompleto(t *testing.T) {
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

	dupResp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("falha ao reenviar cadastro com mesmo email: %v", err)
	}
	defer dupResp.Body.Close()

	if dupResp.StatusCode != http.StatusConflict {
		t.Fatalf("status code duplicado = %d, esperado %d", dupResp.StatusCode, http.StatusConflict)
	}

	var errResp struct {
		Error struct {
			Codigo   string `json:"codigo"`
			Mensagem string `json:"mensagem"`
		} `json:"error"`
	}
	dupBytes, _ := io.ReadAll(dupResp.Body)
	if err := json.Unmarshal(dupBytes, &errResp); err != nil {
		t.Fatalf("falha ao decodificar erro: %v", err)
	}
	if errResp.Error.Codigo != "auth.register.email_taken" {
		t.Errorf("codigo = %q, esperado 'auth.register.email_taken'", errResp.Error.Codigo)
	}
}
