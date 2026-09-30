package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type CatalogoServicer interface {
	Jogos(context.Context, int32, service.CatalogoFiltro) (*service.CatalogoResultado, error)
	Generos(context.Context) ([]service.GeneroCatalogo, error)
}

type CatalogoHandler struct {
	service CatalogoServicer
}

func NewCatalogoHandler(catalogoService CatalogoServicer) *CatalogoHandler {
	return &CatalogoHandler{service: catalogoService}
}

func (h *CatalogoHandler) Jogos(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}
	filtro, err := parseCatalogoFiltro(c)
	if err != nil {
		h.erro(c, err)
		return
	}
	resultado, err := h.service.Jogos(c.Request.Context(), usuarioID, filtro)
	if err != nil {
		h.erro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resultado})
}

func (h *CatalogoHandler) Generos(c *gin.Context) {
	if _, ok := extrairUsuarioIDListas(c); !ok {
		return
	}
	generos, err := h.service.Generos(c.Request.Context())
	if err != nil {
		h.erro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": generos})
}

func parseCatalogoFiltro(c *gin.Context) (service.CatalogoFiltro, error) {
	origem := service.CatalogoOrigem(c.Query("origem"))
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		return service.CatalogoFiltro{}, service.ErrCatalogoParametroInvalido
	}
	ordenar := c.DefaultQuery("ordenar", "populares")
	pagina, err := queryInt(c, "pagina", 1)
	if err != nil {
		return service.CatalogoFiltro{}, service.ErrCatalogoPaginacaoInvalida
	}
	porPagina, err := queryInt(c, "por_pagina", 60)
	if err != nil {
		return service.CatalogoFiltro{}, service.ErrCatalogoPaginacaoInvalida
	}
	filtro := service.CatalogoFiltro{Origem: origem, ID: id, Busca: c.Query("busca"), Ordenar: ordenar, Pagina: pagina, PorPagina: porPagina, Agora: time.Now()}
	if valor := c.Query("genero_id"); valor != "" {
		generoID, parseErr := strconv.ParseInt(valor, 10, 64)
		if parseErr != nil || generoID <= 0 {
			return service.CatalogoFiltro{}, service.ErrCatalogoParametroInvalido
		}
		filtro.GeneroID = &generoID
	}
	if valor := c.Query("plataforma_id"); valor != "" {
		plataformaID, parseErr := strconv.ParseInt(valor, 10, 64)
		if parseErr != nil || plataformaID <= 0 {
			return service.CatalogoFiltro{}, service.ErrCatalogoParametroInvalido
		}
		filtro.PlataformaID = &plataformaID
	}
	return filtro, nil
}

func queryInt(c *gin.Context, name string, defaultValue int) (int, error) {
	value := c.Query(name)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}

func (h *CatalogoHandler) erro(c *gin.Context, err error) {
	status := http.StatusBadGateway
	codigo := "igdb.unavailable"
	switch {
	case errors.Is(err, service.ErrCatalogoParametroInvalido):
		status, codigo = http.StatusBadRequest, "catalogo.parametro_invalido"
	case errors.Is(err, service.ErrCatalogoPaginacaoInvalida):
		status, codigo = http.StatusBadRequest, "catalogo.paginacao_invalida"
	case errors.Is(err, service.ErrListaFranquiaNaoEncontrada):
		status, codigo = http.StatusNotFound, "listas.franquia_nao_encontrada"
	case errors.Is(err, service.ErrIGDBRateLimit):
		status, codigo = http.StatusTooManyRequests, "igdb.rate_limited"
	case errors.Is(err, service.ErrIGDBIndisponivel):
		status, codigo = http.StatusBadGateway, "igdb.unavailable"
	}
	c.JSON(status, gin.H{"error": gin.H{"codigo": codigo, "mensagem": i18n.T(c.GetHeader("Accept-Language"), codigo)}})
}
