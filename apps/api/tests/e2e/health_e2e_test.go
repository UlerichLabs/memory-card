//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/UlerichLabs/memory-card/apps/api/internal/testutil"
)

type healthResponse struct {
	Status string `json:"status"`
}

func setupTestServer(t *testing.T, healthService *service.HealthService) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(gin.Recovery())

	healthHandler := handler.NewHealthHandler(healthService)
	router.GET("/api/v1/health", healthHandler.Check)

	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

func TestE2E_Health_BancoSaudavel(t *testing.T) {
	pg := testutil.SetupPostgres(t)

	healthService := service.NewHealthService(pg.Pool)
	server := setupTestServer(t, healthService)

	// Valida que as migrations do golang-migrate foram aplicadas no container real
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	query := "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'usuarios'"
	if err := pg.Pool.QueryRow(ctx, query).Scan(&count); err != nil {
		t.Fatalf("falha ao consultar tabela usuarios no postgres: %v", err)
	}
	if count != 1 {
		t.Fatalf("esperava 1 tabela 'usuarios' criada pelas migrations, obteve %d", count)
	}

	// Executa requisição HTTP real contra o servidor HTTP
	resp, err := http.Get(server.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("falha ao enviar GET /api/v1/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusOK)
	}

	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, esperado '*'", origin)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("falha ao ler corpo da resposta: %v", err)
	}

	var data healthResponse
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("falha ao decodificar JSON de resposta (%s): %v", string(body), err)
	}

	if data.Status != "ok" {
		t.Errorf("status = %q, esperado 'ok'", data.Status)
	}
}

func TestE2E_Health_BancoIndisponivel(t *testing.T) {
	pg := testutil.SetupPostgres(t)

	healthService := service.NewHealthService(pg.Pool)
	server := setupTestServer(t, healthService)

	// Interrompe o container PostgreSQL para simular indisponibilidade real de rede/infra
	stopTimeout := 5 * time.Second
	if err := pg.Container.Stop(context.Background(), &stopTimeout); err != nil {
		t.Fatalf("falha ao parar container postgres: %v", err)
	}

	resp, err := http.Get(server.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("falha ao enviar GET /api/v1/health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusServiceUnavailable)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("falha ao ler corpo da resposta: %v", err)
	}

	var data healthResponse
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("falha ao decodificar JSON: %v", err)
	}

	if data.Status != "unavailable" {
		t.Errorf("status = %q, esperado 'unavailable'", data.Status)
	}
}
