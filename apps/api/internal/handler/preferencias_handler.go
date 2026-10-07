// Package handler contem os handlers HTTP da API.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type PreferenciasServicer interface {
	ObterPreferencias(ctx context.Context, subject string) (*repository.PreferenciasUsuario, error)
	AtualizarPreferencias(ctx context.Context, subject string, input service.AtualizarPreferenciasInput) (*repository.PreferenciasUsuario, error)
}

type PreferenciasHandler struct {
	service PreferenciasServicer
}

func NewPreferenciasHandler(service PreferenciasServicer) *PreferenciasHandler {
	return &PreferenciasHandler{service: service}
}

func responderErroPreferencias(c *gin.Context, status int, codigo string, idiomaUsuario string) {
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

func tratarErroServicePreferencias(c *gin.Context, err error, idioma string) {
	status := http.StatusBadRequest
	codigo := ""

	switch {
	case errors.Is(err, service.ErrPreferenciasIdiomaInvalido):
		codigo = "preferencias.idioma_invalido"
	case errors.Is(err, service.ErrPreferenciasFormatoDataInvalido):
		codigo = "preferencias.formato_data_invalido"
	case errors.Is(err, service.ErrPreferenciasFusoHorarioInvalido):
		codigo = "preferencias.fuso_horario_invalido"
	case errors.Is(err, service.ErrPreferenciasVisualBibliotecaInvalido):
		codigo = "preferencias.visual_biblioteca_invalido"
	case errors.Is(err, service.ErrPreferenciasItensPorPaginaInvalido):
		codigo = "preferencias.itens_por_pagina_invalido"
	case errors.Is(err, service.ErrPreferenciasModoTemaInvalido):
		codigo = "preferencias.modo_tema_invalido"
	case errors.Is(err, service.ErrPreferenciasEstiloTemaInvalido):
		codigo = "preferencias.estilo_tema_invalido"
	case errors.Is(err, service.ErrPreferenciasEntradaInvalida):
		codigo = "preferencias.entrada_invalida"
	case errors.Is(err, service.ErrTokenInvalido):
		status = http.StatusUnauthorized
		codigo = "auth.session.unauthorized"
	default:
		slog.ErrorContext(c.Request.Context(), "falha no servico de preferencias", "error", err)
		status = http.StatusInternalServerError
		codigo = "server.internal_error"
	}

	responderErroPreferencias(c, status, codigo, idioma)
}

func (h *PreferenciasHandler) ObterPreferencias(c *gin.Context) {
	autenticado, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroPreferencias(c, http.StatusUnauthorized, "auth.session.unauthorized", "")
		return
	}

	pref, err := h.service.ObterPreferencias(c.Request.Context(), autenticado.ID)
	if err != nil {
		tratarErroServicePreferencias(c, err, autenticado.Idioma)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pref})
}

func (h *PreferenciasHandler) AtualizarPreferencias(c *gin.Context) {
	autenticado, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroPreferencias(c, http.StatusUnauthorized, "auth.session.unauthorized", "")
		return
	}

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var input service.AtualizarPreferenciasInput
	if err := decoder.Decode(&input); err != nil {
		responderErroPreferencias(c, http.StatusBadRequest, "preferencias.entrada_invalida", autenticado.Idioma)
		return
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		responderErroPreferencias(c, http.StatusBadRequest, "preferencias.entrada_invalida", autenticado.Idioma)
		return
	}

	pref, err := h.service.AtualizarPreferencias(c.Request.Context(), autenticado.ID, input)
	if err != nil {
		tratarErroServicePreferencias(c, err, autenticado.Idioma)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pref})
}
