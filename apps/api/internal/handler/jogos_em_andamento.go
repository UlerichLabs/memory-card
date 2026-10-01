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
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type JogosEmAndamentoServicer interface {
	Criar(ctx context.Context, input service.SalvarJogoEmAndamentoInput) (*repository.JogoEmAndamento, error)
	Listar(ctx context.Context, usuarioID int32) ([]*repository.JogoEmAndamento, error)
	Excluir(ctx context.Context, id int32, usuarioID int32) error
}

type JogosEmAndamentoHandler struct {
	service JogosEmAndamentoServicer
}

func NewJogosEmAndamentoHandler(service JogosEmAndamentoServicer) *JogosEmAndamentoHandler {
	return &JogosEmAndamentoHandler{service: service}
}

type CriarJogoEmAndamentoRequest struct {
	Nome        string  `json:"nome"`
	IgdbID      *int32  `json:"igdb_id"`
	IgdbCapaURL *string `json:"igdb_capa_url"`
	IniciadoEm  *string `json:"iniciado_em"`
}

func responderErroJogando(c *gin.Context, status int, codigo string) {
	c.JSON(status, gin.H{"error": gin.H{
		"codigo":   codigo,
		"mensagem": i18n.T(c.GetHeader("Accept-Language"), codigo),
	}})
}

func tratarErroServiceJogando(c *gin.Context, err error, logMsg string) {
	codigo := ""
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, service.ErrJogandoNomeObrigatorio):
		codigo = "jogando.nome_obrigatorio"
	case errors.Is(err, service.ErrJogandoNomeMuitoLongo):
		codigo = "jogando.nome_muito_longo"
	case errors.Is(err, service.ErrJogandoIniciadoObrigatorio):
		codigo = "jogando.iniciado_em_obrigatorio"
	case errors.Is(err, service.ErrJogandoIniciadoInvalido):
		codigo = "jogando.iniciado_em_invalido"
	case errors.Is(err, service.ErrJogandoIniciadoFuturo):
		codigo = "jogando.iniciado_em_futuro"
	case errors.Is(err, service.ErrJogandoIdInvalido):
		codigo = "jogando.id_invalido"
	case errors.Is(err, service.ErrJogandoNaoEncontrado):
		codigo = "jogando.nao_encontrado"
		status = http.StatusNotFound
	case errors.Is(err, service.ErrJogandoEntradaInvalida):
		codigo = "jogando.entrada_invalida"
	default:
		slog.ErrorContext(c.Request.Context(), logMsg, "error", err)
		codigo = "server.internal_error"
		status = http.StatusInternalServerError
	}
	responderErroJogando(c, status, codigo)
}

func usuarioJogando(c *gin.Context) (int32, bool) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroJogando(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return 0, false
	}
	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroJogando(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return 0, false
	}
	return int32(usuarioID), true
}

func (h *JogosEmAndamentoHandler) Listar(c *gin.Context) {
	usuarioID, ok := usuarioJogando(c)
	if !ok {
		return
	}
	jogos, err := h.service.Listar(c.Request.Context(), usuarioID)
	if err != nil {
		tratarErroServiceJogando(c, err, "falha ao listar jogos em andamento")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": jogos})
}

func (h *JogosEmAndamentoHandler) Criar(c *gin.Context) {
	usuarioID, ok := usuarioJogando(c)
	if !ok {
		return
	}
	var req CriarJogoEmAndamentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responderErroJogando(c, http.StatusBadRequest, "jogando.entrada_invalida")
		return
	}
	if req.IniciadoEm == nil || strings.TrimSpace(*req.IniciadoEm) == "" {
		responderErroJogando(c, http.StatusBadRequest, "jogando.iniciado_em_obrigatorio")
		return
	}
	iniciadoEm, err := parseDate(*req.IniciadoEm)
	if err != nil {
		responderErroJogando(c, http.StatusBadRequest, "jogando.iniciado_em_invalido")
		return
	}
	jogo, err := h.service.Criar(c.Request.Context(), service.SalvarJogoEmAndamentoInput{
		UsuarioID:   usuarioID,
		Nome:        req.Nome,
		IgdbID:      req.IgdbID,
		IgdbCapaURL: req.IgdbCapaURL,
		IniciadoEm:  &iniciadoEm,
	})
	if err != nil {
		tratarErroServiceJogando(c, err, "falha ao criar jogo em andamento")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": jogo})
}

func (h *JogosEmAndamentoHandler) Excluir(c *gin.Context) {
	usuarioID, ok := usuarioJogando(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		responderErroJogando(c, http.StatusBadRequest, "jogando.id_invalido")
		return
	}
	if err := h.service.Excluir(c.Request.Context(), int32(id), usuarioID); err != nil {
		tratarErroServiceJogando(c, err, "falha ao excluir jogo em andamento")
		return
	}
	c.Status(http.StatusNoContent)
}
