//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
	"github.com/UlerichLabs/memory-card/apps/api/internal/testutil"
)

func TestJogandoAgoraBaixaAoCriarJogo(t *testing.T) {
	pg := testutil.SetupPostgres(t)
	queries := db.New(pg.Pool)
	zerados := repository.NewJogosRepository(pg.Pool, queries)
	abandonados := repository.NewJogosAbandonadosRepository(pg.Pool, queries)
	andamento := repository.NewJogosEmAndamentoRepository(queries)
	ctx := context.Background()

	t.Run("zerado prioriza igdb id", func(t *testing.T) {
		usuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "igdb")
		criarAndamentoE2E(t, ctx, andamento, usuarioID, int32ptr(101), "Nome diferente", time.Now().Add(-time.Hour))
		if _, err := zerados.Criar(ctx, zeradoParams(usuarioID, int32ptr(101), "Jogo zerado")); err != nil {
			t.Fatal(err)
		}
		assertAndamentoCountE2E(t, ctx, andamento, usuarioID, 0)
	})

	t.Run("zerado usa nome normalizado", func(t *testing.T) {
		usuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "nome")
		criarAndamentoE2E(t, ctx, andamento, usuarioID, nil, "  Kázé  ", time.Now().Add(-time.Hour))
		if _, err := zerados.Criar(ctx, zeradoParams(usuarioID, nil, " kAZE ")); err != nil {
			t.Fatal(err)
		}
		assertAndamentoCountE2E(t, ctx, andamento, usuarioID, 0)
	})

	t.Run("nao confunde nome parecido nem ids diferentes", func(t *testing.T) {
		usuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "nao-match")
		criarAndamentoE2E(t, ctx, andamento, usuarioID, int32ptr(202), "Kaze and the Wild Masks (Steam)", time.Now().Add(-time.Hour))
		criarAndamentoE2E(t, ctx, andamento, usuarioID, nil, "Kaze and the Wild Masks", time.Now().Add(-2*time.Hour))
		if _, err := zerados.Criar(ctx, zeradoParams(usuarioID, int32ptr(303), "Kaze and the Wild Masks (Steam)")); err != nil {
			t.Fatal(err)
		}
		assertAndamentoCountE2E(t, ctx, andamento, usuarioID, 2)
	})

	t.Run("baixa apenas o andamento mais antigo e respeita usuario", func(t *testing.T) {
		usuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "mais-antigo")
		outroUsuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "outro-usuario")
		criarAndamentoE2E(t, ctx, andamento, usuarioID, int32ptr(404), "Duplicado", time.Now().Add(-2*time.Hour))
		criarAndamentoE2E(t, ctx, andamento, usuarioID, int32ptr(404), "Duplicado", time.Now().Add(-time.Hour))
		criarAndamentoE2E(t, ctx, andamento, outroUsuarioID, int32ptr(404), "Duplicado", time.Now().Add(-3*time.Hour))
		if _, err := zerados.Criar(ctx, zeradoParams(usuarioID, int32ptr(404), "Duplicado")); err != nil {
			t.Fatal(err)
		}
		jogos, err := andamento.Listar(ctx, usuarioID)
		if err != nil {
			t.Fatal(err)
		}
		if len(jogos) != 1 {
			t.Fatalf("esperava um andamento restante, obteve %d", len(jogos))
		}
		assertAndamentoCountE2E(t, ctx, andamento, outroUsuarioID, 1)
	})

	t.Run("sem correspondente cria normalmente", func(t *testing.T) {
		usuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "sem-match")
		if _, err := zerados.Criar(ctx, zeradoParams(usuarioID, nil, "Novo jogo")); err != nil {
			t.Fatal(err)
		}
		assertAndamentoCountE2E(t, ctx, andamento, usuarioID, 0)
	})

	t.Run("abandonado baixa e tambem cria sem correspondente", func(t *testing.T) {
		usuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "abandonado")
		criarAndamentoE2E(t, ctx, andamento, usuarioID, int32ptr(505), "Abandonado", time.Now().Add(-time.Hour))
		if _, err := abandonados.Criar(ctx, abandonoParams(usuarioID, int32ptr(505), "Abandonado")); err != nil {
			t.Fatal(err)
		}
		assertAndamentoCountE2E(t, ctx, andamento, usuarioID, 0)
		if _, err := abandonados.Criar(ctx, abandonoParams(usuarioID, nil, "Outro abandonado")); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("erro na criacao faz rollback", func(t *testing.T) {
		usuarioID := criarUsuarioE2E(t, ctx, pg.Pool, "rollback")
		criarAndamentoE2E(t, ctx, andamento, usuarioID, int32ptr(606), "Rollback", time.Now().Add(-time.Hour))
		params := zeradoParams(usuarioID, int32ptr(606), "Rollback")
		params.Nota = 12
		if _, err := zerados.Criar(ctx, params); err == nil {
			t.Fatal("esperava erro de validacao do banco")
		}
		assertAndamentoCountE2E(t, ctx, andamento, usuarioID, 1)
	})
}

func criarUsuarioE2E(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sufixo string) int32 {
	t.Helper()
	var id int32
	valor := time.Now().UnixNano()
	err := pool.QueryRow(ctx, `
		INSERT INTO usuarios (nome, email, senha_hash, username)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "Teste "+sufixo, fmt.Sprintf("%d-%s@example.com", valor, sufixo), "hash", fmt.Sprintf("%d-%s", valor, sufixo)).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func criarAndamentoE2E(t *testing.T, ctx context.Context, repo repository.JogosEmAndamentoRepository, usuarioID int32, igdbID *int32, nome string, iniciadoEm time.Time) {
	t.Helper()
	if _, err := repo.Criar(ctx, repository.CriarJogoEmAndamentoParams{UsuarioID: usuarioID, IgdbID: igdbID, Nome: nome, IniciadoEm: iniciadoEm}); err != nil {
		t.Fatal(err)
	}
}

func assertAndamentoCountE2E(t *testing.T, ctx context.Context, repo repository.JogosEmAndamentoRepository, usuarioID int32, esperado int) {
	t.Helper()
	jogos, err := repo.Listar(ctx, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if len(jogos) != esperado {
		t.Fatalf("esperava %d jogos em andamento, obteve %d", esperado, len(jogos))
	}
}

func zeradoParams(usuarioID int32, igdbID *int32, nome string) repository.CriarJogoZeradoParams {
	return repository.CriarJogoZeradoParams{
		UsuarioID: usuarioID, IgdbID: igdbID, Nome: nome, Console: "PC", FinalizadoEm: time.Now(), Nota: 8,
		Dificuldade: "B", TempoJogado: 3600,
	}
}

func abandonoParams(usuarioID int32, igdbID *int32, nome string) repository.CriarJogoAbandonadoParams {
	return repository.CriarJogoAbandonadoParams{UsuarioID: usuarioID, IgdbID: igdbID, Nome: nome, Console: "PC", AbandonadoEm: time.Now()}
}

func int32ptr(value int32) *int32 { return &value }
