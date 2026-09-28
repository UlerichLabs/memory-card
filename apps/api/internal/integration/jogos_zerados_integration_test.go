//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/migration"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

func init() {
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

type postgresIntegrationEnv struct {
	pool    *pgxpool.Pool
	router  *gin.Engine
	tokens  *service.AuthToken
	repo    repository.JogosRepository
	service *service.JogosService
	handler *handler.JogosHandler
	secret  string
}

type criarJogoInput struct {
	Nome                string  `json:"nome"`
	Console             string  `json:"console"`
	Genero              string  `json:"genero,omitempty"`
	Tipo                string  `json:"tipo,omitempty"`
	IniciadoEm          *string `json:"iniciado_em,omitempty"`
	FinalizadoEm        string  `json:"finalizado_em"`
	TempoJogadoHoras    int     `json:"tempo_jogado_horas,omitempty"`
	TempoJogadoMinutos  int     `json:"tempo_jogado_minutos,omitempty"`
	TempoJogadoSegundos int     `json:"tempo_jogado_segundos,omitempty"`
	TempoJogado         *int32  `json:"tempo_jogado,omitempty"`
	Nota                int32   `json:"nota"`
	Dificuldade         string  `json:"dificuldade"`
	Review              string  `json:"review,omitempty"`
	Destaque            bool    `json:"destaque"`
}

type jogoEnvelope struct {
	Data repository.JogoZerado `json:"data"`
}

type erroEnvelope struct {
	Error struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	} `json:"error"`
}

func setupIntegrationEnv(t *testing.T) *postgresIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_integration"),
		tcpostgres.WithUsername("test_user"),
		tcpostgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("falha ao iniciar container postgres de integracao: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao obter connection string: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao interpretar config do pool: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao criar pool pgx: %v", err)
	}

	if _, err := migration.Apply(pool); err != nil {
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao aplicar migrations: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer teardownCancel()
		if err := pgContainer.Terminate(teardownCtx); err != nil {
			t.Logf("aviso: falha ao terminar container: %v", err)
		}
	})

	queries := db.New(pool)
	repo := repository.NewJogosRepository(queries)
	svc := service.NewJogosService(repo)
	h := handler.NewJogosHandler(svc)

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	router := gin.New()
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.POST("/jogos", h.CriarJogo)
	privadas.PUT("/jogos/:id", h.AtualizarJogo)
	privadas.DELETE("/jogos/:id", h.ExcluirJogo)

	return &postgresIntegrationEnv{
		pool:    pool,
		router:  router,
		tokens:  tokens,
		repo:    repo,
		service: svc,
		handler: h,
		secret:  secret,
	}
}

func criarUsuarioTeste(t *testing.T, pool *pgxpool.Pool) int32 {
	t.Helper()
	uid := uuid.NewString()
	email := fmt.Sprintf("user-%s@example.com", uid)
	nome := fmt.Sprintf("User %s", uid[:8])

	var id int32
	err := pool.QueryRow(context.Background(),
		"INSERT INTO usuarios (nome, email, senha_hash) VALUES ($1, $2, $3) RETURNING id",
		nome, email, "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01",
	).Scan(&id)
	if err != nil {
		t.Fatalf("falha ao criar usuario de teste: %v", err)
	}
	return id
}

func gerarAccessTokenTeste(t *testing.T, secret string, usuarioID int32) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
		Tipo:   "access",
		Idioma: "pt-BR",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(usuarioID)),
			Issuer:    "memory-card",
			Audience:  jwt.ClaimStrings{"access"},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("falha ao gerar token de teste: %v", err)
	}
	return token
}

func TestIntegration_JogosZerados(t *testing.T) {
	env := setupIntegrationEnv(t)

	t.Run("1_Criar_Editar_Excluir", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		inputCriar := criarJogoInput{
			Nome:                "Super Mario World",
			Console:             "SNES",
			Genero:              "Platformer",
			Tipo:                "Zeramento Normal",
			FinalizadoEm:        "2026-05-10T14:30:00Z",
			TempoJogadoHoras:    10,
			TempoJogadoMinutos:  30,
			TempoJogadoSegundos: 0,
			Nota:                10,
			Dificuldade:         "A",
			Review:              "Campanha inicial excelente.",
			Destaque:            false,
		}
		bodyCriar, _ := json.Marshal(inputCriar)

		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyCriar))
		reqCriar.Header.Set("Authorization", "Bearer "+token)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)

		if wCriar.Code != http.StatusCreated {
			t.Fatalf("esperava status 201 na criacao, obteve %d: %s", wCriar.Code, wCriar.Body.String())
		}

		var respCriar jogoEnvelope
		if err := json.Unmarshal(wCriar.Body.Bytes(), &respCriar); err != nil {
			t.Fatalf("falha ao decodificar resposta da criacao: %v", err)
		}
		jogoID := respCriar.Data.ID
		if jogoID <= 0 {
			t.Fatalf("ID de jogo invalido retornado: %d", jogoID)
		}

		inputEditar := criarJogoInput{
			Nome:                "Super Mario World Editado",
			Console:             "SNES Classic",
			Genero:              "Platformer 2D",
			Tipo:                "100%",
			FinalizadoEm:        "2026-05-15T18:00:00Z",
			TempoJogadoHoras:    15,
			TempoJogadoMinutos:  45,
			TempoJogadoSegundos: 30,
			Nota:                11,
			Dificuldade:         "AA",
			Review:              "Review atualizada com todos os 96 segredos.",
			Destaque:            false,
		}
		bodyEditar, _ := json.Marshal(inputEditar)

		reqEditar := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jogoID), bytes.NewReader(bodyEditar))
		reqEditar.Header.Set("Authorization", "Bearer "+token)
		reqEditar.Header.Set("Content-Type", "application/json")
		wEditar := httptest.NewRecorder()
		env.router.ServeHTTP(wEditar, reqEditar)

		if wEditar.Code != http.StatusOK {
			t.Fatalf("esperava status 200 na edicao, obteve %d: %s", wEditar.Code, wEditar.Body.String())
		}

		var respEditar jogoEnvelope
		if err := json.Unmarshal(wEditar.Body.Bytes(), &respEditar); err != nil {
			t.Fatalf("falha ao decodificar resposta da edicao: %v", err)
		}
		if respEditar.Data.Nome != "Super Mario World Editado" || respEditar.Data.Nota != 11 {
			t.Fatalf("valores da resposta da edicao incorretos: %+v", respEditar.Data)
		}

		var dbNome, dbReview string
		var dbNota int32
		var dbDeletedAt *time.Time
		err := env.pool.QueryRow(context.Background(),
			"SELECT nome, review, nota, deleted_at FROM jogos_zerados WHERE id = $1",
			jogoID,
		).Scan(&dbNome, &dbReview, &dbNota, &dbDeletedAt)
		if err != nil {
			t.Fatalf("falha ao consultar banco apos edicao: %v", err)
		}
		if dbNome != "Super Mario World Editado" || dbReview != "Review atualizada com todos os 96 segredos." || dbNota != 11 {
			t.Fatalf("banco nao reflete edicao: nome=%s, review=%s, nota=%d", dbNome, dbReview, dbNota)
		}
		if dbDeletedAt != nil {
			t.Fatalf("deleted_at nao deveria estar preenchido antes do delete: %v", dbDeletedAt)
		}

		reqExcluir := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqExcluir.Header.Set("Authorization", "Bearer "+token)
		wExcluir := httptest.NewRecorder()
		env.router.ServeHTTP(wExcluir, reqExcluir)

		if wExcluir.Code != http.StatusNoContent {
			t.Fatalf("esperava status 204 na exclusao, obteve %d: %s", wExcluir.Code, wExcluir.Body.String())
		}

		var dbDeletedAtAfter *time.Time
		err = env.pool.QueryRow(context.Background(),
			"SELECT deleted_at FROM jogos_zerados WHERE id = $1",
			jogoID,
		).Scan(&dbDeletedAtAfter)
		if err != nil {
			t.Fatalf("registro nao encontrado no banco apos soft delete: %v", err)
		}
		if dbDeletedAtAfter == nil {
			t.Fatal("deleted_at deveria estar preenchido apos soft delete")
		}
	})

	t.Run("2_Registro_Excluido_Retorna_404_Em_PUT_E_DELETE", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input := criarJogoInput{
			Nome:         "Castlevania",
			Console:      "NES",
			FinalizadoEm: "2026-01-10T12:00:00Z",
			Nota:         9,
			Dificuldade:  "B",
		}
		body, _ := json.Marshal(input)

		reqCriar := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body))
		reqCriar.Header.Set("Authorization", "Bearer "+token)
		reqCriar.Header.Set("Content-Type", "application/json")
		wCriar := httptest.NewRecorder()
		env.router.ServeHTTP(wCriar, reqCriar)

		if wCriar.Code != http.StatusCreated {
			t.Fatalf("status criacao=%d", wCriar.Code)
		}
		var respCriar jogoEnvelope
		_ = json.Unmarshal(wCriar.Body.Bytes(), &respCriar)
		jogoID := respCriar.Data.ID

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)
		if wDel.Code != http.StatusNoContent {
			t.Fatalf("status delete=%d", wDel.Code)
		}

		reqPut := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jogoID), bytes.NewReader(body))
		reqPut.Header.Set("Authorization", "Bearer "+token)
		reqPut.Header.Set("Content-Type", "application/json")
		wPut := httptest.NewRecorder()
		env.router.ServeHTTP(wPut, reqPut)

		if wPut.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em PUT de registro excluido, obteve %d", wPut.Code)
		}
		var errPut erroEnvelope
		_ = json.Unmarshal(wPut.Body.Bytes(), &errPut)
		if errPut.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errPut.Error.Codigo)
		}

		reqDel2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDel2.Header.Set("Authorization", "Bearer "+token)
		wDel2 := httptest.NewRecorder()
		env.router.ServeHTTP(wDel2, reqDel2)

		if wDel2.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 em DELETE de registro excluido, obteve %d", wDel2.Code)
		}
		var errDel erroEnvelope
		_ = json.Unmarshal(wDel2.Body.Bytes(), &errDel)
		if errDel.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errDel.Error.Codigo)
		}
	})

	t.Run("3_Destaque_Conflito_Mesmo_Ano_Mesmo_Usuario", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input1 := criarJogoInput{
			Nome:         "Zelda I",
			Console:      "NES",
			FinalizadoEm: "2026-03-01T12:00:00Z",
			Nota:         10,
			Dificuldade:  "A",
			Destaque:     true,
		}
		body1, _ := json.Marshal(input1)

		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		env.router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no primeiro destaque, obteve %d: %s", w1.Code, w1.Body.String())
		}

		input2 := criarJogoInput{
			Nome:         "Zelda II",
			Console:      "NES",
			FinalizadoEm: "2026-11-20T12:00:00Z",
			Nota:         9,
			Dificuldade:  "AA",
			Destaque:     true,
		}
		body2, _ := json.Marshal(input2)

		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		env.router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusConflict {
			t.Fatalf("esperava 409 no segundo destaque do mesmo ano, obteve %d: %s", w2.Code, w2.Body.String())
		}

		var resp erroEnvelope
		if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
			t.Fatalf("falha ao decodificar erro 409: %v", err)
		}
		if resp.Error.Codigo != "jogos.destaque_ano_conflito" {
			t.Fatalf("esperava codigo jogos.destaque_ano_conflito, obteve %s", resp.Error.Codigo)
		}
	})

	t.Run("4_Destaque_Anos_Diferentes_Ambos_201", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input1 := criarJogoInput{
			Nome:         "Chrono Trigger",
			Console:      "SNES",
			FinalizadoEm: "2024-06-10T12:00:00Z",
			Nota:         10,
			Dificuldade:  "A",
			Destaque:     true,
		}
		body1, _ := json.Marshal(input1)

		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		env.router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no destaque de 2024, obteve %d", w1.Code)
		}

		input2 := criarJogoInput{
			Nome:         "Final Fantasy VI",
			Console:      "SNES",
			FinalizadoEm: "2025-06-10T12:00:00Z",
			Nota:         10,
			Dificuldade:  "A",
			Destaque:     true,
		}
		body2, _ := json.Marshal(input2)

		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		env.router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no destaque de 2025, obteve %d", w2.Code)
		}
	})

	t.Run("5_Destaque_Mesmo_Ano_Usuarios_Diferentes_Ambos_201", func(t *testing.T) {
		userA := criarUsuarioTeste(t, env.pool)
		userB := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, userA)
		tokenB := gerarAccessTokenTeste(t, env.secret, userB)

		inputA := criarJogoInput{
			Nome:         "Metroid Prime",
			Console:      "GameCube",
			FinalizadoEm: "2026-08-01T12:00:00Z",
			Nota:         10,
			Dificuldade:  "AA",
			Destaque:     true,
		}
		bodyA, _ := json.Marshal(inputA)

		reqA := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyA))
		reqA.Header.Set("Authorization", "Bearer "+tokenA)
		reqA.Header.Set("Content-Type", "application/json")
		wA := httptest.NewRecorder()
		env.router.ServeHTTP(wA, reqA)

		if wA.Code != http.StatusCreated {
			t.Fatalf("esperava 201 para usuario A, obteve %d", wA.Code)
		}

		inputB := criarJogoInput{
			Nome:         "Resident Evil 4",
			Console:      "GameCube",
			FinalizadoEm: "2026-09-01T12:00:00Z",
			Nota:         11,
			Dificuldade:  "AAA",
			Destaque:     true,
		}
		bodyB, _ := json.Marshal(inputB)

		reqB := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyB))
		reqB.Header.Set("Authorization", "Bearer "+tokenB)
		reqB.Header.Set("Content-Type", "application/json")
		wB := httptest.NewRecorder()
		env.router.ServeHTTP(wB, reqB)

		if wB.Code != http.StatusCreated {
			t.Fatalf("esperava 201 para usuario B no mesmo ano, obteve %d", wB.Code)
		}
	})

	t.Run("6_Excluir_Destaque_Libera_Ano_Para_Novo_Destaque", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)
		token := gerarAccessTokenTeste(t, env.secret, usuarioID)

		input1 := criarJogoInput{
			Nome:         "Pokemon Red",
			Console:      "Game Boy",
			FinalizadoEm: "2023-02-10T12:00:00Z",
			Nota:         9,
			Dificuldade:  "B",
			Destaque:     true,
		}
		body1, _ := json.Marshal(input1)

		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body1))
		req1.Header.Set("Authorization", "Bearer "+token)
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		env.router.ServeHTTP(w1, req1)

		if w1.Code != http.StatusCreated {
			t.Fatalf("esperava 201 no primeiro destaque, obteve %d", w1.Code)
		}
		var resp1 jogoEnvelope
		_ = json.Unmarshal(w1.Body.Bytes(), &resp1)
		jogoID := resp1.Data.ID

		reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDel.Header.Set("Authorization", "Bearer "+token)
		wDel := httptest.NewRecorder()
		env.router.ServeHTTP(wDel, reqDel)

		if wDel.Code != http.StatusNoContent {
			t.Fatalf("esperava 204 na exclusao do destaque, obteve %d", wDel.Code)
		}

		input2 := criarJogoInput{
			Nome:         "Pokemon Blue",
			Console:      "Game Boy",
			FinalizadoEm: "2023-11-20T12:00:00Z",
			Nota:         10,
			Dificuldade:  "B",
			Destaque:     true,
		}
		body2, _ := json.Marshal(input2)

		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(body2))
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		env.router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusCreated {
			t.Fatalf("esperava 201 ao cadastrar novo destaque no mesmo ano apos exclusao, obteve %d: %s", w2.Code, w2.Body.String())
		}
	})

	t.Run("7_Isolamento_Usuario_B_Nao_Edita_Nem_Exclui_Registro_De_Usuario_A", func(t *testing.T) {
		userA := criarUsuarioTeste(t, env.pool)
		userB := criarUsuarioTeste(t, env.pool)
		tokenA := gerarAccessTokenTeste(t, env.secret, userA)
		tokenB := gerarAccessTokenTeste(t, env.secret, userB)

		inputA := criarJogoInput{
			Nome:         "Silent Hill",
			Console:      "PlayStation",
			FinalizadoEm: "2026-07-04T12:00:00Z",
			Nota:         10,
			Dificuldade:  "AA",
		}
		bodyA, _ := json.Marshal(inputA)

		reqA := httptest.NewRequest(http.MethodPost, "/api/v1/jogos", bytes.NewReader(bodyA))
		reqA.Header.Set("Authorization", "Bearer "+tokenA)
		reqA.Header.Set("Content-Type", "application/json")
		wA := httptest.NewRecorder()
		env.router.ServeHTTP(wA, reqA)

		if wA.Code != http.StatusCreated {
			t.Fatalf("esperava 201 na criacao de user A, obteve %d", wA.Code)
		}
		var respA jogoEnvelope
		_ = json.Unmarshal(wA.Body.Bytes(), &respA)
		jogoID := respA.Data.ID

		inputHack := criarJogoInput{
			Nome:         "Silent Hill Modificado por B",
			Console:      "PlayStation",
			FinalizadoEm: "2026-07-04T12:00:00Z",
			Nota:         1,
			Dificuldade:  "C",
		}
		bodyHack, _ := json.Marshal(inputHack)

		reqPutB := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/jogos/%d", jogoID), bytes.NewReader(bodyHack))
		reqPutB.Header.Set("Authorization", "Bearer "+tokenB)
		reqPutB.Header.Set("Content-Type", "application/json")
		wPutB := httptest.NewRecorder()
		env.router.ServeHTTP(wPutB, reqPutB)

		if wPutB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 quando usuario B tenta editar registro de A, obteve %d", wPutB.Code)
		}
		var errPut erroEnvelope
		_ = json.Unmarshal(wPutB.Body.Bytes(), &errPut)
		if errPut.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errPut.Error.Codigo)
		}

		reqDelB := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/jogos/%d", jogoID), nil)
		reqDelB.Header.Set("Authorization", "Bearer "+tokenB)
		wDelB := httptest.NewRecorder()
		env.router.ServeHTTP(wDelB, reqDelB)

		if wDelB.Code != http.StatusNotFound {
			t.Fatalf("esperava 404 quando usuario B tenta excluir registro de A, obteve %d", wDelB.Code)
		}
		var errDel erroEnvelope
		_ = json.Unmarshal(wDelB.Body.Bytes(), &errDel)
		if errDel.Error.Codigo != "jogos.not_found" {
			t.Fatalf("esperava codigo jogos.not_found, obteve %s", errDel.Error.Codigo)
		}

		var dbNome string
		var dbDeletedAt *time.Time
		err := env.pool.QueryRow(context.Background(),
			"SELECT nome, deleted_at FROM jogos_zerados WHERE id = $1",
			jogoID,
		).Scan(&dbNome, &dbDeletedAt)
		if err != nil {
			t.Fatalf("falha ao consultar registro original de A: %v", err)
		}
		if dbNome != "Silent Hill" || dbDeletedAt != nil {
			t.Fatalf("registro de A foi alterado: nome=%s, deleted_at=%v", dbNome, dbDeletedAt)
		}
	})

	t.Run("8_Nota_12_Rejeitada_Pelo_CHECK_Do_Banco_Via_Repository", func(t *testing.T) {
		usuarioID := criarUsuarioTeste(t, env.pool)

		_, err := env.repo.Criar(context.Background(), repository.CriarJogoZeradoParams{
			UsuarioID:    usuarioID,
			Nome:         "Jogo Nota Invalida DB",
			Console:      "Mega Drive",
			FinalizadoEm: time.Now(),
			TempoJogado:  3600,
			Nota:         12,
			Dificuldade:  "A",
		})
		if err == nil {
			t.Fatal("esperava erro de violacao de CHECK constraint ao enviar nota 12 ao repository")
		}

		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			t.Fatalf("esperava erro do tipo *pgconn.PgError, obteve: %v", err)
		}
		if pgErr.Code != "23514" {
			t.Fatalf("esperava codigo postgres 23514 (check_violation), obteve: %s", pgErr.Code)
		}
	})
}
