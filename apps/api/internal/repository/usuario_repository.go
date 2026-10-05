package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

var (
	ErrUsernameEmUso             = errors.New("perfil.username_em_uso")
	ErrJogoFavoritoNaoEncontrado = errors.New("perfil.jogo_favorito_nao_encontrado")
)

type Usuario struct {
	ID        int32     `json:"id"`
	Nome      string    `json:"nome"`
	Email     string    `json:"email"`
	Idioma    string    `json:"idioma"`
	CreatedAt time.Time `json:"created_at"`
}

type JogoFavoritoResumo struct {
	ID      int32   `json:"id"`
	Nome    string  `json:"nome"`
	CapaURL *string `json:"capa_url"`
}

type PerfilUsuario struct {
	Nome                string              `json:"nome"`
	Email               string              `json:"email"`
	Username            *string             `json:"username"`
	Bio                 *string             `json:"bio"`
	AvatarURL           *string             `json:"avatar_url"`
	JogoFavorito        *JogoFavoritoResumo `json:"jogo_favorito"`
	ConsoleFavorito     *string             `json:"console_favorito"`
	JogandoDesde        *int16              `json:"jogando_desde"`
	JogandoDesdeEfetivo *int16              `json:"jogando_desde_efetivo"`
}

type AtualizarPerfilParams struct {
	ID              int32
	Nome            string
	Username        *string
	Bio             *string
	JogoFavoritoID  *int32
	ConsoleFavorito *string
	JogandoDesde    *int16
}

type UsuarioRepository interface {
	ExistePorEmail(ctx context.Context, email string) (bool, error)
	Criar(ctx context.Context, nome, email, senhaHash string) (*Usuario, error)
	BuscarPorEmail(ctx context.Context, email string) (*CredenciaisUsuario, error)
	BuscarCredenciaisPorID(ctx context.Context, id int32) (*CredenciaisUsuario, error)
	AtualizarSenha(ctx context.Context, id int32, senhaHash string) error
	BuscarPorID(ctx context.Context, id int32) (*Usuario, error)
	BuscarPerfil(ctx context.Context, id int32) (*PerfilUsuario, error)
	AtualizarPerfil(ctx context.Context, params AtualizarPerfilParams) (*PerfilUsuario, error)
}

type SQLUsuarioRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewUsuarioRepository(queries *db.Queries, pool ...*pgxpool.Pool) *SQLUsuarioRepository {
	var p *pgxpool.Pool
	if len(pool) > 0 {
		p = pool[0]
	}
	return &SQLUsuarioRepository{queries: queries, pool: p}
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

func (r *SQLUsuarioRepository) BuscarCredenciaisPorID(ctx context.Context, id int32) (*CredenciaisUsuario, error) {
	row, err := r.queries.BuscarCredenciaisUsuarioPorID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscar credenciais por id: %w", err)
	}
	return &CredenciaisUsuario{Usuario: Usuario{ID: row.ID, Nome: row.Nome, Email: row.Email, Idioma: row.Idioma, CreatedAt: row.CreatedAt.Time}, SenhaHash: row.SenhaHash}, nil
}

func (r *SQLUsuarioRepository) AtualizarSenha(ctx context.Context, id int32, senhaHash string) error {
	if err := r.queries.AtualizarSenhaUsuario(ctx, db.AtualizarSenhaUsuarioParams{ID: id, SenhaHash: senhaHash}); err != nil {
		return fmt.Errorf("atualizar senha do usuario: %w", err)
	}
	return nil
}

func (r *SQLUsuarioRepository) BuscarPorID(ctx context.Context, id int32) (*Usuario, error) {
	row, err := r.queries.BuscarUsuarioPorID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscar perfil: %w", err)
	}
	return &Usuario{ID: row.ID, Nome: row.Nome, Email: row.Email, Idioma: row.Idioma, CreatedAt: row.CreatedAt.Time}, nil
}

func (r *SQLUsuarioRepository) BuscarPerfil(ctx context.Context, id int32) (*PerfilUsuario, error) {
	row, err := r.queries.BuscarPerfilCompletoPorID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("buscar perfil completo: %w", err)
	}
	return mapearPerfilUsuario(row), nil
}

func (r *SQLUsuarioRepository) AtualizarPerfil(ctx context.Context, params AtualizarPerfilParams) (*PerfilUsuario, error) {
	var (
		qtx = r.queries
		tx  pgx.Tx
		err error
	)

	if r.pool != nil {
		tx, err = r.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("iniciar transacao: %w", err)
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback(ctx)
			}
		}()
		qtx = r.queries.WithTx(tx)
	}

	if params.JogoFavoritoID != nil {
		existe, errVerif := qtx.ExisteJogoZeradoDoUsuario(ctx, db.ExisteJogoZeradoDoUsuarioParams{
			ID:        *params.JogoFavoritoID,
			UsuarioID: params.ID,
		})
		if errVerif != nil {
			return nil, fmt.Errorf("verificar jogo favorito: %w", errVerif)
		}
		if !existe {
			return nil, ErrJogoFavoritoNaoEncontrado
		}
	}

	var username pgtype.Text
	if params.Username != nil && *params.Username != "" {
		username = pgtype.Text{String: *params.Username, Valid: true}
	}
	var bio pgtype.Text
	if params.Bio != nil && *params.Bio != "" {
		bio = pgtype.Text{String: *params.Bio, Valid: true}
	}
	var jogoFavoritoID pgtype.Int4
	if params.JogoFavoritoID != nil {
		jogoFavoritoID = pgtype.Int4{Int32: *params.JogoFavoritoID, Valid: true}
	}
	var consoleFavorito pgtype.Text
	if params.ConsoleFavorito != nil && *params.ConsoleFavorito != "" {
		consoleFavorito = pgtype.Text{String: *params.ConsoleFavorito, Valid: true}
	}
	var jogandoDesde pgtype.Int2
	if params.JogandoDesde != nil {
		jogandoDesde = pgtype.Int2{Int16: *params.JogandoDesde, Valid: true}
	}

	_, err = qtx.AtualizarPerfilUsuario(ctx, db.AtualizarPerfilUsuarioParams{
		ID:              params.ID,
		Nome:            params.Nome,
		Username:        username,
		Bio:             bio,
		JogoFavoritoID:  jogoFavoritoID,
		ConsoleFavorito: consoleFavorito,
		JogandoDesde:    jogandoDesde,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, ErrUsernameEmUso
		}
		return nil, fmt.Errorf("atualizar perfil no banco: %w", err)
	}

	row, err := qtx.BuscarPerfilCompletoPorID(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("buscar perfil apos atualizacao: %w", err)
	}

	if tx != nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit transacao: %w", err)
		}
		tx = nil
	}

	return mapearPerfilUsuario(row), nil
}

func mapearPerfilUsuario(row db.BuscarPerfilCompletoPorIDRow) *PerfilUsuario {
	perfil := &PerfilUsuario{
		Nome:  row.Nome,
		Email: row.Email,
	}
	if row.Username.Valid && row.Username.String != "" {
		perfil.Username = &row.Username.String
	}
	if row.Bio.Valid && row.Bio.String != "" {
		perfil.Bio = &row.Bio.String
	}
	if row.AvatarUrl.Valid && row.AvatarUrl.String != "" {
		perfil.AvatarURL = &row.AvatarUrl.String
	}
	if row.ConsoleFavorito.Valid && row.ConsoleFavorito.String != "" {
		perfil.ConsoleFavorito = &row.ConsoleFavorito.String
	}
	if row.JogandoDesde.Valid {
		valor := row.JogandoDesde.Int16
		perfil.JogandoDesde = &valor
	}

	if perfil.JogandoDesde != nil {
		perfil.JogandoDesdeEfetivo = perfil.JogandoDesde
	} else if row.PrimeiroAnoZerado > 0 {
		ano := int16(row.PrimeiroAnoZerado)
		perfil.JogandoDesdeEfetivo = &ano
	}

	if row.FavID.Valid {
		var capaURL *string
		if row.FavCapaUrl.Valid && row.FavCapaUrl.String != "" {
			capaURL = &row.FavCapaUrl.String
		}
		perfil.JogoFavorito = &JogoFavoritoResumo{
			ID:      row.FavID.Int32,
			Nome:    row.FavNome.String,
			CapaURL: capaURL,
		}
	}

	return perfil
}
