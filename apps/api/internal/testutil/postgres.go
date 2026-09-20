//go:build e2e

// Package testutil fornece utilitários para suporte a testes e2e com dependências reais.
package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/UlerichLabs/memory-card/apps/api/internal/migration"
)

func init() {
	// Desabilita o container Ryuk por padrão se não configurado externamente,
	// garantindo compatibilidade com ambientes Linux com SELinux e runners de CI restritos.
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

// PostgresContainerResult agrupa os recursos do container PostgreSQL provisionado para testes.
type PostgresContainerResult struct {
	Container        *tcpostgres.PostgresContainer
	Pool             *pgxpool.Pool
	ConnectionString string
}

// SetupPostgres inicializa um container PostgreSQL isolado com migrations aplicadas.
// Registra teardown automático do container e do pool de conexões via t.Cleanup.
func SetupPostgres(t *testing.T) *PostgresContainerResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("memory_card_test"),
		tcpostgres.WithUsername("test_user"),
		tcpostgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("falha ao iniciar container postgres: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao obter connection string do container: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao interpretar configuracao do pool: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao criar pool de conexoes pgx: %v", err)
	}

	if err := runMigrations(pool); err != nil {
		pool.Close()
		_ = pgContainer.Terminate(context.Background())
		t.Fatalf("falha ao aplicar migrations do schema: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer teardownCancel()
		if err := pgContainer.Terminate(teardownCtx); err != nil {
			t.Logf("aviso: falha ao finalizar container postgres: %v", err)
		}
	})

	return &PostgresContainerResult{
		Container:        pgContainer,
		Pool:             pool,
		ConnectionString: connStr,
	}
}

func runMigrations(pool *pgxpool.Pool) error {
	_, err := migration.Apply(pool)
	return err
}
