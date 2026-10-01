// Package handler contem os handlers HTTP da API.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type DashboardServicer interface {
	ObterResumo(ctx context.Context, usuarioID int32) (*service.ResumoResponse, error)
	ListarPorAno(ctx context.Context, usuarioID int32) ([]service.PorAnoItem, error)
	ObterRankingPlataformas(ctx context.Context, usuarioID int32, limite int) ([]service.RankingPlataformaItem, error)
	ObterRankingGeneros(ctx context.Context, usuarioID int32, limite int) ([]service.RankingGeneroItem, error)
	ObterBreakdownTipo(ctx context.Context, usuarioID int32, genero string) ([]service.BreakdownTipoItem, error)
	ObterRecordes(ctx context.Context, usuarioID int32) (*service.RecordesResponse, error)
	ObterNotas(ctx context.Context, usuarioID int32) (*service.NotasResponse, error)
	ObterDificuldade(ctx context.Context, usuarioID int32) ([]service.DificuldadeItem, error)
}

type DashboardHandler struct {
	service DashboardServicer
}

func NewDashboardHandler(service DashboardServicer) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func responderErroDashboard(c *gin.Context, status int, codigo string) {
	lang := c.GetHeader("Accept-Language")
	if usuario, ok := middleware.UsuarioDoContexto(c.Request.Context()); ok && usuario.Idioma != "" {
		if lang == "" {
			lang = usuario.Idioma
		}
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"codigo":   codigo,
			"mensagem": i18n.T(lang, codigo),
		},
	})
}

func tratarErroServiceDashboard(c *gin.Context, err error, logMsg string) {
	switch {
	case errors.Is(err, service.ErrDashboardLimiteInvalido):
		responderErroDashboard(c, http.StatusBadRequest, "dashboard.limite_invalido")
	case errors.Is(err, service.ErrDashboardGeneroObrigatorio):
		responderErroDashboard(c, http.StatusBadRequest, "dashboard.genero_obrigatorio")
	default:
		slog.Error(logMsg, "error", err)
		responderErroDashboard(c, http.StatusInternalServerError, "server.internal_error")
	}
}

func extrairUsuarioID(c *gin.Context) (int32, bool) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroDashboard(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return 0, false
	}
	val, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroDashboard(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return 0, false
	}
	return int32(val), true
}

func parseLimite(limiteStr string) (int, error) {
	if limiteStr == "" {
		return 5, nil
	}
	val, err := strconv.Atoi(limiteStr)
	if err != nil || val < 1 || val > 20 {
		return 0, service.ErrDashboardLimiteInvalido
	}
	return val, nil
}

func (h *DashboardHandler) ObterResumo(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	resumo, err := h.service.ObterResumo(c.Request.Context(), usuarioID)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao obter resumo do dashboard")
		return
	}

	c.JSON(http.StatusOK, resumo)
}

func (h *DashboardHandler) ListarPorAno(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	itens, err := h.service.ListarPorAno(c.Request.Context(), usuarioID)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao listar por ano")
		return
	}

	c.JSON(http.StatusOK, itens)
}

func (h *DashboardHandler) ObterRankingPlataformas(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	limite, err := parseLimite(c.Query("limite"))
	if err != nil {
		tratarErroServiceDashboard(c, err, "limite invalido")
		return
	}

	itens, err := h.service.ObterRankingPlataformas(c.Request.Context(), usuarioID, limite)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao obter ranking de plataformas")
		return
	}

	c.JSON(http.StatusOK, itens)
}

func (h *DashboardHandler) ObterRankingGeneros(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	limite, err := parseLimite(c.Query("limite"))
	if err != nil {
		tratarErroServiceDashboard(c, err, "limite invalido")
		return
	}

	itens, err := h.service.ObterRankingGeneros(c.Request.Context(), usuarioID, limite)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao obter ranking de generos")
		return
	}

	c.JSON(http.StatusOK, itens)
}

func (h *DashboardHandler) ObterBreakdownTipo(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	genero := c.Query("genero")
	if strings.TrimSpace(genero) == "" {
		tratarErroServiceDashboard(c, service.ErrDashboardGeneroObrigatorio, "genero obrigatorio")
		return
	}

	itens, err := h.service.ObterBreakdownTipo(c.Request.Context(), usuarioID, genero)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao obter breakdown de tipo")
		return
	}

	c.JSON(http.StatusOK, itens)
}

func (h *DashboardHandler) ObterRecordes(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	recordes, err := h.service.ObterRecordes(c.Request.Context(), usuarioID)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao obter recordes")
		return
	}

	c.JSON(http.StatusOK, recordes)
}

func (h *DashboardHandler) ObterNotas(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	notas, err := h.service.ObterNotas(c.Request.Context(), usuarioID)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao obter notas")
		return
	}

	c.JSON(http.StatusOK, notas)
}

func (h *DashboardHandler) ObterDificuldade(c *gin.Context) {
	usuarioID, ok := extrairUsuarioID(c)
	if !ok {
		return
	}

	dificuldade, err := h.service.ObterDificuldade(c.Request.Context(), usuarioID)
	if err != nil {
		tratarErroServiceDashboard(c, err, "falha ao obter dificuldade")
		return
	}

	c.JSON(http.StatusOK, dificuldade)
}
