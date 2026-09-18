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

type CadastroServicer interface {
	Cadastrar(ctx context.Context, nome, email, senha string) (*repository.Usuario, error)
}

type AuthHandler struct {
	service CadastroServicer
	login   LoginServicer
	logout  LogoutServicer
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

type LoginServicer interface {
	Login(ctx context.Context, email, senha string) (*service.LoginResult, error)
	Refresh(ctx context.Context, token string) (*service.RefreshResult, error)
}

func NewLoginHandler(login LoginServicer) *AuthHandler {
	return &AuthHandler{login: login}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, http.StatusBadRequest, "auth.login.invalid_input")
		return
	}
	result, err := h.login.Login(c.Request.Context(), req.Email, req.Senha)
	if err != nil {
		if errors.Is(err, service.ErrCredenciaisInvalidas) {
			authError(c, http.StatusUnauthorized, "auth.login.invalid_credentials")
		} else {
			slog.ErrorContext(c.Request.Context(), "falha no login", "error", err)
			authError(c, http.StatusInternalServerError, "server.internal_error")
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, http.StatusUnauthorized, "auth.session.expired")
		return
	}
	result, err := h.login.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrSessaoExpirada) {
			authError(c, http.StatusUnauthorized, "auth.session.expired")
		} else {
			slog.ErrorContext(c.Request.Context(), "falha no refresh", "error", err)
			authError(c, http.StatusInternalServerError, "server.internal_error")
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func authError(c *gin.Context, status int, codigo string) {
	lang := c.GetHeader("Accept-Language")
	if usuario, ok := middleware.UsuarioDoContexto(c.Request.Context()); ok {
		lang = usuario.Idioma
	}
	c.JSON(status, gin.H{"error": gin.H{"codigo": codigo, "mensagem": i18n.T(lang, codigo)}})
}

type LogoutServicer interface {
	Logout(ctx context.Context, subject, token string) error
}

func NewLogoutHandler(logout LogoutServicer) *AuthHandler {
	return &AuthHandler{logout: logout}
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, http.StatusUnauthorized, "auth.session.expired")
		return
	}
	usuario, _ := middleware.UsuarioDoContexto(c.Request.Context())
	if err := h.logout.Logout(c.Request.Context(), usuario.ID, req.RefreshToken); err != nil {
		switch {
		case errors.Is(err, service.ErrSessaoExpirada):
			authError(c, http.StatusUnauthorized, "auth.session.expired")
		case errors.Is(err, service.ErrTokenInvalido):
			authError(c, http.StatusUnauthorized, "auth.session.unauthorized")
		default:
			slog.ErrorContext(c.Request.Context(), "falha no logout", "error", err)
			authError(c, http.StatusInternalServerError, "server.internal_error")
		}
		return
	}
	c.Status(http.StatusNoContent)
}
