package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/UlerichLabs/memory-card/apps/api/internal/i18n"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type usuarioContextKey struct{}

type UsuarioAutenticado struct {
	ID     string
	Idioma string
}

func UsuarioDoContexto(ctx context.Context) (UsuarioAutenticado, bool) {
	usuario, ok := ctx.Value(usuarioContextKey{}).(UsuarioAutenticado)
	return usuario, ok
}

func GrupoPrivado(router *gin.Engine, tokens *service.AuthToken) *gin.RouterGroup {
	return router.Group("/api/v1", Auth(tokens))
}

func Auth(tokens *service.AuthToken) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(c)
			return
		}
		claims, err := tokens.ValidarAccess(parts[1])
		if err != nil {
			unauthorized(c)
			return
		}
		usuario := UsuarioAutenticado{ID: claims.Subject, Idioma: claims.Idioma}
		c.Set("usuario", usuario)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), usuarioContextKey{}, usuario))
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	const codigo = "auth.session.unauthorized"
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"codigo": codigo, "mensagem": i18n.T(c.GetHeader("Accept-Language"), codigo)}})
}
