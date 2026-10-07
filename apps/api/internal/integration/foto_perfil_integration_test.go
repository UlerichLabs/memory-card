//go:build integration

// Package integration contem os testes de integracao com banco real.
package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/migration"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
	"github.com/UlerichLabs/memory-card/apps/api/internal/storage"
	migfs "github.com/UlerichLabs/memory-card/apps/api/migrations"
)

type fotoPerfilIntegrationEnv struct {
	pool      *pgxpool.Pool
	sqlDB     *sql.DB
	router    *gin.Engine
	tokens    *service.AuthToken
	secret    string
	avatarDir string
}

func setupFotoPerfilIntegrationEnv(t *testing.T) *fotoPerfilIntegrationEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_foto_test"),
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

	sqlDB := stdlib.OpenDBFromPool(pool)

	if _, err := migration.Apply(pool); err != nil {
		sqlDB.Close()
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao aplicar migrations: %v", err)
	}

	avatarDir, err := os.MkdirTemp("", "avatar-integration-*")
	if err != nil {
		sqlDB.Close()
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao criar avatarDir temporario: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(avatarDir)
		sqlDB.Close()
		pool.Close()
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer teardownCancel()
		if err := pgContainer.Terminate(teardownCtx); err != nil {
			t.Logf("aviso: falha ao terminar container: %v", err)
		}
	})

	queries := db.New(pool)
	usuarioRepo := repository.NewUsuarioRepository(queries, pool)

	avatarStorage, err := storage.NewAvatarFS(avatarDir)
	if err != nil {
		t.Fatalf("falha ao criar AvatarFS: %v", err)
	}

	secret := uuid.NewString()
	tokens, err := service.NewAuthToken(secret, time.Hour, 24*time.Hour)
	if err != nil {
		t.Fatalf("falha ao instanciar auth token: %v", err)
	}

	perfilSvc := service.NewPerfilService(usuarioRepo)
	perfilH := handler.NewPerfilHandler(perfilSvc)

	fotoSvc := service.NewFotoPerfilService(usuarioRepo, avatarStorage)
	fotoH := handler.NewFotoPerfilHandler(fotoSvc)

	router := gin.New()
	router.GET("/api/v1/avatares/:arquivo", fotoH.ServirArquivo)

	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/me/perfil", perfilH.ObterPerfil)
	privadas.PUT("/me/foto", fotoH.Upload)
	privadas.PUT("/me/foto/capa", fotoH.DefinirCapa)
	privadas.DELETE("/me/foto", fotoH.RemoverFoto)

	return &fotoPerfilIntegrationEnv{
		pool:      pool,
		sqlDB:     sqlDB,
		router:    router,
		tokens:    tokens,
		secret:    secret,
		avatarDir: avatarDir,
	}
}

func criarUsuarioFoto(t *testing.T, pool *pgxpool.Pool) int32 {
	t.Helper()
	email := fmt.Sprintf("user_%s@example.com", uuid.NewString()[:8])
	var id int32
	err := pool.QueryRow(context.Background(), `
		INSERT INTO usuarios (nome, email, senha_hash, username)
		VALUES ($1, $2, 'hash_inutil', $3)
		RETURNING id
	`, "Test User", email, "u_"+uuid.NewString()[:8]).Scan(&id)
	if err != nil {
		t.Fatalf("criar usuario: %v", err)
	}
	return id
}

func gerarTokenFoto(t *testing.T, secret string, usuarioID int32) string {
	t.Helper()
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, service.AuthClaims{
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
		t.Fatalf("gerar token: %v", err)
	}
	return raw
}

func gerarJPEGIntegration(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

func criarJogoComCapa(t *testing.T, pool *pgxpool.Pool, usuarioID int32, capaURL string) int32 {
	t.Helper()
	var id int32
	err := pool.QueryRow(context.Background(), `
		INSERT INTO jogos_zerados (usuario_id, nome, console, finalizado_em, tempo_jogado, nota, dificuldade, igdb_capa_url)
		VALUES ($1, $2, 'PS5', now(), 3600, 10, 'A', $3)
		RETURNING id
	`, usuarioID, "Jogo com Capa", capaURL).Scan(&id)
	if err != nil {
		t.Fatalf("criar jogo com capa: %v", err)
	}
	return id
}

func TestIntegration_FotoPerfil(t *testing.T) {
	env := setupFotoPerfilIntegrationEnv(t)

	t.Run("Migration 0017 sobe e desce e CHECK impede ambos preenchidos", func(t *testing.T) {
		driver, err := pgxmigrate.WithInstance(env.sqlDB, &pgxmigrate.Config{})
		if err != nil {
			t.Fatalf("falha ao criar driver do migrate: %v", err)
		}
		src, err := iofs.New(migfs.FS, ".")
		if err != nil {
			t.Fatalf("falha ao criar source iofs do migrate: %v", err)
		}
		mig, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
		if err != nil {
			t.Fatalf("falha ao instanciar migrate: %v", err)
		}
		defer func() { _, _ = mig.Close() }()

		if err := mig.Migrate(16); err != nil {
			t.Fatalf("migracao down 0017 falhou: %v", err)
		}

		var colunaExiste bool
		err = env.pool.QueryRow(context.Background(), `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'usuarios' AND column_name = 'avatar_jogo_id'
			)
		`).Scan(&colunaExiste)
		if err != nil {
			t.Fatalf("verificar coluna apos down: %v", err)
		}
		if colunaExiste {
			t.Fatalf("coluna avatar_jogo_id ainda existe apos down da 0017")
		}

		if err := mig.Up(); err != nil {
			t.Fatalf("migracao up 0017 falhou: %v", err)
		}

		err = env.pool.QueryRow(context.Background(), `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'usuarios' AND column_name = 'avatar_jogo_id'
			)
		`).Scan(&colunaExiste)
		if err != nil {
			t.Fatalf("verificar coluna apos up: %v", err)
		}
		if !colunaExiste {
			t.Fatalf("coluna avatar_jogo_id deveria existir apos up da 0017")
		}

		usuarioID := criarUsuarioFoto(t, env.pool)
		jogoID := criarJogoComCapa(t, env.pool, usuarioID, "https://example.com/cover.jpg")

		_, err = env.pool.Exec(context.Background(), `
			UPDATE usuarios
			SET avatar_url = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.jpg',
			    avatar_jogo_id = $1
			WHERE id = $2
		`, jogoID, usuarioID)
		if err == nil {
			t.Fatalf("esperava erro de CHECK constraint ao tentar preencher avatar_url e avatar_jogo_id simultaneamente")
		}
		if !strings.Contains(err.Error(), "check_usuarios_avatar_mutuamente_exclusivo") {
			t.Fatalf("esperava violacao de check_usuarios_avatar_mutuamente_exclusivo, obteve: %v", err)
		}
	})

	t.Run("Fluxo completo upload GET capa delete", func(t *testing.T) {
		usuarioID := criarUsuarioFoto(t, env.pool)
		token := gerarTokenFoto(t, env.secret, usuarioID)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("arquivo", "minha_foto.jpg")
		if err != nil {
			t.Fatalf("criar form file: %v", err)
		}
		jpegData := gerarJPEGIntegration(120, 120)
		if _, err := part.Write(jpegData); err != nil {
			t.Fatalf("escrever dados: %v", err)
		}
		_ = writer.Close()

		reqUpload := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto", body)
		reqUpload.Header.Set("Authorization", "Bearer "+token)
		reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
		recUpload := httptest.NewRecorder()
		env.router.ServeHTTP(recUpload, reqUpload)

		if recUpload.Code != http.StatusOK {
			t.Fatalf("upload falhou: %d %s", recUpload.Code, recUpload.Body.String())
		}

		var respUpload struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		if err := json.Unmarshal(recUpload.Body.Bytes(), &respUpload); err != nil {
			t.Fatalf("unmarshal upload: %v", err)
		}
		if respUpload.Data.AvatarTipo == nil || *respUpload.Data.AvatarTipo != "upload" {
			t.Fatalf("avatar_tipo esperado upload, obteve %+v", respUpload.Data.AvatarTipo)
		}
		if respUpload.Data.AvatarURL == nil || !strings.HasPrefix(*respUpload.Data.AvatarURL, "/api/v1/avatares/") {
			t.Fatalf("avatar_url esperado caminho relativo /api/v1/avatares/..., obteve %+v", respUpload.Data.AvatarURL)
		}

		avatarPath := *respUpload.Data.AvatarURL
		nomeArquivo := strings.TrimPrefix(avatarPath, "/api/v1/avatares/")

		reqPerfil := httptest.NewRequest(http.MethodGet, "/api/v1/me/perfil", nil)
		reqPerfil.Header.Set("Authorization", "Bearer "+token)
		recPerfil := httptest.NewRecorder()
		env.router.ServeHTTP(recPerfil, reqPerfil)

		if recPerfil.Code != http.StatusOK {
			t.Fatalf("GET perfil falhou: %d", recPerfil.Code)
		}
		var respPerfil struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recPerfil.Body.Bytes(), &respPerfil)
		if respPerfil.Data.AvatarURL == nil || *respPerfil.Data.AvatarURL != avatarPath {
			t.Fatalf("avatar_url no GET perfil divergiu do upload: %+v", respPerfil.Data.AvatarURL)
		}

		reqArquivo := httptest.NewRequest(http.MethodGet, avatarPath, nil)
		recArquivo := httptest.NewRecorder()
		env.router.ServeHTTP(recArquivo, reqArquivo)

		if recArquivo.Code != http.StatusOK {
			t.Fatalf("servir arquivo falhou: %d %s", recArquivo.Code, recArquivo.Body.String())
		}
		if got := recArquivo.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Fatalf("Cache-Control inesperado: %q", got)
		}
		if got := recArquivo.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("X-Content-Type-Options inesperado: %q", got)
		}
		if got := recArquivo.Header().Get("Content-Type"); !strings.HasPrefix(got, "image/jpeg") {
			t.Fatalf("Content-Type inesperado: %q", got)
		}

		caminhoFisico := filepath.Join(env.avatarDir, nomeArquivo)
		if _, err := os.Stat(caminhoFisico); err != nil {
			t.Fatalf("arquivo fisico %s nao existe no disco: %v", caminhoFisico, err)
		}

		capaURL := "https://images.igdb.com/cover_game_123.jpg"
		jogoID := criarJogoComCapa(t, env.pool, usuarioID, capaURL)

		reqCapa := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto/capa", strings.NewReader(fmt.Sprintf(`{"jogo_id": %d}`, jogoID)))
		reqCapa.Header.Set("Authorization", "Bearer "+token)
		reqCapa.Header.Set("Content-Type", "application/json")
		recCapa := httptest.NewRecorder()
		env.router.ServeHTTP(recCapa, reqCapa)

		if recCapa.Code != http.StatusOK {
			t.Fatalf("trocar por capa falhou: %d %s", recCapa.Code, recCapa.Body.String())
		}
		var respCapa struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recCapa.Body.Bytes(), &respCapa)
		if respCapa.Data.AvatarTipo == nil || *respCapa.Data.AvatarTipo != "jogo" {
			t.Fatalf("avatar_tipo esperado jogo, obteve %+v", respCapa.Data.AvatarTipo)
		}
		if respCapa.Data.AvatarURL == nil || *respCapa.Data.AvatarURL != capaURL {
			t.Fatalf("avatar_url esperado %s, obteve %+v", capaURL, respCapa.Data.AvatarURL)
		}

		if _, err := os.Stat(caminhoFisico); !os.IsNotExist(err) {
			t.Fatalf("arquivo fisico %s deveria ter sido apagado do disco apos troca por capa", caminhoFisico)
		}

		reqDelete := httptest.NewRequest(http.MethodDelete, "/api/v1/me/foto", nil)
		reqDelete.Header.Set("Authorization", "Bearer "+token)
		recDelete := httptest.NewRecorder()
		env.router.ServeHTTP(recDelete, reqDelete)

		if recDelete.Code != http.StatusOK {
			t.Fatalf("DELETE foto falhou: %d %s", recDelete.Code, recDelete.Body.String())
		}
		var respDelete struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		_ = json.Unmarshal(recDelete.Body.Bytes(), &respDelete)
		if respDelete.Data.AvatarURL != nil || respDelete.Data.AvatarTipo != nil {
			t.Fatalf("perfil deveria estar sem foto: %+v", respDelete.Data)
		}
	})

	t.Run("Jogo da capa deletado perfil volta sem foto", func(t *testing.T) {
		usuarioID := criarUsuarioFoto(t, env.pool)
		token := gerarTokenFoto(t, env.secret, usuarioID)

		capaURL := "https://images.igdb.com/cover_soft_delete.jpg"
		jogoID := criarJogoComCapa(t, env.pool, usuarioID, capaURL)

		reqCapa := httptest.NewRequest(http.MethodPut, "/api/v1/me/foto/capa", strings.NewReader(fmt.Sprintf(`{"jogo_id": %d}`, jogoID)))
		reqCapa.Header.Set("Authorization", "Bearer "+token)
		reqCapa.Header.Set("Content-Type", "application/json")
		recCapa := httptest.NewRecorder()
		env.router.ServeHTTP(recCapa, reqCapa)
		if recCapa.Code != http.StatusOK {
			t.Fatalf("definir capa falhou: %d", recCapa.Code)
		}

		_, err := env.pool.Exec(context.Background(), "UPDATE jogos_zerados SET deleted_at = now() WHERE id = $1", jogoID)
		if err != nil {
			t.Fatalf("soft delete no jogo falhou: %v", err)
		}

		reqPerfil := httptest.NewRequest(http.MethodGet, "/api/v1/me/perfil", nil)
		reqPerfil.Header.Set("Authorization", "Bearer "+token)
		recPerfil := httptest.NewRecorder()
		env.router.ServeHTTP(recPerfil, reqPerfil)

		if recPerfil.Code != http.StatusOK {
			t.Fatalf("GET perfil com jogo deletado falhou: %d %s", recPerfil.Code, recPerfil.Body.String())
		}
		var respPerfil struct {
			Data repository.PerfilUsuario `json:"data"`
		}
		if err := json.Unmarshal(recPerfil.Body.Bytes(), &respPerfil); err != nil {
			t.Fatalf("unmarshal perfil: %v", err)
		}
		if respPerfil.Data.AvatarURL != nil || respPerfil.Data.AvatarTipo != nil {
			t.Fatalf("esperava avatar_url e avatar_tipo nulos apos jogo soft deleted, obteve: url=%+v tipo=%+v", respPerfil.Data.AvatarURL, respPerfil.Data.AvatarTipo)
		}
	})
}
