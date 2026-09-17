package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type Usuario struct {
	ID        int32     `json:"id"`
	Nome      string    `json:"nome"`
	Email     string    `json:"email"`
	Idioma    string    `json:"idioma"`
	CreatedAt time.Time `json:"created_at"`
}

type UsuarioRepository interface {
	ExistePorEmail(ctx context.Context, email string) (bool, error)
	Criar(ctx context.Context, nome, email, senhaHash string) (*Usuario, error)
}

type SQLUsuarioRepository struct {
	queries *db.Queries
}

func NewUsuarioRepository(queries *db.Queries) *SQLUsuarioRepository {
	return &SQLUsuarioRepository{queries: queries}
}

func (r *SQLUsuarioRepository) ExistePorEmail(ctx context.Context, email string) (bool, error) {
	return r.queries.ExisteUsuarioComEmail(ctx, email)
}

func (r *SQLUsuarioRepository) Criar(ctx context.Context, nome, email, senhaHash string) (*Usuario, error) {
	row, err := r.queries.CriarUsuario(ctx, db.CriarUsuarioParams{
		Nome:      nome,
		Email:     email,
		SenhaHash: senhaHash,
	})
	if err != nil {
		return nil, err
	}

	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	return &Usuario{
		ID:        row.ID,
		Nome:      row.Nome,
		Email:     row.Email,
		Idioma:    row.Idioma,
		CreatedAt: createdAt,
	}, nil
}

type CredenciaisUsuario struct {
	Usuario
	SenhaHash string `json:"-"`
}

func (r *SQLUsuarioRepository) BuscarPorEmail(ctx context.Context, email string) (*CredenciaisUsuario, error) {
	row, err := r.queries.BuscarUsuarioPorEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscar usuario: %w", err)
	}
	return &CredenciaisUsuario{Usuario: Usuario{ID: row.ID, Nome: row.Nome, Email: row.Email, Idioma: row.Idioma, CreatedAt: row.CreatedAt.Time}, SenhaHash: row.SenhaHash}, nil
}
