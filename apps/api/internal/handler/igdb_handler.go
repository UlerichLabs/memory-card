package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type IGDBServicer interface {
	BuscarJogos(ctx context.Context, termo string) ([]igdbclient.Game, error)
	BuscarJogo(ctx context.Context, id int64) (*igdbclient.Game, error)
	ListarPlataformas(ctx context.Context) ([]igdbclient.Platform, error)
	JogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error)
	AtualizarJogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error)
	BuscarFranquias(ctx context.Context, termo string) ([]igdbclient.Franchise, error)
	JogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error)
	AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error)
}

type IGDBHandler struct {
	service IGDBServicer
}

func NewIGDBHandler(service IGDBServicer) *IGDBHandler {
	return &IGDBHandler{service: service}
}

func (h *IGDBHandler) BuscarJogos(c *gin.Context) {
	result, err := h.service.BuscarJogos(c.Request.Context(), c.Query("q"))
	h.respond(c, result, err)
}

func (h *IGDBHandler) BuscarJogo(c *gin.Context) {
	id, ok := parseIGDBID(c)
	if !ok {
		return
	}
	result, err := h.service.BuscarJogo(c.Request.Context(), id)
	h.respond(c, result, err)
}

func (h *IGDBHandler) ListarPlataformas(c *gin.Context) {
	result, err := h.service.ListarPlataformas(c.Request.Context())
	h.respond(c, result, err)
}

func (h *IGDBHandler) JogosDaPlataforma(c *gin.Context) {
	id, ok := parseIGDBID(c)
	if !ok {
		return
	}
	result, err := h.service.JogosDaPlataforma(c.Request.Context(), id)
	h.respond(c, result, err)
}

func (h *IGDBHandler) AtualizarJogosDaPlataforma(c *gin.Context) {
	id, ok := parseIGDBID(c)
	if !ok {
		return
	}
	result, err := h.service.AtualizarJogosDaPlataforma(c.Request.Context(), id)
	h.respond(c, result, err)
}

func (h *IGDBHandler) BuscarFranquias(c *gin.Context) {
	result, err := h.service.BuscarFranquias(c.Request.Context(), c.Query("q"))
	h.respond(c, result, err)
}

func (h *IGDBHandler) JogosDaFranquia(c *gin.Context) {
	id, ok := parseIGDBID(c)
	if !ok {
		return
	}
	result, err := h.service.JogosDaFranquia(c.Request.Context(), id)
	h.respond(c, result, err)
}

func (h *IGDBHandler) AtualizarJogosDaFranquia(c *gin.Context) {
	id, ok := parseIGDBID(c)
	if !ok {
		return
	}
	result, err := h.service.AtualizarJogosDaFranquia(c.Request.Context(), id)
	h.respond(c, result, err)
}

func (h *IGDBHandler) respond(c *gin.Context, data any, err error) {
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"data": data})
		return
	}
	status, code := http.StatusBadGateway, "igdb.unavailable"
	switch {
	case errors.Is(err, service.ErrTermoIGDBVazio):
		status, code = http.StatusBadRequest, "igdb.search.empty"
	case errors.Is(err, igdbclient.ErrRateLimited):
		status, code = http.StatusTooManyRequests, "igdb.rate_limited"
	default:
		slog.ErrorContext(c.Request.Context(), "falha no proxy IGDB", "error", err)
	}
	c.JSON(status, gin.H{"error": gin.H{"codigo": code, "mensagem": i18n.T(c.GetHeader("Accept-Language"), code)}})
}

func parseIGDBID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "igdb.invalid_id", "mensagem": i18n.T(c.GetHeader("Accept-Language"), "igdb.invalid_id")}})
		return 0, false
	}
	return id, true
}
