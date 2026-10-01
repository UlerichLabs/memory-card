package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestAuth_Cenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	sign := func(tipo, key string, expiry time.Time) string {
		t.Helper()
		raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{Idioma: "en", Tipo: tipo, RegisteredClaims: jwt.RegisteredClaims{
			Subject: "42", Issuer: "memory-card", Audience: jwt.ClaimStrings{tipo}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), ExpiresAt: jwt.NewNumericDate(expiry),
		}}).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	valid := sign("access", secret, time.Now().Add(time.Hour))
	for _, tc := range []struct {
		name, header string
		status       int
	}{
		{"ausente", "", 401},
		{"malformado", "Bearer invalido", 401},
		{"expirado", "Bearer " + sign("access", secret, time.Now().Add(-time.Minute)), 401},
		{"assinatura incorreta", "Bearer " + sign("access", uuid.NewString(), time.Now().Add(time.Hour)), 401},
		{"refresh como access", "Bearer " + sign("refresh", secret, time.Now().Add(time.Hour)), 401},
		{"esquema incorreto", "Basic " + valid, 401},
		{"valido", "Bearer " + valid, 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			called := false
			GrupoPrivado(router, tokens).GET("/teste", func(ctx *gin.Context) {
				called = true
				usuario, ok := UsuarioDoContexto(ctx.Request.Context())
				ginUsuario, exists := ctx.Get("usuario")
				if !ok || !exists || ginUsuario != usuario || usuario.ID != "42" || usuario.Idioma != "en" {
					t.Fatal("contexto incorreto")
				}
				ctx.Status(http.StatusNoContent)
			})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/teste", nil)
			req.Header.Set("Authorization", tc.header)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code != tc.status || called != (tc.status == 204) {
				t.Fatalf("status %d; body %s", recorder.Code, recorder.Body.String())
			}
			if tc.status == 401 && (!strings.Contains(recorder.Body.String(), "auth.session.unauthorized") || !strings.Contains(recorder.Body.String(), "Não autorizado. Faça login novamente.")) {
				t.Fatal(recorder.Body.String())
			}
		})
	}
}
