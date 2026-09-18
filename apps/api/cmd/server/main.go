package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/UlerichLabs/memory-card/apps/api/internal/config"
	"github.com/UlerichLabs/memory-card/apps/api/internal/handler"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("servidor encerrado", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("carregar configuração: %w", err)
	}
	authCfg, err := config.LoadAuth()
	if err != nil {
		return fmt.Errorf("carregar auth: %w", err)
	}
	corsOrigins, err := config.LoadCORS()
	if err != nil {
		return fmt.Errorf("carregar CORS: %w", err)
	}
	tokens, err := service.NewAuthToken(authCfg.Secret, authCfg.AccessTTL, authCfg.RefreshTTL)
	if err != nil {
		return fmt.Errorf("configurar tokens: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("criar pool PostgreSQL: %w", err)
	}
	defer pool.Close()
	startupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = pool.Ping(startupCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("conectar ao PostgreSQL: %w", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORS(corsOrigins))
	if err := router.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("configurar proxies: %w", err)
	}

	queries := db.New(pool)
	usuarioRepo := repository.NewUsuarioRepository(queries)
	cadastroService := service.NewCadastroService(usuarioRepo)
	authHandler := handler.NewAuthHandler(cadastroService)

	revogadosRepo := repository.NewTokenRevogadoRepository(queries)
	loginService, err := service.NewLoginService(usuarioRepo, tokens, revogadosRepo)
	if err != nil {
		return fmt.Errorf("configurar login: %w", err)
	}
	loginHandler := handler.NewLoginHandler(loginService)

	healthService := service.NewHealthService(pool)
	healthHandler := handler.NewHealthHandler(healthService)
	router.GET("/api/v1/health", healthHandler.Check)
	publicas := router.Group("/api/v1/auth")
	publicas.POST("/register", authHandler.Register)
	publicas.POST("/login", loginHandler.Login)
	publicas.POST("/refresh", loginHandler.Refresh)
	meHandler := handler.NewMeHandler(service.NewPerfilService(usuarioRepo))
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/me", meHandler.Me)
	logoutHandler := handler.NewLogoutHandler(service.NewLogoutService(tokens, revogadosRepo))
	privadas.POST("/auth/logout", logoutHandler.Logout)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()
	slog.Info("servidor iniciado", "port", cfg.Port)
	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("servir HTTP: %w", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			closeErr := server.Close()
			return fmt.Errorf("encerrar HTTP: %w", errors.Join(err, closeErr))
		}
	}
	return nil
}
