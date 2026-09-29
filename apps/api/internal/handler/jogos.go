package handler

import (
	"context"
	"errors"
	"fmt"
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

type JogosServicer interface {
	CriarJogoZerado(ctx context.Context, params repository.CriarJogoZeradoParams) (*repository.JogoZerado, error)
	AtualizarJogoZerado(ctx context.Context, params repository.AtualizarJogoZeradoParams) (*repository.JogoZerado, error)
	ExcluirJogoZerado(ctx context.Context, id int32, usuarioID int32) error
	ListarJogosZerados(ctx context.Context, params service.ListarJogosParams) (*service.ResultadoListagem, error)
	ObterOpcoesFiltros(ctx context.Context, usuarioID int32) (*repository.OpcoesFiltros, error)
	ObterDetalhesJogoZerado(ctx context.Context, id int32, usuarioID int32) (*repository.JogoZerado, error)
	ObterResumoGameDoAno(ctx context.Context, usuarioID int32) ([]*repository.ItemResumoGameDoAno, error)
	DefinirGameDoAno(ctx context.Context, id int32, usuarioID int32) (*repository.DefinirGameDoAnoResultado, error)
	RemoverGameDoAno(ctx context.Context, id int32, usuarioID int32) error
}

type JogosHandler struct {
	service JogosServicer
}

func NewJogosHandler(service JogosServicer) *JogosHandler {
	return &JogosHandler{service: service}
}

type CriarJogoRequest struct {
	IgdbID              *int32  `json:"igdb_id"`
	Nome                string  `json:"nome"`
	Console             string  `json:"console"`
	Genero              string  `json:"genero"`
	Tipo                string  `json:"tipo"`
	IniciadoEm          *string `json:"iniciado_em"`
	FinalizadoEm        string  `json:"finalizado_em"`
	TempoJogadoHoras    int     `json:"tempo_jogado_horas"`
	TempoJogadoMinutos  int     `json:"tempo_jogado_minutos"`
	TempoJogadoSegundos int     `json:"tempo_jogado_segundos"`
	TempoJogado         *int32  `json:"tempo_jogado"`
	Nota                int32   `json:"nota"`
	Dificuldade         string  `json:"dificuldade"`
	Review              string  `json:"review"`
	Destaque            bool    `json:"destaque"`
	IgdbCapaURL         string  `json:"igdb_capa_url"`
	IgdbDescricao       string  `json:"igdb_descricao"`
}

func (h *JogosHandler) CriarJogo(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	var req CriarJogoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.invalid_input",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.invalid_input"),
			},
		})
		return
	}

	if strings.TrimSpace(req.FinalizadoEm) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.finalizado_em_required",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.finalizado_em_required"),
			},
		})
		return
	}

	finalizadoEm, err := parseDate(req.FinalizadoEm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.finalizado_em_invalid",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.finalizado_em_invalid"),
			},
		})
		return
	}

	var iniciadoEmPtr *time.Time
	if req.IniciadoEm != nil && strings.TrimSpace(*req.IniciadoEm) != "" {
		t, err := parseDate(*req.IniciadoEm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"codigo":   "jogos.iniciado_em_invalid",
					"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.iniciado_em_invalid"),
				},
			})
			return
		}
		iniciadoEmPtr = &t
	}

	if req.TempoJogadoHoras < 0 || req.TempoJogadoMinutos < 0 || req.TempoJogadoMinutos > 59 || req.TempoJogadoSegundos < 0 || req.TempoJogadoSegundos > 59 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.tempo_jogado_invalid",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.tempo_jogado_invalid"),
			},
		})
		return
	}

	var tempoJogadoSegundos int32
	if req.TempoJogadoHoras > 0 || req.TempoJogadoMinutos > 0 || req.TempoJogadoSegundos > 0 {
		tempoJogadoSegundos = int32(req.TempoJogadoHoras*3600 + req.TempoJogadoMinutos*60 + req.TempoJogadoSegundos)
	} else if req.TempoJogado != nil {
		if *req.TempoJogado < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"codigo":   "jogos.tempo_jogado_invalid",
					"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.tempo_jogado_invalid"),
				},
			})
			return
		}
		tempoJogadoSegundos = *req.TempoJogado
	}

	jogo, err := h.service.CriarJogoZerado(c.Request.Context(), repository.CriarJogoZeradoParams{
		UsuarioID:     int32(usuarioID),
		IgdbID:        req.IgdbID,
		Nome:          req.Nome,
		Console:       req.Console,
		Genero:        req.Genero,
		Tipo:          req.Tipo,
		IniciadoEm:    iniciadoEmPtr,
		FinalizadoEm:  finalizadoEm,
		TempoJogado:   tempoJogadoSegundos,
		Nota:          req.Nota,
		Dificuldade:   req.Dificuldade,
		Review:        req.Review,
		Destaque:      req.Destaque,
		IgdbCapaURL:   req.IgdbCapaURL,
		IgdbDescricao: req.IgdbDescricao,
	})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrNomeObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nome_required", "mensagem": i18n.T(lang, "jogos.nome_required")}})
		case errors.Is(err, service.ErrConsoleObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.console_required", "mensagem": i18n.T(lang, "jogos.console_required")}})
		case errors.Is(err, service.ErrNomeMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nome_muito_longo", "mensagem": i18n.T(lang, "jogos.nome_muito_longo")}})
		case errors.Is(err, service.ErrConsoleMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.console_muito_longo", "mensagem": i18n.T(lang, "jogos.console_muito_longo")}})
		case errors.Is(err, service.ErrGeneroMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.genero_muito_longo", "mensagem": i18n.T(lang, "jogos.genero_muito_longo")}})
		case errors.Is(err, service.ErrTipoMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.tipo_muito_longo", "mensagem": i18n.T(lang, "jogos.tipo_muito_longo")}})
		case errors.Is(err, service.ErrFinalizadoEmObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.finalizado_em_required", "mensagem": i18n.T(lang, "jogos.finalizado_em_required")}})
		case errors.Is(err, service.ErrTempoJogadoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.tempo_jogado_invalid", "mensagem": i18n.T(lang, "jogos.tempo_jogado_invalid")}})
		case errors.Is(err, service.ErrNotaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_invalid", "mensagem": i18n.T(lang, "jogos.nota_invalid")}})
		case errors.Is(err, service.ErrDificuldadeInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.dificuldade_invalid", "mensagem": i18n.T(lang, "jogos.dificuldade_invalid")}})
		case errors.Is(err, service.ErrReviewMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.review_muito_longo", "mensagem": i18n.T(lang, "jogos.review_muito_longo")}})
		case errors.Is(err, service.ErrDestaqueAnoConflito):
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"codigo": "jogos.destaque_ano_conflito", "mensagem": i18n.T(lang, "jogos.destaque_ano_conflito")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao criar jogo zerado", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": jogo})
}

func (h *JogosHandler) AtualizarJogo(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	jogoID, ok := parseJogoID(c)
	if !ok {
		return
	}

	var req CriarJogoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.invalid_input",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.invalid_input"),
			},
		})
		return
	}

	if strings.TrimSpace(req.FinalizadoEm) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.finalizado_em_required",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.finalizado_em_required"),
			},
		})
		return
	}

	finalizadoEm, err := parseDate(req.FinalizadoEm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.finalizado_em_invalid",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.finalizado_em_invalid"),
			},
		})
		return
	}

	var iniciadoEmPtr *time.Time
	if req.IniciadoEm != nil && strings.TrimSpace(*req.IniciadoEm) != "" {
		t, err := parseDate(*req.IniciadoEm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"codigo":   "jogos.iniciado_em_invalid",
					"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.iniciado_em_invalid"),
				},
			})
			return
		}
		iniciadoEmPtr = &t
	}

	if req.TempoJogadoHoras < 0 || req.TempoJogadoMinutos < 0 || req.TempoJogadoMinutos > 59 || req.TempoJogadoSegundos < 0 || req.TempoJogadoSegundos > 59 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.tempo_jogado_invalid",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.tempo_jogado_invalid"),
			},
		})
		return
	}

	var tempoJogadoSegundos int32
	if req.TempoJogadoHoras > 0 || req.TempoJogadoMinutos > 0 || req.TempoJogadoSegundos > 0 {
		tempoJogadoSegundos = int32(req.TempoJogadoHoras*3600 + req.TempoJogadoMinutos*60 + req.TempoJogadoSegundos)
	} else if req.TempoJogado != nil {
		if *req.TempoJogado < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"codigo":   "jogos.tempo_jogado_invalid",
					"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.tempo_jogado_invalid"),
				},
			})
			return
		}
		tempoJogadoSegundos = *req.TempoJogado
	}

	jogo, err := h.service.AtualizarJogoZerado(c.Request.Context(), repository.AtualizarJogoZeradoParams{
		ID:            jogoID,
		UsuarioID:     int32(usuarioID),
		IgdbID:        req.IgdbID,
		Nome:          req.Nome,
		Console:       req.Console,
		Genero:        req.Genero,
		Tipo:          req.Tipo,
		IniciadoEm:    iniciadoEmPtr,
		FinalizadoEm:  finalizadoEm,
		TempoJogado:   tempoJogadoSegundos,
		Nota:          req.Nota,
		Dificuldade:   req.Dificuldade,
		Review:        req.Review,
		Destaque:      req.Destaque,
		IgdbCapaURL:   req.IgdbCapaURL,
		IgdbDescricao: req.IgdbDescricao,
	})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrNomeObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nome_required", "mensagem": i18n.T(lang, "jogos.nome_required")}})
		case errors.Is(err, service.ErrConsoleObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.console_required", "mensagem": i18n.T(lang, "jogos.console_required")}})
		case errors.Is(err, service.ErrNomeMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nome_muito_longo", "mensagem": i18n.T(lang, "jogos.nome_muito_longo")}})
		case errors.Is(err, service.ErrConsoleMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.console_muito_longo", "mensagem": i18n.T(lang, "jogos.console_muito_longo")}})
		case errors.Is(err, service.ErrGeneroMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.genero_muito_longo", "mensagem": i18n.T(lang, "jogos.genero_muito_longo")}})
		case errors.Is(err, service.ErrTipoMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.tipo_muito_longo", "mensagem": i18n.T(lang, "jogos.tipo_muito_longo")}})
		case errors.Is(err, service.ErrFinalizadoEmObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.finalizado_em_required", "mensagem": i18n.T(lang, "jogos.finalizado_em_required")}})
		case errors.Is(err, service.ErrTempoJogadoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.tempo_jogado_invalid", "mensagem": i18n.T(lang, "jogos.tempo_jogado_invalid")}})
		case errors.Is(err, service.ErrNotaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_invalid", "mensagem": i18n.T(lang, "jogos.nota_invalid")}})
		case errors.Is(err, service.ErrDificuldadeInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.dificuldade_invalid", "mensagem": i18n.T(lang, "jogos.dificuldade_invalid")}})
		case errors.Is(err, service.ErrReviewMuitoLongo):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.review_muito_longo", "mensagem": i18n.T(lang, "jogos.review_muito_longo")}})
		case errors.Is(err, service.ErrDestaqueAnoConflito):
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"codigo": "jogos.destaque_ano_conflito", "mensagem": i18n.T(lang, "jogos.destaque_ano_conflito")}})
		case errors.Is(err, service.ErrJogoNaoEncontrado):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "jogos.not_found", "mensagem": i18n.T(lang, "jogos.not_found")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao atualizar jogo zerado", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": jogo})
}

func (h *JogosHandler) ExcluirJogo(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	jogoID, ok := parseJogoID(c)
	if !ok {
		return
	}

	err = h.service.ExcluirJogoZerado(c.Request.Context(), jogoID, int32(usuarioID))
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		if errors.Is(err, service.ErrJogoNaoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"codigo":   "jogos.not_found",
					"mensagem": i18n.T(lang, "jogos.not_found"),
				},
			})
			return
		}
		slog.ErrorContext(c.Request.Context(), "falha ao excluir jogo zerado", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"codigo":   "server.internal_error",
				"mensagem": i18n.T(lang, "server.internal_error"),
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *JogosHandler) ListarJogos(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	lang := c.GetHeader("Accept-Language")

	var pagina *int
	if pStr := c.Query("pagina"); pStr != "" {
		p, err := strconv.Atoi(pStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.pagina_invalida", "mensagem": i18n.T(lang, "jogos.pagina_invalida")}})
			return
		}
		pagina = &p
	}

	var porPagina *int
	if ppStr := c.Query("por_pagina"); ppStr != "" {
		pp, err := strconv.Atoi(ppStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.por_pagina_invalido", "mensagem": i18n.T(lang, "jogos.por_pagina_invalido")}})
			return
		}
		porPagina = &pp
	}

	var notaMin *int
	if nmStr := c.Query("nota_min"); nmStr != "" {
		nm, err := strconv.Atoi(nmStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_filtro_invalida", "mensagem": i18n.T(lang, "jogos.nota_filtro_invalida")}})
			return
		}
		notaMin = &nm
	}

	var notaMax *int
	if nmStr := c.Query("nota_max"); nmStr != "" {
		nm, err := strconv.Atoi(nmStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_filtro_invalida", "mensagem": i18n.T(lang, "jogos.nota_filtro_invalida")}})
			return
		}
		notaMax = &nm
	}

	var ano *int
	if aStr := c.Query("ano"); aStr != "" {
		a, err := strconv.Atoi(aStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.ano_invalido", "mensagem": i18n.T(lang, "jogos.ano_invalido")}})
			return
		}
		ano = &a
	}

	params := service.ListarJogosParams{
		UsuarioID:   int32(usuarioID),
		Busca:       c.Query("busca"),
		Console:     c.Query("console"),
		Genero:      c.Query("genero"),
		Tipo:        c.Query("tipo"),
		NotaMin:     notaMin,
		NotaMax:     notaMax,
		Ano:         ano,
		Dificuldade: c.Query("dificuldade"),
		Ordenar:     c.Query("ordenar"),
		Pagina:      pagina,
		PorPagina:   porPagina,
	}

	res, err := h.service.ListarJogosZerados(c.Request.Context(), params)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPaginaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.pagina_invalida", "mensagem": i18n.T(lang, "jogos.pagina_invalida")}})
		case errors.Is(err, service.ErrPorPaginaInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.por_pagina_invalido", "mensagem": i18n.T(lang, "jogos.por_pagina_invalido")}})
		case errors.Is(err, service.ErrNotaFiltroInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_filtro_invalida", "mensagem": i18n.T(lang, "jogos.nota_filtro_invalida")}})
		case errors.Is(err, service.ErrNotaFaixaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_faixa_invalida", "mensagem": i18n.T(lang, "jogos.nota_faixa_invalida")}})
		case errors.Is(err, service.ErrAnoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.ano_invalido", "mensagem": i18n.T(lang, "jogos.ano_invalido")}})
		case errors.Is(err, service.ErrDificuldadeInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.dificuldade_invalida", "mensagem": i18n.T(lang, "jogos.dificuldade_invalida")}})
		case errors.Is(err, service.ErrOrdenacaoInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.ordenacao_invalida", "mensagem": i18n.T(lang, "jogos.ordenacao_invalida")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao listar jogos zerados", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	jogos := res.Jogos
	if jogos == nil {
		jogos = []*repository.JogoZerado{}
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

func (h *JogosHandler) ObterFiltros(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	filtros, err := h.service.ObterOpcoesFiltros(c.Request.Context(), int32(usuarioID))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "falha ao obter opcoes de filtros", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"codigo":   "server.internal_error",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "server.internal_error"),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filtros,
	})
}

func (h *JogosHandler) ObterDetalhesJogo(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	jogoID, ok := parseJogoID(c)
	if !ok {
		return
	}

	jogo, err := h.service.ObterDetalhesJogoZerado(c.Request.Context(), jogoID, int32(usuarioID))
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		if errors.Is(err, service.ErrJogoNaoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"codigo":   "jogos.not_found",
					"mensagem": i18n.T(lang, "jogos.not_found"),
				},
			})
			return
		}
		slog.ErrorContext(c.Request.Context(), "falha ao obter detalhes do jogo zerado", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"codigo":   "server.internal_error",
				"mensagem": i18n.T(lang, "server.internal_error"),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": jogo})
}

func (h *JogosHandler) ObterGameDoAnoResumo(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	resumo, err := h.service.ObterResumoGameDoAno(c.Request.Context(), int32(usuarioID))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "falha ao obter resumo de game do ano", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"codigo":   "server.internal_error",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "server.internal_error"),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resumo})
}

func (h *JogosHandler) DefinirGameDoAno(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	jogoID, ok := parseJogoID(c)
	if !ok {
		return
	}

	res, err := h.service.DefinirGameDoAno(c.Request.Context(), jogoID, int32(usuarioID))
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrJogoNaoEncontrado):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "jogos.not_found", "mensagem": i18n.T(lang, "jogos.not_found")}})
		case errors.Is(err, service.ErrDestaqueAnoConflito):
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"codigo": "jogos.destaque_ano_conflito", "mensagem": i18n.T(lang, "jogos.destaque_ano_conflito")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao definir game do ano", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"ano":         res.Ano,
			"game_do_ano": res.GameDoAno,
			"anterior_id": res.AnteriorID,
		},
	})
}

func (h *JogosHandler) RemoverGameDoAno(c *gin.Context) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	usuarioID, err := strconv.Atoi(usuario.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.session.unauthorized"),
			},
		})
		return
	}

	jogoID, ok := parseJogoID(c)
	if !ok {
		return
	}

	err = h.service.RemoverGameDoAno(c.Request.Context(), jogoID, int32(usuarioID))
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		if errors.Is(err, service.ErrJogoNaoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"codigo":   "jogos.not_found",
					"mensagem": i18n.T(lang, "jogos.not_found"),
				},
			})
			return
		}
		slog.ErrorContext(c.Request.Context(), "falha ao remover game do ano", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"codigo":   "server.internal_error",
				"mensagem": i18n.T(lang, "server.internal_error"),
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseJogoID(c *gin.Context) (int32, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "jogos.invalid_id",
				"mensagem": i18n.T(c.GetHeader("Accept-Language"), "jogos.invalid_id"),
			},
		})
		return 0, false
	}
	return int32(id), true
}

func parseDate(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("formato de data invalido: %q", value)
}
