// Package handler contem os handlers HTTP da API.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type PerfilCompletoServicer interface {
	ObterPerfil(ctx context.Context, subject string) (*repository.PerfilUsuario, error)
	AtualizarPerfil(ctx context.Context, subject string, input service.AtualizarPerfilInput) (*repository.PerfilUsuario, error)
}

type PerfilHandler struct {
	service PerfilCompletoServicer
}

func NewPerfilHandler(service PerfilCompletoServicer) *PerfilHandler {
	return &PerfilHandler{service: service}
}

func responderErroPerfil(c *gin.Context, status int, codigo string, idiomaUsuario string) {
	idioma := idiomaUsuario
	if idioma == "" {
		idioma = c.GetHeader("Accept-Language")
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"codigo":   codigo,
			"mensagem": i18n.T(idioma, codigo),
		},
	})
}

func tratarErroServicePerfil(c *gin.Context, err error, idioma string) {
	status := http.StatusBadRequest
	codigo := ""

	switch {
	case errors.Is(err, service.ErrPerfilNomeObrigatorio):
		codigo = "perfil.nome_obrigatorio"
	case errors.Is(err, service.ErrPerfilNomeInvalido):
		codigo = "perfil.nome_invalido"
	case errors.Is(err, service.ErrPerfilUsernameInvalido):
		codigo = "perfil.username_invalido"
	case errors.Is(err, service.ErrPerfilUsernameEmUso):
		status = http.StatusConflict
		codigo = "perfil.username_em_uso"
	case errors.Is(err, service.ErrPerfilBioMuitoLonga):
		codigo = "perfil.bio_muito_longa"
	case errors.Is(err, service.ErrPerfilJogoFavoritoNaoEncontrado):
		status = http.StatusNotFound
		codigo = "perfil.jogo_favorito_nao_encontrado"
	case errors.Is(err, service.ErrPerfilConsoleInvalido):
		codigo = "perfil.console_invalido"
	case errors.Is(err, service.ErrPerfilJogandoDesdeInvalido):
		codigo = "perfil.jogando_desde_invalido"
	case errors.Is(err, service.ErrPerfilEntradaInvalida):
		codigo = "perfil.entrada_invalida"
	case errors.Is(err, service.ErrTokenInvalido):
		status = http.StatusUnauthorized
		codigo = "auth.session.unauthorized"
	default:
		slog.ErrorContext(c.Request.Context(), "falha no servico de perfil", "error", err)
		status = http.StatusInternalServerError
		codigo = "server.internal_error"
	}

	responderErroPerfil(c, status, codigo, idioma)
}

func (h *PerfilHandler) ObterPerfil(c *gin.Context) {
	autenticado, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroPerfil(c, http.StatusUnauthorized, "auth.session.unauthorized", "")
		return
	}

	perfil, err := h.service.ObterPerfil(c.Request.Context(), autenticado.ID)
	if err != nil {
		tratarErroServicePerfil(c, err, autenticado.Idioma)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": perfil})
}

func (h *PerfilHandler) AtualizarPerfil(c *gin.Context) {
	autenticado, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroPerfil(c, http.StatusUnauthorized, "auth.session.unauthorized", "")
		return
	}

	var req service.AtualizarPerfilInput
	if err := c.ShouldBindJSON(&req); err != nil {
		responderErroPerfil(c, http.StatusBadRequest, "perfil.entrada_invalida", autenticado.Idioma)
		return
	}

	perfil, err := h.service.AtualizarPerfil(c.Request.Context(), autenticado.ID, req)
	if err != nil {
		tratarErroServicePerfil(c, err, autenticado.Idioma)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": perfil})
}
