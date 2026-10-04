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
	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/middleware"
	"github.com/UlerichLabs/memory-card/apps/api/internal/migration"
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
	slog.Info("cadastro de usuários", "habilitado", cfg.RegistrationEnabled)
	authCfg, err := config.LoadAuth()
	if err != nil {
		return fmt.Errorf("carregar auth: %w", err)
	}
	emailCfg, err := config.LoadEmail()
	if err != nil {
		return fmt.Errorf("carregar email: %w", err)
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
	migrated, err := migration.Apply(pool)
	if err != nil {
		return fmt.Errorf("aplicar migrations: %w", err)
	}
	if migrated {
		slog.Info("migrations aplicadas")
	} else {
		slog.Info("schema já atualizado")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORS(corsOrigins))
	if err := router.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("configurar proxies: %w", err)
	}

	queries := db.New(pool)
	usuarioRepo := repository.NewUsuarioRepository(queries)
	resetRepo := repository.NewRecuperacaoSenhaRepository(queries)
	cadastroService := service.NewCadastroService(usuarioRepo)
	authHandler := handler.NewAuthHandler(cadastroService, cfg.RegistrationEnabled)

	revogadosRepo := repository.NewTokenRevogadoRepository(queries)
	loginService, err := service.NewLoginService(usuarioRepo, tokens, revogadosRepo, resetRepo)
	if err != nil {
		return fmt.Errorf("configurar login: %w", err)
	}
	loginHandler := handler.NewLoginHandler(loginService)
	var emailSender service.EmailSender
	if emailCfg.Modo == "smtp" {
		emailSender = service.NewSMTPEmailSender(emailCfg.SMTPHost, emailCfg.SMTPPort, emailCfg.SMTPUser, emailCfg.SMTPPassword, emailCfg.SMTPFrom)
	} else {
		emailSender = service.NewLogEmailSender()
	}
	resetService := service.NewRecuperacaoSenhaService(emailCfg.ResetURL, usuarioRepo, resetRepo, revogadosRepo, emailSender)
	resetHandler := handler.NewRecuperacaoSenhaHandler(resetService)

	healthService := service.NewHealthService(pool)
	healthHandler := handler.NewHealthHandler(healthService)
	router.GET("/api/v1/health", healthHandler.Check)
	publicas := router.Group("/api/v1/auth")
	publicas.POST("/register", authHandler.Register)
	publicas.POST("/login", loginHandler.Login)
	publicas.POST("/refresh", loginHandler.Refresh)
	publicas.POST("/solicitar-reset", resetHandler.SolicitarReset)
	publicas.GET("/validar-token-reset", resetHandler.ValidarTokenReset)
	publicas.POST("/redefinir-senha", resetHandler.RedefinirSenha)
	meHandler := handler.NewMeHandler(service.NewPerfilService(usuarioRepo))
	privadas := middleware.GrupoPrivado(router, tokens)
	privadas.GET("/me", meHandler.Me)
	logoutHandler := handler.NewLogoutHandler(service.NewLogoutService(tokens, revogadosRepo))
	privadas.POST("/auth/logout", logoutHandler.Logout)
	trocaSenhaHandler := handler.NewTrocaSenhaHandler(service.NewTrocaSenhaService(usuarioRepo, resetRepo, revogadosRepo))
	privadas.POST("/auth/trocar-senha", trocaSenhaHandler.TrocarSenha)
	igdbClient := igdbclient.New(igdbclient.Config{ClientID: cfg.IGDBClientID, ClientSecret: cfg.IGDBClientSecret})
	igdbCache := repository.NewIGDBCacheRepository(queries)
	igdbService := service.NewIGDBService(igdbClient, igdbCache)
	igdbHandler := handler.NewIGDBHandler(igdbService)
	privadas.GET("/igdb/jogos/busca", igdbHandler.BuscarJogos)
	privadas.GET("/jogos/igdb/:id", igdbHandler.BuscarJogo)
	privadas.GET("/igdb/plataformas", igdbHandler.ListarPlataformas)
	privadas.GET("/igdb/plataformas/:id/jogos", igdbHandler.JogosDaPlataforma)
	privadas.POST("/igdb/plataformas/:id/jogos/refresh", igdbHandler.AtualizarJogosDaPlataforma)
	privadas.GET("/igdb/franquias/busca", igdbHandler.BuscarFranquias)
	privadas.GET("/igdb/franquias/:id/jogos", igdbHandler.JogosDaFranquia)
	privadas.POST("/igdb/franquias/:id/jogos/refresh", igdbHandler.AtualizarJogosDaFranquia)
	jogosRepo := repository.NewJogosRepository(pool, queries)
	jogosService := service.NewJogosService(jogosRepo)
	jogosHandler := handler.NewJogosHandler(jogosService)
	privadas.GET("/jogos", jogosHandler.ListarJogos)
	privadas.GET("/jogos/filtros", jogosHandler.ObterFiltros)
	privadas.GET("/jogos/game-do-ano", jogosHandler.ObterGameDoAnoResumo)
	privadas.GET("/jogos/:id", jogosHandler.ObterDetalhesJogo)
	privadas.POST("/jogos", jogosHandler.CriarJogo)
	privadas.PUT("/jogos/:id", jogosHandler.AtualizarJogo)
	privadas.PUT("/jogos/:id/game-do-ano", jogosHandler.DefinirGameDoAno)
	privadas.DELETE("/jogos/:id", jogosHandler.ExcluirJogo)
	privadas.DELETE("/jogos/:id/game-do-ano", jogosHandler.RemoverGameDoAno)
	abandonadosRepo := repository.NewJogosAbandonadosRepository(pool, queries)
	abandonadosService := service.NewJogosAbandonadosService(abandonadosRepo)
	abandonadosHandler := handler.NewJogosAbandonadosHandler(abandonadosService)
	privadas.GET("/jogos-abandonados", abandonadosHandler.Listar)
	privadas.GET("/jogos-abandonados/filtros", abandonadosHandler.ObterFiltros)
	privadas.GET("/jogos-abandonados/total", abandonadosHandler.ObterTotal)
	privadas.GET("/jogos-abandonados/:id", abandonadosHandler.ObterDetalhes)
	privadas.POST("/jogos-abandonados", abandonadosHandler.Criar)
	privadas.PUT("/jogos-abandonados/:id", abandonadosHandler.Atualizar)
	privadas.DELETE("/jogos-abandonados/:id", abandonadosHandler.Excluir)
	jogandoRepo := repository.NewJogosEmAndamentoRepository(queries)
	jogandoService := service.NewJogosEmAndamentoService(jogandoRepo)
	jogandoHandler := handler.NewJogosEmAndamentoHandler(jogandoService)
	privadas.GET("/jogando", jogandoHandler.Listar)
	privadas.POST("/jogando", jogandoHandler.Criar)
	privadas.DELETE("/jogando/:id", jogandoHandler.Excluir)
	listasRepo := repository.NewListasRepository(pool, queries)
	catalogoHandler := handler.NewCatalogoHandler(service.NewCatalogoService(igdbClient, igdbCache, listasRepo))
	privadas.GET("/igdb/generos", catalogoHandler.Generos)
	privadas.GET("/catalogo/jogos", catalogoHandler.Jogos)
	listasService := service.NewListasService(listasRepo, igdbService)
	listasHandler := handler.NewListasHandler(listasService)
	privadas.GET("/listas", listasHandler.ListarListas)
	privadas.POST("/listas", listasHandler.CriarLista)
	privadas.PUT("/listas/ordem", listasHandler.ReordenarListas)
	privadas.GET("/listas/:id", listasHandler.ObterLista)
	privadas.PUT("/listas/:id", listasHandler.AtualizarLista)
	privadas.DELETE("/listas/:id", listasHandler.ExcluirLista)
	privadas.POST("/listas/:id/itens", listasHandler.AdicionarItem)
	privadas.POST("/listas/:id/itens/lote", listasHandler.AdicionarItensLote)
	privadas.DELETE("/listas/:id/itens/:itemId", listasHandler.ExcluirItem)
	privadas.PUT("/listas/:id/ordem", listasHandler.ReordenarItens)
	privadas.PUT("/listas/:id/itens/:itemId/zeramento", listasHandler.AssociarJogoZerado)
	privadas.DELETE("/listas/:id/itens/:itemId/zeramento", listasHandler.DesassociarJogoZerado)
	dashboardRepo := repository.NewDashboardRepository(queries)
	dashboardService := service.NewDashboardService(dashboardRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	privadas.GET("/dashboard/resumo", dashboardHandler.ObterResumo)
	privadas.GET("/dashboard/por-ano", dashboardHandler.ListarPorAno)
	privadas.GET("/dashboard/ranking-plataformas", dashboardHandler.ObterRankingPlataformas)
	privadas.GET("/dashboard/ranking-generos", dashboardHandler.ObterRankingGeneros)
	privadas.GET("/dashboard/breakdown-tipo", dashboardHandler.ObterBreakdownTipo)
	privadas.GET("/dashboard/recordes", dashboardHandler.ObterRecordes)
	privadas.GET("/dashboard/notas", dashboardHandler.ObterNotas)
	privadas.GET("/dashboard/dificuldade", dashboardHandler.ObterDificuldade)

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
