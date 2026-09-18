//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/UlerichLabs/memory-card/apps/api/internal/testutil"
)

type apiErrorResponse struct {
	Error struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	} `json:"error"`
}

type authDataResponse[T any] struct {
	Data T `json:"data"`
}

func TestE2E_LoginSessao(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pg := testutil.SetupPostgres(t)
	repo := repository.NewUsuarioRepository(db.New(pg.Pool))
	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	login, err := service.NewLoginService(repo, tokens)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.POST("/api/v1/auth/register", handler.NewAuthHandler(service.NewCadastroService(repo)).Register)
	loginHandler := handler.NewLoginHandler(login)
	router.POST("/api/v1/auth/login", loginHandler.Login)
	router.POST("/api/v1/auth/refresh", loginHandler.Refresh)
	meHandler := handler.NewMeHandler(service.NewPerfilService(repo))
	middleware.GrupoPrivado(router, tokens).GET("/me", meHandler.Me)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	payload := map[string]string{"nome": "Usuario Login", "email": "login@example.com", "senha": "SenhaForte@123"}
	cadastro := authRequest[authDataResponse[repository.Usuario]](t, server, http.MethodPost, "/api/v1/auth/register", payload, "", http.StatusCreated).Data
	sessao := authRequest[authDataResponse[service.LoginResult]](t, server, http.MethodPost, "/api/v1/auth/login", payload, "", http.StatusOK).Data
	if sessao.Usuario != cadastro || cadastro.ID <= 0 || sessao.AccessToken == "" || sessao.RefreshToken == "" || sessao.AccessToken == sessao.RefreshToken {
		t.Fatal("login nao retornou o usuario cadastrado e tokens distintos")
	}
	for _, tc := range []struct {
		tipo, raw string
		ttl       time.Duration
	}{
		{"access", sessao.AccessToken, 15 * time.Minute},
		{"refresh", sessao.RefreshToken, 168 * time.Hour},
	} {
		claims := validarTokenE2E(t, tc.raw, secret, tc.tipo)
		if claims.Subject != strconv.FormatInt(int64(cadastro.ID), 10) || claims.Idioma != cadastro.Idioma || claims.ExpiresAt.Sub(claims.IssuedAt.Time) != tc.ttl {
			t.Fatalf("claims incorretas no token %s", tc.tipo)
		}
	}

	t.Run("rota privada e refresh preservam identidade", func(t *testing.T) {
		perfil := authRequest[authDataResponse[repository.Usuario]](t, server, http.MethodGet, "/api/v1/me", nil, sessao.AccessToken, http.StatusOK).Data
		if perfil != cadastro {
			t.Fatal("perfil nao corresponde ao usuario autenticado")
		}
		semToken := authRequest[apiErrorResponse](t, server, http.MethodGet, "/api/v1/me", nil, "", http.StatusUnauthorized)
		assertAuthError(t, semToken, "auth.session.unauthorized", "Não autorizado. Faça login novamente.")
		refresh := authRequest[authDataResponse[service.RefreshResult]](t, server, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": sessao.RefreshToken}, "", http.StatusOK).Data
		if refresh.AccessToken == sessao.AccessToken || refresh.AccessToken == "" {
			t.Fatal("refresh nao emitiu novo access token")
		}
		validarTokenE2E(t, refresh.AccessToken, secret, "access")
		perfil = authRequest[authDataResponse[repository.Usuario]](t, server, http.MethodGet, "/api/v1/me", nil, refresh.AccessToken, http.StatusOK).Data
		if perfil != cadastro {
			t.Fatal("novo access token nao preservou identidade")
		}
	})

	t.Run("credenciais invalidas nao distinguem email e senha", func(t *testing.T) {
		emailAusente := authRequest[apiErrorResponse](t, server, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "ausente@example.com", "senha": payload["senha"]}, "", http.StatusUnauthorized)
		senhaErrada := authRequest[apiErrorResponse](t, server, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": payload["email"], "senha": "Incorreta@123"}, "", http.StatusUnauthorized)
		assertAuthError(t, emailAusente, "auth.login.invalid_credentials", "Email ou senha inválidos.")
		if emailAusente != senhaErrada {
			t.Fatal("respostas diferentes para email ausente e senha incorreta")
		}
	})

	for _, tc := range []struct{ name, token string }{
		{"expirado", expirarTokenE2E(t, sessao.AccessToken, secret, "access")},
		{"malformado", "invalido"},
		{"refresh como access", sessao.RefreshToken},
	} {
		t.Run("rota privada rejeita "+tc.name, func(t *testing.T) {
			result := authRequest[apiErrorResponse](t, server, http.MethodGet, "/api/v1/me", nil, tc.token, http.StatusUnauthorized)
			assertAuthError(t, result, "auth.session.unauthorized", "Não autorizado. Faça login novamente.")
		})
	}
	for _, tc := range []struct{ name, token string }{
		{"expirado", expirarTokenE2E(t, sessao.RefreshToken, secret, "refresh")},
		{"malformado", "invalido"},
		{"ausente", ""},
		{"access como refresh", sessao.AccessToken},
	} {
		t.Run("refresh rejeita "+tc.name, func(t *testing.T) {
			result := authRequest[apiErrorResponse](t, server, http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": tc.token}, "", http.StatusUnauthorized)
			assertAuthError(t, result, "auth.session.expired", "Sessão expirada. Faça login novamente.")
		})
	}
}

func authRequest[T any](t *testing.T, server *httptest.Server, method, path string, payload any, token string, status int) T {
	t.Helper()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, server.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != status {
		t.Fatalf("%s %s: status %d, esperado %d", method, path, resp.StatusCode, status)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(`"senha"`)) || bytes.Contains(raw, []byte(`"senha_hash"`)) {
		t.Fatal("resposta expoe credenciais")
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertAuthError(t *testing.T, result apiErrorResponse, codigo, mensagem string) {
	t.Helper()
	if result.Error.Codigo != codigo || result.Error.Mensagem != mensagem {
		t.Fatalf("erro = %+v, esperado %s: %s", result.Error, codigo, mensagem)
	}
}

func validarTokenE2E(t *testing.T, raw, secret, tipo string) *service.AuthClaims {
	t.Helper()
	claims := &service.AuthClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return []byte(secret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithIssuer("memory-card"), jwt.WithAudience(tipo))
	if err != nil {
		t.Fatal(err)
	}
	if !token.Valid || claims.Tipo != tipo || claims.IssuedAt == nil || claims.ID == "" {
		t.Fatal("token sem claims obrigatorias")
	}
	return claims
}

func expirarTokenE2E(t *testing.T, raw, secret, tipo string) string {
	t.Helper()
	claims := validarTokenE2E(t, raw, secret, tipo)
	claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	expirado, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return expirado
}

func TestE2E_Cadastro(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pg := testutil.SetupPostgres(t)

	queries := db.New(pg.Pool)
	usuarioRepo := repository.NewUsuarioRepository(queries)
	cadastroService := service.NewCadastroService(usuarioRepo)
	authHandler := handler.NewAuthHandler(cadastroService)

	router := gin.New()
	router.POST("/api/v1/auth/register", authHandler.Register)

	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
	})

	t.Run("fluxo completo com persistencia real no banco", func(t *testing.T) {
		payload := map[string]string{
			"nome":  "Lucas",
			"email": "lucas@example.com",
			"senha": "SenhaForte@123",
		}
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("falha ao serializar json: %v", err)
		}

		resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao enviar POST /api/v1/auth/register: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusCreated)
		}

		var successResp struct {
			Data repository.Usuario `json:"data"`
		}
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("falha ao ler corpo da resposta: %v", err)
		}
		if err := json.Unmarshal(respBytes, &successResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de sucesso: %v", err)
		}

		if successResp.Data.ID <= 0 || successResp.Data.Email != "lucas@example.com" {
			t.Fatalf("dados do usuario incorretos na resposta: %+v", successResp.Data)
		}

		var dbID int32
		var dbSenhaHash string
		var dbIdioma string
		query := "SELECT id, senha_hash, idioma FROM usuarios WHERE email = $1"
		err = pg.Pool.QueryRow(context.Background(), query, "lucas@example.com").Scan(&dbID, &dbSenhaHash, &dbIdioma)
		if err != nil {
			t.Fatalf("falha ao consultar usuario persistido no banco: %v", err)
		}

		if dbID != successResp.Data.ID {
			t.Errorf("dbID = %d, responseID = %d", dbID, successResp.Data.ID)
		}
		if dbIdioma != "pt-BR" {
			t.Errorf("dbIdioma = %q, esperado 'pt-BR'", dbIdioma)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(dbSenhaHash), []byte("SenhaForte@123")); err != nil {
			t.Fatalf("hash da senha no banco invalido: %v", err)
		}
	})

	t.Run("email duplicado retorna 409", func(t *testing.T) {
		payload := map[string]string{
			"nome":  "Lucas Duplicado",
			"email": "lucas@example.com",
			"senha": "SenhaForte@123",
		}
		bodyBytes, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao enviar POST com email duplicado: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusConflict)
		}

		var errResp apiErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de erro: %v", err)
		}
		if errResp.Error.Codigo != "auth.register.email_taken" {
			t.Errorf("codigo = %q, esperado 'auth.register.email_taken'", errResp.Error.Codigo)
		}
		if errResp.Error.Mensagem != "Este email já está em uso." {
			t.Errorf("mensagem = %q, esperado 'Este email já está em uso.'", errResp.Error.Mensagem)
		}
	})

	t.Run("senha fraca retorna 400 sem persistir no banco", func(t *testing.T) {
		emailFraco := "senhafraca@example.com"
		payload := map[string]string{
			"nome":  "Usuario Fraco",
			"email": emailFraco,
			"senha": "123",
		}
		bodyBytes, _ := json.Marshal(payload)

		resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao enviar POST com senha fraca: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusBadRequest)
		}

		var errResp apiErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de erro: %v", err)
		}
		if errResp.Error.Codigo != "auth.register.weak_password" {
			t.Errorf("codigo = %q, esperado 'auth.register.weak_password'", errResp.Error.Codigo)
		}
		if errResp.Error.Mensagem != "A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial." {
			t.Errorf("mensagem inesperada: %q", errResp.Error.Mensagem)
		}

		var count int
		err = pg.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM usuarios WHERE email = $1", emailFraco).Scan(&count)
		if err != nil {
			t.Fatalf("falha ao verificar persistencia no banco: %v", err)
		}
		if count != 0 {
			t.Fatalf("usuario com senha fraca nao deveria ter sido persistido no banco")
		}
	})

	t.Run("accept language en retorna mensagem em ingles", func(t *testing.T) {
		payload := map[string]string{
			"nome":  "English User",
			"email": "en-weak@example.com",
			"senha": "weak",
		}
		bodyBytes, _ := json.Marshal(payload)

		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/register", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("falha ao criar requisicao: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Language", "en")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("falha ao enviar requisicao com Accept-Language: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status code = %d, esperado %d", resp.StatusCode, http.StatusBadRequest)
		}

		var errResp apiErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errResp); err != nil {
			t.Fatalf("falha ao decodificar JSON de erro: %v", err)
		}
		if errResp.Error.Codigo != "auth.register.weak_password" {
			t.Errorf("codigo = %q, esperado 'auth.register.weak_password'", errResp.Error.Codigo)
		}
		if errResp.Error.Mensagem != "Password must be at least 8 characters long, including an uppercase letter, a number, and a special character." {
			t.Errorf("mensagem em ingles = %q, inesperada", errResp.Error.Mensagem)
		}

		dupPayload := map[string]string{
			"nome":  "English User 2",
			"email": "lucas@example.com",
			"senha": "SenhaForte@123",
		}
		dupBytes, _ := json.Marshal(dupPayload)
		dupReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/register", bytes.NewReader(dupBytes))
		dupReq.Header.Set("Content-Type", "application/json")
		dupReq.Header.Set("Accept-Language", "en-US")

		dupResp, err := http.DefaultClient.Do(dupReq)
		if err != nil {
			t.Fatalf("falha ao enviar requisicao de duplicidade com Accept-Language: %v", err)
		}
		defer dupResp.Body.Close()

		if dupResp.StatusCode != http.StatusConflict {
			t.Fatalf("status code = %d, esperado %d", dupResp.StatusCode, http.StatusConflict)
		}

		var dupErrResp apiErrorResponse
		dupBody, _ := io.ReadAll(dupResp.Body)
		_ = json.Unmarshal(dupBody, &dupErrResp)
		if dupErrResp.Error.Codigo != "auth.register.email_taken" {
			t.Errorf("codigo = %q, esperado 'auth.register.email_taken'", dupErrResp.Error.Codigo)
		}
		if dupErrResp.Error.Mensagem != "This email is already in use." {
			t.Errorf("mensagem em ingles duplicado = %q, inesperada", dupErrResp.Error.Mensagem)
		}
	})
}
