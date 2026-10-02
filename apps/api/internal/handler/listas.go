// Package handler contem os handlers HTTP da API.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type ListasServicer interface {
	CriarLista(ctx context.Context, input service.CriarListaInput) (*service.ListaDetalhada, error)
	ObterLista(ctx context.Context, id int64, usuarioID int32) (*service.ListaDetalhada, error)
	ListarListas(ctx context.Context, usuarioID int32) ([]*service.ListaResumo, error)
	ReordenarListas(ctx context.Context, usuarioID int32, listaIDs []int64) ([]*service.ListaResumo, error)
	AtualizarLista(ctx context.Context, input service.AtualizarListaInput) (*service.ListaDetalhada, error)
	ExcluirLista(ctx context.Context, id int64, usuarioID int32) error
	AdicionarItem(ctx context.Context, input service.AdicionarItemInput) (*service.ListaItemDetalhe, error)
	AdicionarItensLote(ctx context.Context, input service.AdicionarItensLoteInput) (*service.AdicionarItensLoteResultado, error)
	ExcluirItem(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error
	RestaurarItem(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error)
	ReordenarItens(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*service.ListaItemDetalhe, error)
	AssociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*service.ListaItemDetalhe, error)
	DesassociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*service.ListaItemDetalhe, error)
	SincronizarFranquia(ctx context.Context, listaID int64, usuarioID int32) (*service.SincronizarResultado, error)
	PreviaDesafioFranquia(ctx context.Context, franquiaID int64, usuarioID int32) (*service.PreviaDesafioResultado, error)
}

type ListasHandler struct {
	service ListasServicer
}

func NewListasHandler(service ListasServicer) *ListasHandler {
	return &ListasHandler{service: service}
}

type CriarListaRegraRequest struct {
	Tipo             string  `json:"tipo"`
	Valor            string  `json:"valor"`
	IgdbID           *int32  `json:"igdb_id"`
	IgdbIDsIgnorados []int32 `json:"igdb_ids_ignorados,omitempty"`
}

type CriarListaRequest struct {
	Tipo      string                   `json:"tipo"`
	Nome      string                   `json:"nome"`
	Descricao *string                  `json:"descricao"`
	Regra     *CriarListaRegraRequest  `json:"regra"`
	Meta      *int                     `json:"meta"`
	Origem    *CriarListaOrigemRequest `json:"origem"`
	Itens     []CriarListaItemRequest  `json:"itens"`
}

type CriarListaOrigemRequest struct {
	Tipo   string `json:"tipo"`
	IgdbID int32  `json:"igdb_id"`
	Nome   string `json:"nome"`
}

type CriarListaItemRequest struct {
	IgdbID        int32   `json:"igdb_id"`
	Nome          string  `json:"nome"`
	IgdbCapaURL   *string `json:"igdb_capa_url"`
	AnoLancamento *int    `json:"ano_lancamento"`
}

type AtualizarListaRequest struct {
	Nome      *string                 `json:"nome"`
	Descricao *string                 `json:"descricao"`
	Meta      *int                    `json:"meta"`
	Tipo      *string                 `json:"tipo"`
	Regra     *CriarListaRegraRequest `json:"regra"`
}

type AdicionarItemRequest struct {
	IgdbID        *int32  `json:"igdb_id"`
	Nome          string  `json:"nome"`
	Console       *string `json:"console"`
	IgdbCapaURL   *string `json:"igdb_capa_url"`
	AnoLancamento *int    `json:"ano_lancamento"`
}

type AdicionarItensLoteRequest struct {
	Itens []CriarListaItemRequest `json:"itens"`
}

type ReordenarItensRequest struct {
	ItemIDs []int64 `json:"item_ids"`
}

type ReordenarListasRequest struct {
	ListaIDs []int64 `json:"lista_ids"`
}

type AssociarZeramentoRequest struct {
	JogoZeradoID int32 `json:"jogo_zerado_id"`
}

func parseListasID(c *gin.Context, paramName string) (int64, bool) {
	val := c.Param(paramName)
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil || id <= 0 {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "listas.id_invalido",
				"mensagem": i18n.T(lang, "listas.id_invalido"),
			},
		})
		return 0, false
	}
	return id, true
}

func extrairUsuarioIDListas(c *gin.Context) (int32, bool) {
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(lang, "auth.session.unauthorized"),
			},
		})
		return 0, false
	}

	uid, err := strconv.Atoi(usuario.ID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"codigo":   "auth.session.unauthorized",
				"mensagem": i18n.T(lang, "auth.session.unauthorized"),
			},
		})
		return 0, false
	}

	return int32(uid), true
}

func respondListasError(c *gin.Context, status int, codigo string) {
	c.JSON(status, gin.H{"error": gin.H{"codigo": codigo, "mensagem": i18n.T(c.GetHeader("Accept-Language"), codigo)}})
}

func (h *ListasHandler) ListarListas(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	listas, err := h.service.ListarListas(c.Request.Context(), usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		slog.ErrorContext(c.Request.Context(), "falha ao listar listas", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"codigo":   "server.internal_error",
				"mensagem": i18n.T(lang, "server.internal_error"),
			},
		})
		return
	}

	if listas == nil {
		listas = []*service.ListaResumo{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": listas,
	})
}

func (h *ListasHandler) CriarLista(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "listas.tipo_invalido",
				"mensagem": i18n.T(lang, "listas.tipo_invalido"),
			},
		})
		return
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		respondListasError(c, http.StatusBadRequest, "listas.tipo_invalido")
		return
	}
	var req CriarListaRequest
	if err := json.Unmarshal(encoded, &req); err != nil {
		respondListasError(c, http.StatusBadRequest, "listas.tipo_invalido")
		return
	}
	if _, found := raw["meta"]; found {
		respondListasError(c, http.StatusBadRequest, "listas.campo_nao_permitido")
		return
	}
	if _, found := raw["regra"]; found {
		respondListasError(c, http.StatusBadRequest, "listas.campo_nao_permitido")
		return
	}

	var regraInput *service.CriarListaRegraInput
	if req.Regra != nil {
		regraInput = &service.CriarListaRegraInput{
			Tipo:             req.Regra.Tipo,
			Valor:            req.Regra.Valor,
			IgdbID:           req.Regra.IgdbID,
			IgdbIDsIgnorados: req.Regra.IgdbIDsIgnorados,
		}
	}
	var origemInput *service.CriarListaOrigemInput
	if req.Origem != nil {
		origemInput = &service.CriarListaOrigemInput{Tipo: req.Origem.Tipo, IgdbID: req.Origem.IgdbID, Nome: req.Origem.Nome}
	}
	itensInput := make([]service.CriarListaItemInput, 0, len(req.Itens))
	for _, item := range req.Itens {
		itensInput = append(itensInput, service.CriarListaItemInput{IgdbID: item.IgdbID, Nome: item.Nome, IgdbCapaURL: item.IgdbCapaURL, AnoLancamento: item.AnoLancamento})
	}

	lista, err := h.service.CriarLista(c.Request.Context(), service.CriarListaInput{
		UsuarioID: usuarioID,
		Tipo:      req.Tipo,
		Nome:      req.Nome,
		Descricao: req.Descricao,
		Regra:     regraInput,
		Meta:      req.Meta,
		Origem:    origemInput,
		Itens:     itensInput,
	})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNomeObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.nome_obrigatorio", "mensagem": i18n.T(lang, "listas.nome_obrigatorio")}})
		case errors.Is(err, service.ErrListaNomeInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.nome_invalido", "mensagem": i18n.T(lang, "listas.nome_invalido")}})
		case errors.Is(err, service.ErrListaDescricaoMuitoLonga):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.descricao_muito_longa", "mensagem": i18n.T(lang, "listas.descricao_muito_longa")}})
		case errors.Is(err, service.ErrListaTipoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.tipo_invalido", "mensagem": i18n.T(lang, "listas.tipo_invalido")}})
		case errors.Is(err, service.ErrListaRegraInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.regra_invalida", "mensagem": i18n.T(lang, "listas.regra_invalida")}})
		case errors.Is(err, service.ErrListaMetaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.meta_invalida", "mensagem": i18n.T(lang, "listas.meta_invalida")}})
		case errors.Is(err, service.ErrListaFranquiaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.franquia_nao_encontrada", "mensagem": i18n.T(lang, "listas.franquia_nao_encontrada")}})
		case errors.Is(err, service.ErrListaFranquiaSemJogos):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.franquia_sem_jogos", "mensagem": i18n.T(lang, "listas.franquia_sem_jogos")}})
		case errors.Is(err, service.ErrListaCampoNaoPermitido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.campo_nao_permitido", "mensagem": i18n.T(lang, "listas.campo_nao_permitido")}})
		case errors.Is(err, service.ErrListaDesafioSemJogos):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.desafio_sem_jogos", "mensagem": i18n.T(lang, "listas.desafio_sem_jogos")}})
		case errors.Is(err, service.ErrListaItensDemais):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.itens_demais", "mensagem": i18n.T(lang, "listas.itens_demais")}})
		case errors.Is(err, service.ErrIGDBRateLimit):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"codigo": "igdb.rate_limited", "mensagem": i18n.T(lang, "igdb.rate_limited")}})
		case errors.Is(err, service.ErrIGDBIndisponivel):
			c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"codigo": "igdb.unavailable", "mensagem": i18n.T(lang, "igdb.unavailable")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao criar lista", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": lista})
}

func (h *ListasHandler) ObterLista(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	lista, err := h.service.ObterLista(c.Request.Context(), id, usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao obter lista", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": lista})
}

func (h *ListasHandler) AtualizarLista(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "listas.nome_invalido",
				"mensagem": i18n.T(lang, "listas.nome_invalido"),
			},
		})
		return
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		respondListasError(c, http.StatusBadRequest, "listas.campo_nao_permitido")
		return
	}
	var req AtualizarListaRequest
	if err := json.Unmarshal(encoded, &req); err != nil {
		respondListasError(c, http.StatusBadRequest, "listas.campo_nao_permitido")
		return
	}
	for _, campo := range []string{"meta", "tipo", "regra", "origem"} {
		if _, found := raw[campo]; found {
			respondListasError(c, http.StatusBadRequest, "listas.campo_nao_permitido")
			return
		}
	}

	var regraInput *service.CriarListaRegraInput
	if req.Regra != nil {
		regraInput = &service.CriarListaRegraInput{
			Tipo:   req.Regra.Tipo,
			Valor:  req.Regra.Valor,
			IgdbID: req.Regra.IgdbID,
		}
	}

	lista, err := h.service.AtualizarLista(c.Request.Context(), service.AtualizarListaInput{
		ID:        id,
		UsuarioID: usuarioID,
		Nome:      req.Nome,
		Descricao: req.Descricao,
		Meta:      req.Meta,
		Tipo:      req.Tipo,
		Regra:     regraInput,
	})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaCampoNaoPermitido), errors.Is(err, service.ErrListaCampoImutavel):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.campo_nao_permitido", "mensagem": i18n.T(lang, "listas.campo_nao_permitido")}})
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		case errors.Is(err, service.ErrListaNomeObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.nome_obrigatorio", "mensagem": i18n.T(lang, "listas.nome_obrigatorio")}})
		case errors.Is(err, service.ErrListaNomeInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.nome_invalido", "mensagem": i18n.T(lang, "listas.nome_invalido")}})
		case errors.Is(err, service.ErrListaDescricaoMuitoLonga):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.descricao_muito_longa", "mensagem": i18n.T(lang, "listas.descricao_muito_longa")}})
		case errors.Is(err, service.ErrListaMetaInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.meta_invalida", "mensagem": i18n.T(lang, "listas.meta_invalida")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao atualizar lista", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": lista})
}

func (h *ListasHandler) ExcluirLista(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	err := h.service.ExcluirLista(c.Request.Context(), id, usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao excluir lista", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ListasHandler) AdicionarItem(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	var req AdicionarItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "listas.nome_invalido",
				"mensagem": i18n.T(lang, "listas.nome_invalido"),
			},
		})
		return
	}

	item, err := h.service.AdicionarItem(c.Request.Context(), service.AdicionarItemInput{
		ListaID:       id,
		UsuarioID:     usuarioID,
		IgdbID:        req.IgdbID,
		Nome:          req.Nome,
		Console:       req.Console,
		IgdbCapaURL:   req.IgdbCapaURL,
		AnoLancamento: req.AnoLancamento,
	})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		case errors.Is(err, service.ErrListaItensNaoPermitidos):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.itens_nao_permitidos", "mensagem": i18n.T(lang, "listas.itens_nao_permitidos")}})
		case errors.Is(err, service.ErrListaNomeObrigatorio):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.nome_obrigatorio", "mensagem": i18n.T(lang, "listas.nome_obrigatorio")}})
		case errors.Is(err, service.ErrListaNomeInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.nome_invalido", "mensagem": i18n.T(lang, "listas.nome_invalido")}})
		case errors.Is(err, service.ErrListaItemDuplicado):
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"codigo": "listas.item_duplicado", "mensagem": i18n.T(lang, "listas.item_duplicado")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao adicionar item", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *ListasHandler) ExcluirItem(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	itemID, ok := parseListasID(c, "itemId")
	if !ok {
		return
	}

	err := h.service.ExcluirItem(c.Request.Context(), itemID, id, usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaItemNaoEncontrado), errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.item_nao_encontrado", "mensagem": i18n.T(lang, "listas.item_nao_encontrado")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao excluir item", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ListasHandler) AdicionarItensLote(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}
	listaID, ok := parseListasID(c, "id")
	if !ok {
		return
	}
	var req AdicionarItensLoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondListasError(c, http.StatusBadRequest, "listas.itens_demais")
		return
	}
	itens := make([]service.CriarListaItemInput, 0, len(req.Itens))
	for _, item := range req.Itens {
		itens = append(itens, service.CriarListaItemInput{IgdbID: item.IgdbID, Nome: item.Nome, IgdbCapaURL: item.IgdbCapaURL, AnoLancamento: item.AnoLancamento})
	}
	resultado, err := h.service.AdicionarItensLote(c.Request.Context(), service.AdicionarItensLoteInput{ListaID: listaID, UsuarioID: usuarioID, Itens: itens})
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		status := http.StatusBadRequest
		codigo := "listas.itens_demais"
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			status, codigo = http.StatusNotFound, "listas.nao_encontrada"
		case errors.Is(err, service.ErrListaNomeObrigatorio):
			codigo = "listas.nome_obrigatorio"
		case errors.Is(err, service.ErrListaNomeInvalido):
			codigo = "listas.nome_invalido"
		case errors.Is(err, service.ErrListaRegraInvalida):
			codigo = "listas.regra_invalida"
		}
		c.JSON(status, gin.H{"error": gin.H{"codigo": codigo, "mensagem": i18n.T(lang, codigo)}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resultado})
}

func (h *ListasHandler) ReordenarItens(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	var req ReordenarItensRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "listas.ordem_invalida",
				"mensagem": i18n.T(lang, "listas.ordem_invalida"),
			},
		})
		return
	}

	itens, err := h.service.ReordenarItens(c.Request.Context(), id, usuarioID, req.ItemIDs)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		case errors.Is(err, service.ErrListaOrdemInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.ordem_invalida", "mensagem": i18n.T(lang, "listas.ordem_invalida")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao reordenar itens", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": itens})
}

func (h *ListasHandler) ReordenarListas(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	var req ReordenarListasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondListasError(c, http.StatusBadRequest, "listas.ordem_listas_invalida")
		return
	}

	listas, err := h.service.ReordenarListas(c.Request.Context(), usuarioID, req.ListaIDs)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		if errors.Is(err, service.ErrListaOrdemListasInvalida) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.ordem_listas_invalida", "mensagem": i18n.T(lang, "listas.ordem_listas_invalida")}})
			return
		}
		slog.ErrorContext(c.Request.Context(), "falha ao reordenar listas", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": listas})
}

func (h *ListasHandler) AssociarJogoZerado(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	itemID, ok := parseListasID(c, "itemId")
	if !ok {
		return
	}

	var req AssociarZeramentoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.JogoZeradoID <= 0 {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "listas.id_invalido",
				"mensagem": i18n.T(lang, "listas.id_invalido"),
			},
		})
		return
	}

	item, err := h.service.AssociarJogoZerado(c.Request.Context(), itemID, id, usuarioID, req.JogoZeradoID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		case errors.Is(err, service.ErrListaItemNaoEncontrado):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.item_nao_encontrado", "mensagem": i18n.T(lang, "listas.item_nao_encontrado")}})
		case errors.Is(err, service.ErrJogoNaoEncontrado):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "jogos.not_found", "mensagem": i18n.T(lang, "jogos.not_found")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao associar zeramento", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *ListasHandler) DesassociarJogoZerado(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	itemID, ok := parseListasID(c, "itemId")
	if !ok {
		return
	}

	item, err := h.service.DesassociarJogoZerado(c.Request.Context(), itemID, id, usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		case errors.Is(err, service.ErrListaItemNaoEncontrado):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.item_nao_encontrado", "mensagem": i18n.T(lang, "listas.item_nao_encontrado")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao desassociar zeramento", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *ListasHandler) SincronizarFranquia(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	res, err := h.service.SincronizarFranquia(c.Request.Context(), id, usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		case errors.Is(err, service.ErrListaSincronizacaoNaoPermitida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.sincronizacao_nao_permitida", "mensagem": i18n.T(lang, "listas.sincronizacao_nao_permitida")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao sincronizar franquia", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        res,
		"adicionados": res.Adicionados,
	})
}

func (h *ListasHandler) RestaurarItem(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	id, ok := parseListasID(c, "id")
	if !ok {
		return
	}

	itemID, ok := parseListasID(c, "itemId")
	if !ok {
		return
	}

	item, err := h.service.RestaurarItem(c.Request.Context(), itemID, id, usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.nao_encontrada", "mensagem": i18n.T(lang, "listas.nao_encontrada")}})
		case errors.Is(err, service.ErrListaRestauracaoNaoPermitida):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.restauracao_nao_permitida", "mensagem": i18n.T(lang, "listas.restauracao_nao_permitida")}})
		case errors.Is(err, service.ErrListaItemNaoEncontrado):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.item_nao_encontrado", "mensagem": i18n.T(lang, "listas.item_nao_encontrado")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao restaurar item", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *ListasHandler) PreviaDesafioFranquia(c *gin.Context) {
	usuarioID, ok := extrairUsuarioIDListas(c)
	if !ok {
		return
	}

	igdbID, ok := parseListasID(c, "igdbId")
	if !ok {
		return
	}

	previa, err := h.service.PreviaDesafioFranquia(c.Request.Context(), igdbID, usuarioID)
	if err != nil {
		lang := c.GetHeader("Accept-Language")
		switch {
		case errors.Is(err, service.ErrListaIDInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.id_invalido", "mensagem": i18n.T(lang, "listas.id_invalido")}})
		case errors.Is(err, service.ErrListaFranquiaNaoEncontrada):
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"codigo": "listas.franquia_nao_encontrada", "mensagem": i18n.T(lang, "listas.franquia_nao_encontrada")}})
		case errors.Is(err, service.ErrListaFranquiaSemJogos):
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"codigo": "listas.franquia_sem_jogos", "mensagem": i18n.T(lang, "listas.franquia_sem_jogos")}})
		case errors.Is(err, service.ErrIGDBRateLimit):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"codigo": "igdb.rate_limited", "mensagem": i18n.T(lang, "igdb.rate_limited")}})
		case errors.Is(err, service.ErrIGDBIndisponivel):
			c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"codigo": "igdb.unavailable", "mensagem": i18n.T(lang, "igdb.unavailable")}})
		default:
			slog.ErrorContext(c.Request.Context(), "falha ao obter previa de desafio da franquia", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"codigo": "server.internal_error", "mensagem": i18n.T(lang, "server.internal_error")}})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": previa})
}
