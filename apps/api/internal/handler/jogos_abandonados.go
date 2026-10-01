// Package handler contem os handlers HTTP da API.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type JogosAbandonadosServicer interface {
	CriarJogoAbandonado(
		ctx context.Context,
		input service.SalvarJogoAbandonadoInput,
	) (*repository.JogoAbandonado, error)
	AtualizarJogoAbandonado(
		ctx context.Context,
		input service.SalvarJogoAbandonadoInput,
	) (*repository.JogoAbandonado, error)
	ExcluirJogoAbandonado(ctx context.Context, id int32, usuarioID int32) error
	ListarJogosAbandonados(
		ctx context.Context,
		params service.ListarJogosAbandonadosInput,
	) (*service.ResultadoListagemAbandonados, error)
	ObterJogoAbandonado(ctx context.Context, id int32, usuarioID int32) (*repository.JogoAbandonado, error)
	ObterConsoles(ctx context.Context, usuarioID int32) ([]string, error)
	ObterTotal(ctx context.Context, usuarioID int32) (int64, error)
}

type JogosAbandonadosHandler struct {
	service JogosAbandonadosServicer
}

func NewJogosAbandonadosHandler(service JogosAbandonadosServicer) *JogosAbandonadosHandler {
	return &JogosAbandonadosHandler{service: service}
}

type SalvarJogoAbandonadoRequest struct {
	Nome                string  `json:"nome"`
	Console             string  `json:"console"`
	IgdbID              *int32  `json:"igdb_id"`
	IgdbCapaURL         *string `json:"igdb_capa_url"`
	TempoJogadoHoras    int     `json:"tempo_jogado_horas"`
	TempoJogadoMinutos  int     `json:"tempo_jogado_minutos"`
	TempoJogadoSegundos int     `json:"tempo_jogado_segundos"`
	TempoJogado         *int32  `json:"tempo_jogado"`
	Motivo              *string `json:"motivo"`
	AbandonadoEm        *string `json:"abandonado_em"`
	IniciadoEm          *string `json:"iniciado_em"`
}

func responderErroAbandonados(c *gin.Context, status int, codigo string) {
	lang := c.GetHeader("Accept-Language")
	c.JSON(status, gin.H{
		"error": gin.H{
			"codigo":   codigo,
			"mensagem": i18n.T(lang, codigo),
		},
	})
}

func tratarErroServiceAbandonados(c *gin.Context, err error, logMsg string) {
	switch {
	case errors.Is(err, service.ErrAbandonadoNomeObrigatorio):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.nome_obrigatorio")
	case errors.Is(err, service.ErrAbandonadoNomeMuitoLongo):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.nome_muito_longo")
	case errors.Is(err, service.ErrAbandonadoConsoleObrigatorio):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.console_obrigatorio")
	case errors.Is(err, service.ErrAbandonadoConsoleMuitoLongo):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.console_muito_longo")
	case errors.Is(err, service.ErrAbandonadoTempoInvalido):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.tempo_invalido")
	case errors.Is(err, service.ErrAbandonadoMotivoMuitoLongo):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.motivo_muito_longo")
	case errors.Is(err, service.ErrAbandonadoDataInvalida):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.data_invalida")
	case errors.Is(err, service.ErrAbandonadoDataFutura):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.data_futura")
	case errors.Is(err, service.ErrAbandonadoIniciadoInvalido):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.iniciado_em_invalido")
	case errors.Is(err, service.ErrAbandonadoIniciadoFuturo):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.iniciado_em_futuro")
	case errors.Is(err, service.ErrAbandonadoPaginaInvalida):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.pagina_invalida")
	case errors.Is(err, service.ErrAbandonadoPorPaginaInvalida):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.por_pagina_invalida")
	case errors.Is(err, service.ErrAbandonadoOrdenarInvalido):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.ordenar_invalido")
	case errors.Is(err, service.ErrAbandonadoIdInvalido):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.id_invalido")
	case errors.Is(err, service.ErrAbandonadoEntradaInvalida):
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.entrada_invalida")
	case errors.Is(err, service.ErrAbandonadoNaoEncontrado):
		responderErroAbandonados(c, http.StatusNotFound, "abandonados.nao_encontrado")
	default:
		slog.ErrorContext(c.Request.Context(), logMsg, "error", err)
		responderErroAbandonados(c, http.StatusInternalServerError, "server.internal_error")
	}
}

func parseJogoAbandonadoID(c *gin.Context) (int32, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.id_invalido")
		return 0, false
	}
	return int32(id), true
}

func (h *JogosAbandonadosHandler) Criar(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	var req SalvarJogoAbandonadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.entrada_invalida")
		return
	}

	var abandonadoEmPtr *time.Time
	if req.AbandonadoEm != nil && strings.TrimSpace(*req.AbandonadoEm) != "" {
		t, err := parseDate(*req.AbandonadoEm)
		if err != nil {
			responderErroAbandonados(c, http.StatusBadRequest, "abandonados.data_invalida")
			return
		}
		abandonadoEmPtr = &t
	}
	var iniciadoEmPtr *time.Time
	if req.IniciadoEm != nil && strings.TrimSpace(*req.IniciadoEm) != "" {
		t, err := parseDate(*req.IniciadoEm)
		if err != nil {
			responderErroAbandonados(c, http.StatusBadRequest, "abandonados.iniciado_em_invalido")
			return
		}
		iniciadoEmPtr = &t
	}

	jogo, err := h.service.CriarJogoAbandonado(c.Request.Context(), service.SalvarJogoAbandonadoInput{
		UsuarioID:           int32(usuarioID),
		Nome:                req.Nome,
		Console:             req.Console,
		IgdbID:              req.IgdbID,
		IgdbCapaURL:         req.IgdbCapaURL,
		TempoJogadoHoras:    req.TempoJogadoHoras,
		TempoJogadoMinutos:  req.TempoJogadoMinutos,
		TempoJogadoSegundos: req.TempoJogadoSegundos,
		TempoJogado:         req.TempoJogado,
		Motivo:              req.Motivo,
		AbandonadoEm:        abandonadoEmPtr,
		IniciadoEm:          iniciadoEmPtr,
	})
	if err != nil {
		tratarErroServiceAbandonados(c, err, "falha ao criar jogo abandonado")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": jogo})
}

func (h *JogosAbandonadosHandler) Atualizar(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	id, ok := parseJogoAbandonadoID(c)
	if !ok {
		return
	}

	var req SalvarJogoAbandonadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responderErroAbandonados(c, http.StatusBadRequest, "abandonados.entrada_invalida")
		return
	}

	var abandonadoEmPtr *time.Time
	if req.AbandonadoEm != nil && strings.TrimSpace(*req.AbandonadoEm) != "" {
		t, err := parseDate(*req.AbandonadoEm)
		if err != nil {
			responderErroAbandonados(c, http.StatusBadRequest, "abandonados.data_invalida")
			return
		}
		abandonadoEmPtr = &t
	}
	var iniciadoEmPtr *time.Time
	if req.IniciadoEm != nil && strings.TrimSpace(*req.IniciadoEm) != "" {
		t, err := parseDate(*req.IniciadoEm)
		if err != nil {
			responderErroAbandonados(c, http.StatusBadRequest, "abandonados.iniciado_em_invalido")
			return
		}
		iniciadoEmPtr = &t
	}

	jogo, err := h.service.AtualizarJogoAbandonado(c.Request.Context(), service.SalvarJogoAbandonadoInput{
		ID:                  id,
		UsuarioID:           int32(usuarioID),
		Nome:                req.Nome,
		Console:             req.Console,
		IgdbID:              req.IgdbID,
		IgdbCapaURL:         req.IgdbCapaURL,
		TempoJogadoHoras:    req.TempoJogadoHoras,
		TempoJogadoMinutos:  req.TempoJogadoMinutos,
		TempoJogadoSegundos: req.TempoJogadoSegundos,
		TempoJogado:         req.TempoJogado,
		Motivo:              req.Motivo,
		AbandonadoEm:        abandonadoEmPtr,
		IniciadoEm:          iniciadoEmPtr,
	})
	if err != nil {
		tratarErroServiceAbandonados(c, err, "falha ao atualizar jogo abandonado")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": jogo})
}

func (h *JogosAbandonadosHandler) Excluir(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	id, ok := parseJogoAbandonadoID(c)
	if !ok {
		return
	}

	err = h.service.ExcluirJogoAbandonado(c.Request.Context(), id, int32(usuarioID))
	if err != nil {
		tratarErroServiceAbandonados(c, err, "falha ao excluir jogo abandonado")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *JogosAbandonadosHandler) ObterDetalhes(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	id, ok := parseJogoAbandonadoID(c)
	if !ok {
		return
	}

	jogo, err := h.service.ObterJogoAbandonado(c.Request.Context(), id, int32(usuarioID))
	if err != nil {
		tratarErroServiceAbandonados(c, err, "falha ao obter jogo abandonado")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": jogo})
}

func (h *JogosAbandonadosHandler) Listar(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	var pagina *int
	if pStr := c.Query("pagina"); pStr != "" {
		p, err := strconv.Atoi(pStr)
		if err != nil {
			responderErroAbandonados(c, http.StatusBadRequest, "abandonados.pagina_invalida")
			return
		}
		pagina = &p
	}

	var porPagina *int
	if ppStr := c.Query("por_pagina"); ppStr != "" {
		pp, err := strconv.Atoi(ppStr)
		if err != nil {
			responderErroAbandonados(c, http.StatusBadRequest, "abandonados.por_pagina_invalida")
			return
		}
		porPagina = &pp
	}

	res, err := h.service.ListarJogosAbandonados(c.Request.Context(), service.ListarJogosAbandonadosInput{
		UsuarioID: int32(usuarioID),
		Busca:     c.Query("busca"),
		Console:   c.Query("console"),
		Ordenar:   c.Query("ordenar"),
		Pagina:    pagina,
		PorPagina: porPagina,
	})
	if err != nil {
		tratarErroServiceAbandonados(c, err, "falha ao listar jogos abandonados")
		return
	}

	jogos := res.Jogos
	if jogos == nil {
		jogos = []*repository.JogoAbandonado{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": jogos,
		"meta": gin.H{
			"pagina":        res.Pagina,
			"por_pagina":    res.PorPagina,
			"total":         res.Total,
			"total_paginas": res.TotalPaginas,
		},
	})
}

func (h *JogosAbandonadosHandler) ObterFiltros(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	consoles, err := h.service.ObterConsoles(c.Request.Context(), int32(usuarioID))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "falha ao obter consoles abandonados", "error", err)
		responderErroAbandonados(c, http.StatusInternalServerError, "server.internal_error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"consoles": consoles,
		},
	})
}

func (h *JogosAbandonadosHandler) ObterTotal(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		responderErroAbandonados(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}

	total, err := h.service.ObterTotal(c.Request.Context(), int32(usuarioID))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "falha ao obter total abandonados", "error", err)
		responderErroAbandonados(c, http.StatusInternalServerError, "server.internal_error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total": total,
		},
	})
}
