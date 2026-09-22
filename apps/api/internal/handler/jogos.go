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
	CondicaoZeramento   string  `json:"condicao_zeramento"`
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
		UsuarioID:         int32(usuarioID),
		IgdbID:            req.IgdbID,
		Nome:              req.Nome,
		Console:           req.Console,
		Genero:            req.Genero,
		Tipo:              req.Tipo,
		IniciadoEm:        iniciadoEmPtr,
		FinalizadoEm:      finalizadoEm,
		TempoJogado:       tempoJogadoSegundos,
		Nota:              req.Nota,
		Dificuldade:       req.Dificuldade,
		CondicaoZeramento: req.CondicaoZeramento,
		Destaque:          req.Destaque,
		IgdbCapaURL:       req.IgdbCapaURL,
		IgdbDescricao:     req.IgdbDescricao,
	})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrNomeObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nome_required", "mensagem": i18n.T(lang, "jogos.nome_required")}})
		case errors.Is(err, service.ErrConsoleObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.console_required", "mensagem": i18n.T(lang, "jogos.console_required")}})
		case errors.Is(err, service.ErrFinalizadoEmObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.finalizado_em_required", "mensagem": i18n.T(lang, "jogos.finalizado_em_required")}})
		case errors.Is(err, service.ErrTempoJogadoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.tempo_jogado_invalid", "mensagem": i18n.T(lang, "jogos.tempo_jogado_invalid")}})
		case errors.Is(err, service.ErrNotaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_invalid", "mensagem": i18n.T(lang, "jogos.nota_invalid")}})
		case errors.Is(err, service.ErrDificuldadeInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.dificuldade_invalid", "mensagem": i18n.T(lang, "jogos.dificuldade_invalid")}})
		case errors.Is(err, service.ErrCondicaoZeramentoInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.condicao_zeramento_too_long", "mensagem": i18n.T(lang, "jogos.condicao_zeramento_too_long")}})
		case errors.Is(err, service.ErrDestaqueAnoConflito):
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"codigo": "jogos.destaque_ano_conflito", "mensagem": "já existe um destaque para este ano"}})
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
		ID:                jogoID,
		UsuarioID:         int32(usuarioID),
		IgdbID:            req.IgdbID,
		Nome:              req.Nome,
		Console:           req.Console,
		Genero:            req.Genero,
		Tipo:              req.Tipo,
		IniciadoEm:        iniciadoEmPtr,
		FinalizadoEm:      finalizadoEm,
		TempoJogado:       tempoJogadoSegundos,
		Nota:              req.Nota,
		Dificuldade:       req.Dificuldade,
		CondicaoZeramento: req.CondicaoZeramento,
		Destaque:          req.Destaque,
		IgdbCapaURL:       req.IgdbCapaURL,
		IgdbDescricao:     req.IgdbDescricao,
	})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrNomeObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nome_required", "mensagem": i18n.T(lang, "jogos.nome_required")}})
		case errors.Is(err, service.ErrConsoleObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.console_required", "mensagem": i18n.T(lang, "jogos.console_required")}})
		case errors.Is(err, service.ErrFinalizadoEmObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.finalizado_em_required", "mensagem": i18n.T(lang, "jogos.finalizado_em_required")}})
		case errors.Is(err, service.ErrTempoJogadoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.tempo_jogado_invalid", "mensagem": i18n.T(lang, "jogos.tempo_jogado_invalid")}})
		case errors.Is(err, service.ErrNotaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.nota_invalid", "mensagem": i18n.T(lang, "jogos.nota_invalid")}})
		case errors.Is(err, service.ErrDificuldadeInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.dificuldade_invalid", "mensagem": i18n.T(lang, "jogos.dificuldade_invalid")}})
		case errors.Is(err, service.ErrCondicaoZeramentoInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "jogos.condicao_zeramento_too_long", "mensagem": i18n.T(lang, "jogos.condicao_zeramento_too_long")}})
		case errors.Is(err, service.ErrDestaqueAnoConflito):
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"codigo": "jogos.destaque_ano_conflito", "mensagem": "já existe um destaque para este ano"}})
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
