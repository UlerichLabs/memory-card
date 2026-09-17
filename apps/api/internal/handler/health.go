// Package handler implementa os handlers HTTP da API.
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthChecker define o contrato do serviço de verificação de integridade da API.
type HealthChecker interface {
	CheckDatabase(ctx context.Context) error
}

// HealthHandler processa requisições de verificação de saúde da aplicação.
type HealthHandler struct {
	service HealthChecker
}

// NewHealthHandler cria uma nova instância de HealthHandler.
func NewHealthHandler(service HealthChecker) *HealthHandler {
	return &HealthHandler{service: service}
}

// Check executa o health check e responde com o status dos serviços.
func (h *HealthHandler) Check(c *gin.Context) {
	// Apenas este endpoint público permite leitura por outras origens.
	// Não habilita CORS ou credenciais nas futuras rotas autenticadas.
	c.Header("Access-Control-Allow-Origin", "*")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.service.CheckDatabase(ctx); err != nil {
		slog.ErrorContext(ctx, "health check: banco indisponível", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Health retorna uma função compatível com gin.HandlerFunc para o health check.
func Health(service HealthChecker) gin.HandlerFunc {
	h := NewHealthHandler(service)
	return h.Check
}
