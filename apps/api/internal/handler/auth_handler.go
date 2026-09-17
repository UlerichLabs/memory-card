package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

type CadastroServicer interface {
	Cadastrar(ctx context.Context, nome, email, senha string) (*repository.Usuario, error)
}

type AuthHandler struct {
	service CadastroServicer
}

func NewAuthHandler(service CadastroServicer) *AuthHandler {
	return &AuthHandler{service: service}
}

type registerRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lang := c.GetHeader("Accept-Language")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"codigo":   "auth.register.invalid_input",
				"mensagem": i18n.T(lang, "auth.register.invalid_input"),
			},
		})
		return
	}

	lang := c.GetHeader("Accept-Language")
	usuario, err := h.service.Cadastrar(c.Request.Context(), req.Nome, req.Email, req.Senha)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailJaCadastrado):
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"codigo":   "auth.register.email_taken",
					"mensagem": i18n.T(lang, "auth.register.email_taken"),
				},
			})
		case errors.Is(err, service.ErrSenhaFraca):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"codigo":   "auth.register.weak_password",
					"mensagem": i18n.T(lang, "auth.register.weak_password"),
				},
			})
		case errors.Is(err, service.ErrEmailInvalido):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"codigo":   "auth.register.invalid_email",
					"mensagem": i18n.T(lang, "auth.register.invalid_email"),
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"codigo":   "server.internal_error",
					"mensagem": i18n.T(lang, "server.internal_error"),
				},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": usuario,
	})
}
