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

type PerfilServicer interface {
	Perfil(ctx context.Context, subject string) (*repository.Usuario, error)
}

type MeHandler struct {
	service PerfilServicer
}

func NewMeHandler(service PerfilServicer) *MeHandler {
	return &MeHandler{service: service}
}

func (h *MeHandler) Me(c *gin.Context) {
	autenticado, _ := middleware.UsuarioDoContexto(c.Request.Context())
	usuario, err := h.service.Perfil(c.Request.Context(), autenticado.ID)
	if err != nil {
		status, codigo := http.StatusInternalServerError, "server.internal_error"
		if errors.Is(err, service.ErrTokenInvalido) {
			status, codigo = http.StatusUnauthorized, "auth.session.unauthorized"
		} else {
			slog.ErrorContext(c.Request.Context(), "falha ao consultar perfil", "error", err)
		}
		lang := autenticado.Idioma
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
		}
		c.JSON(status, gin.H{"error": gin.H{"codigo": codigo, "mensagem": i18n.T(lang, codigo)}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": usuario})
}
