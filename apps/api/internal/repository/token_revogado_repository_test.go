package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type mockDBTX struct {
	exec     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	query    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRow func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockDBTX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.exec != nil {
		return m.exec(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, errors.New("exec nao implementado no mock")
}

func (m *mockDBTX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.query != nil {
		return m.query(ctx, sql, args...)
	}
	return nil, errors.New("query nao implementado no mock")
}

func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRow != nil {
		return m.queryRow(ctx, sql, args...)
	}
	return mockRow{err: errors.New("queryRow nao implementado no mock")}
}

type mockRow struct {
	scan func(dest ...any) error
	err  error
}

func (r mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if r.scan != nil {
		return r.scan(dest...)
	}
	return nil
}

func TestTokenRevogadoRepository_Revogar(t *testing.T) {
	ctx := context.Background()
	expiraEm := time.Now().Add(time.Hour).Truncate(time.Microsecond)
	jti := "test-jti-123"
	dbErr := errors.New("db connection failure")

	t.Run("sucesso ao revogar token", func(t *testing.T) {
		execCalled := false
		mock := &mockDBTX{
			exec: func(gotCtx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				execCalled = true
				if gotCtx != ctx {
					t.Error("contexto incorreto")
				}
				if len(args) != 2 {
					t.Fatalf("esperava 2 argumentos, recebeu %d", len(args))
				}
				if args[0] != jti {
					t.Errorf("jti = %v, esperado %s", args[0], jti)
				}
				argExp, ok := args[1].(pgtype.Timestamptz)
				if !ok || !argExp.Time.Equal(expiraEm) {
					t.Errorf("expira_em incorreto: %v", args[1])
				}
				return pgconn.CommandTag{}, nil
			},
		}

		repo := NewTokenRevogadoRepository(db.New(mock))
		err := repo.Revogar(ctx, jti, expiraEm)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if !execCalled {
			t.Fatal("exec nao foi chamado")
		}
	})

	t.Run("falha no banco ao revogar propaga erro", func(t *testing.T) {
		mock := &mockDBTX{
			exec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, dbErr
			},
		}

		repo := NewTokenRevogadoRepository(db.New(mock))
		err := repo.Revogar(ctx, jti, expiraEm)
		if !errors.Is(err, dbErr) {
			t.Fatalf("esperava erro %v, recebeu %v", dbErr, err)
		}
	})
}

func TestTokenRevogadoRepository_EstaRevogado(t *testing.T) {
	ctx := context.Background()
	jti := "check-jti-456"
	dbErr := errors.New("query failure")

	t.Run("token revogado retorna true", func(t *testing.T) {
		limpezaChamada := false
		consultaChamada := false
		mock := &mockDBTX{
			exec: func(gotCtx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				limpezaChamada = true
				return pgconn.CommandTag{}, nil
			},
			queryRow: func(gotCtx context.Context, sql string, args ...any) pgx.Row {
				consultaChamada = true
				if len(args) != 1 || args[0] != jti {
					t.Errorf("argumentos incorretos na consulta: %v", args)
				}
				return mockRow{
					scan: func(dest ...any) error {
						if len(dest) != 1 {
							t.Fatalf("esperava 1 destino, recebeu %d", len(dest))
						}
						val, ok := dest[0].(*bool)
						if !ok {
							t.Fatal("destino nao e *bool")
						}
						*val = true
						return nil
					},
				}
			},
		}

		repo := NewTokenRevogadoRepository(db.New(mock))
		revogado, err := repo.EstaRevogado(ctx, jti)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if !revogado {
			t.Fatal("esperava revogado = true")
		}
		if !limpezaChamada || !consultaChamada {
			t.Fatalf("limpeza=%v, consulta=%v", limpezaChamada, consultaChamada)
		}
	})

	t.Run("token nao revogado retorna false", func(t *testing.T) {
		mock := &mockDBTX{
			exec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			queryRow: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return mockRow{
					scan: func(dest ...any) error {
						val := dest[0].(*bool)
						*val = false
						return nil
					},
				}
			},
		}

		repo := NewTokenRevogadoRepository(db.New(mock))
		revogado, err := repo.EstaRevogado(ctx, jti)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if revogado {
			t.Fatal("esperava revogado = false")
		}
	})

	t.Run("falha ao limpar expirados propaga erro", func(t *testing.T) {
		mock := &mockDBTX{
			exec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, dbErr
			},
		}

		repo := NewTokenRevogadoRepository(db.New(mock))
		revogado, err := repo.EstaRevogado(ctx, jti)
		if !errors.Is(err, dbErr) {
			t.Fatalf("esperava erro %v, recebeu %v", dbErr, err)
		}
		if revogado {
			t.Fatal("revogado deve ser false em caso de erro")
		}
	})

	t.Run("falha ao consultar existencia propaga erro", func(t *testing.T) {
		mock := &mockDBTX{
			exec: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
			queryRow: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return mockRow{err: dbErr}
			},
		}

		repo := NewTokenRevogadoRepository(db.New(mock))
		revogado, err := repo.EstaRevogado(ctx, jti)
		if !errors.Is(err, dbErr) {
			t.Fatalf("esperava erro %v, recebeu %v", dbErr, err)
		}
		if revogado {
			t.Fatal("revogado deve ser false em caso de erro")
		}
	})
}
