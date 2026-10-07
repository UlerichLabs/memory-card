// Package handler contem os handlers HTTP da API.
package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

var regexArquivoValido = regexp.MustCompile(`^[a-f0-9]{32}\.jpg$`)

type FotoPerfilServicer interface {
	Upload(ctx context.Context, subject string, r io.Reader) (*repository.PerfilUsuario, error)
	DefinirCapa(ctx context.Context, subject string, jogoID int32) (*repository.PerfilUsuario, error)
	RemoverFoto(ctx context.Context, subject string) (*repository.PerfilUsuario, error)
	BuscarArquivo(ctx context.Context, nome string) (io.ReadSeekCloser, time.Time, error)
}

type FotoPerfilHandler struct {
	service FotoPerfilServicer
}

func NewFotoPerfilHandler(service FotoPerfilServicer) *FotoPerfilHandler {
	return &FotoPerfilHandler{service: service}
}

type definirCapaBody struct {
	JogoID *int32 `json:"jogo_id"`
}

func responderErroFoto(c *gin.Context, status int, codigo string, idiomaUsuario string) {
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

func tratarErroServiceFoto(c *gin.Context, err error, idioma string) {
	status := http.StatusBadRequest
	codigo := ""

	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.As(err, &maxBytesErr) || errors.Is(err, service.ErrFotoArquivoMuitoGrande):
		status = http.StatusRequestEntityTooLarge
		codigo = "foto.arquivo_muito_grande"
	case errors.Is(err, service.ErrFotoArquivoObrigatorio):
		codigo = "foto.arquivo_obrigatorio"
	case errors.Is(err, service.ErrFotoTipoNaoSuportado):
		status = http.StatusUnsupportedMediaType
		codigo = "foto.tipo_nao_suportado"
	case errors.Is(err, service.ErrFotoImagemInvalida):
		codigo = "foto.imagem_invalida"
	case errors.Is(err, service.ErrFotoDimensoesExcessivas):
		codigo = "foto.dimensoes_excessivas"
	case errors.Is(err, service.ErrFotoJogoIDInvalido):
		codigo = "foto.jogo_id_invalido"
	case errors.Is(err, service.ErrFotoJogoSemCapa):
		codigo = "foto.jogo_sem_capa"
	case errors.Is(err, service.ErrFotoJogoNaoEncontrado):
		status = http.StatusNotFound
		codigo = "foto.jogo_nao_encontrado"
	case errors.Is(err, service.ErrFotoArquivoNaoEncontrado):
		status = http.StatusNotFound
		codigo = "foto.arquivo_nao_encontrado"
	case errors.Is(err, service.ErrTokenInvalido):
		status = http.StatusUnauthorized
		codigo = "auth.session.unauthorized"
	default:
		slog.ErrorContext(c.Request.Context(), "falha no servico de foto de perfil", "error", err)
		status = http.StatusInternalServerError
		codigo = "server.internal_error"
	}

	responderErroFoto(c, status, codigo, idioma)
}

func (h *FotoPerfilHandler) Upload(c *gin.Context) {
	autenticado, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroFoto(c, http.StatusUnauthorized, "auth.session.unauthorized", "")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxFotoTamanho+64*1024)

	fileHeader, err := c.FormFile("arquivo")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			responderErroFoto(c, http.StatusRequestEntityTooLarge, "foto.arquivo_muito_grande", autenticado.Idioma)
			return
		}
		responderErroFoto(c, http.StatusBadRequest, "foto.arquivo_obrigatorio", autenticado.Idioma)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		responderErroFoto(c, http.StatusBadRequest, "foto.arquivo_obrigatorio", autenticado.Idioma)
		return
	}
	defer file.Close()

	limitedReader := http.MaxBytesReader(c.Writer, file, service.MaxFotoTamanho)

	perfil, err := h.service.Upload(c.Request.Context(), autenticado.ID, limitedReader)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			responderErroFoto(c, http.StatusRequestEntityTooLarge, "foto.arquivo_muito_grande", autenticado.Idioma)
			return
		}
		tratarErroServiceFoto(c, err, autenticado.Idioma)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": perfil})
}

func (h *FotoPerfilHandler) DefinirCapa(c *gin.Context) {
	autenticado, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroFoto(c, http.StatusUnauthorized, "auth.session.unauthorized", "")
		return
	}

	var req definirCapaBody
	if err := c.ShouldBindJSON(&req); err != nil || req.JogoID == nil || *req.JogoID <= 0 {
		responderErroFoto(c, http.StatusBadRequest, "foto.jogo_id_invalido", autenticado.Idioma)
		return
	}

	perfil, err := h.service.DefinirCapa(c.Request.Context(), autenticado.ID, *req.JogoID)
	if err != nil {
		tratarErroServiceFoto(c, err, autenticado.Idioma)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": perfil})
}

func (h *FotoPerfilHandler) RemoverFoto(c *gin.Context) {
	autenticado, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		responderErroFoto(c, http.StatusUnauthorized, "auth.session.unauthorized", "")
		return
	}

	perfil, err := h.service.RemoverFoto(c.Request.Context(), autenticado.ID)
	if err != nil {
		tratarErroServiceFoto(c, err, autenticado.Idioma)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": perfil})
}

func (h *FotoPerfilHandler) ServirArquivo(c *gin.Context) {
	arquivo := c.Param("arquivo")
	if !regexArquivoValido.MatchString(arquivo) {
		responderErroFoto(c, http.StatusNotFound, "foto.arquivo_nao_encontrado", c.GetHeader("Accept-Language"))
		return
	}

	f, modTime, err := h.service.BuscarArquivo(c.Request.Context(), arquivo)
	if err != nil {
		if errors.Is(err, service.ErrFotoArquivoNaoEncontrado) {
			responderErroFoto(c, http.StatusNotFound, "foto.arquivo_nao_encontrado", c.GetHeader("Accept-Language"))
			return
		}
		slog.ErrorContext(c.Request.Context(), "falha ao buscar avatar", "arquivo", arquivo, "error", err)
		responderErroFoto(c, http.StatusInternalServerError, "server.internal_error", c.GetHeader("Accept-Language"))
		return
	}
	defer f.Close()

	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Type", "image/jpeg")

	http.ServeContent(c.Writer, c.Request, arquivo, modTime, f)
}
