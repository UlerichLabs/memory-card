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
	reset   ResetSenhaServicer
	troca   TrocaSenhaServicer
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

type TrocaSenhaServicer interface {
	Trocar(ctx context.Context, subject, senhaAtual, novaSenha string) error
}

func NewTrocaSenhaHandler(troca TrocaSenhaServicer) *AuthHandler {
	return &AuthHandler{troca: troca}
}

func (h *AuthHandler) TrocarSenha(c *gin.Context) {
	var req struct {
		SenhaAtual string `json:"senha_atual"`
		NovaSenha  string `json:"nova_senha"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, http.StatusBadRequest, "auth.password_change.invalid_input")
		return
	}
	usuario, ok := middleware.UsuarioDoContexto(c.Request.Context())
	if !ok {
		authError(c, http.StatusUnauthorized, "auth.session.unauthorized")
		return
	}
	if err := h.troca.Trocar(c.Request.Context(), usuario.ID, req.SenhaAtual, req.NovaSenha); err != nil {
		switch {
		case errors.Is(err, service.ErrSenhaAtualIncorreta):
			authError(c, http.StatusBadRequest, "auth.password_change.current_password_invalid")
		case errors.Is(err, service.ErrSenhaFraca):
			authError(c, http.StatusBadRequest, "auth.password_change.weak_password")
		case errors.Is(err, service.ErrNovaSenhaIgualAtual):
			authError(c, http.StatusBadRequest, "auth.password_change.same_password")
		default:
			slog.ErrorContext(c.Request.Context(), "falha na troca de senha", "error", err)
			authError(c, http.StatusInternalServerError, "server.internal_error")
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"mensagem": i18n.T(c.GetHeader("Accept-Language"), "auth.password_change.success")}})
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

type ResetSenhaServicer interface {
	Solicitar(ctx context.Context, email string) error
	ValidarToken(ctx context.Context, token string) error
	RedefinirSenha(ctx context.Context, token, senha string) error
}

func NewRecuperacaoSenhaHandler(reset ResetSenhaServicer) *AuthHandler {
	return &AuthHandler{reset: reset}
}

func (h *AuthHandler) SolicitarReset(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err == nil {
		if err := h.reset.Solicitar(c.Request.Context(), req.Email); err != nil {
			if errors.Is(err, service.ErrLimiteSolicitacoesReset) {
				authError(c, http.StatusTooManyRequests, "auth.password_reset.rate_limited")
				return
			}
			slog.ErrorContext(c.Request.Context(), "falha ao solicitar reset de senha", "error", err)
			authError(c, http.StatusInternalServerError, "server.internal_error")
			return
		}
	}
	lang := c.GetHeader("Accept-Language")
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"mensagem": i18n.T(lang, "auth.password_reset.request_accepted")}})
}

func (h *AuthHandler) ValidarTokenReset(c *gin.Context) {
	if err := h.reset.ValidarToken(c.Request.Context(), c.Query("token")); err != nil {
		handleResetError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{}})
}

func (h *AuthHandler) RedefinirSenha(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
		Senha string `json:"senha"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		authError(c, http.StatusBadRequest, "auth.password_reset.invalid_token")
		return
	}
	if err := h.reset.RedefinirSenha(c.Request.Context(), req.Token, req.Senha); err != nil {
		handleResetError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{}})
}

func handleResetError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSenhaFraca):
		authError(c, http.StatusBadRequest, "auth.password_reset.weak_password")
	case errors.Is(err, service.ErrTokenResetExpirado):
		authError(c, http.StatusGone, "auth.password_reset.token_expired")
	case errors.Is(err, service.ErrTokenResetUtilizado):
		authError(c, http.StatusConflict, "auth.password_reset.token_used")
	case errors.Is(err, service.ErrTokenResetInvalido):
		authError(c, http.StatusBadRequest, "auth.password_reset.invalid_token")
	default:
		slog.ErrorContext(c.Request.Context(), "falha no fluxo de reset de senha", "error", err)
		authError(c, http.StatusInternalServerError, "server.internal_error")
	}
}
