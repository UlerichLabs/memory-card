// Package repository contem as implementacoes de persistencia.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/UlerichLabs/memory-card/apps/api/internal/repository/db"
)

type Lista struct {
	ID          int64
	UsuarioID   int32
	Tipo        string
	Nome        string
	Descricao   *string
	RegraTipo   *string
	RegraValor  *string
	RegraIgdbID *int32
	Meta        *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListaItem struct {
	ID            int64
	ListaID       int64
	IgdbID        *int32
	Nome          string
	Console       *string
	IgdbCapaURL   *string
	AnoLancamento *int
	Posicao       int
	JogoZeradoID  *int32
	Ignorado      bool
	CreatedAt     time.Time
}

type JogoZeradoResumo struct {
	ID           int32
	UsuarioID    int32
	IgdbID       *int32
	Nome         string
	Console      string
	Genero       string
	FinalizadoEm time.Time
	Nota         int32
	IgdbCapaURL  *string
}

type CriarListaParams struct {
	UsuarioID   int32
	Tipo        string
	Nome        string
	Descricao   *string
	RegraTipo   *string
	RegraValor  *string
	RegraIgdbID *int32
	Meta        *int
}

type CriarItemParams struct {
	IgdbID        *int32
	Nome          string
	Console       *string
	IgdbCapaURL   *string
	AnoLancamento *int
	Posicao       int
	JogoZeradoID  *int32
	Ignorado      bool
}

type CriarListaComItensParams struct {
	Lista CriarListaParams
	Itens []CriarItemParams
}

type AdicionarItensLoteResultado struct {
	Adicionados  int
	JaExistentes int
}

type AtualizarListaParams struct {
	ID        int64
	UsuarioID int32
	Nome      string
	Descricao *string
	Meta      *int
}

type ListasRepository interface {
	Criar(ctx context.Context, params CriarListaParams) (*Lista, error)
	CriarComItens(ctx context.Context, params CriarListaComItensParams) (*Lista, []*ListaItem, error)
	BuscarPorID(ctx context.Context, id int64, usuarioID int32) (*Lista, error)
	ListarPorUsuario(ctx context.Context, usuarioID int32) ([]*Lista, error)
	Atualizar(ctx context.Context, params AtualizarListaParams) (*Lista, error)
	Excluir(ctx context.Context, id int64, usuarioID int32) error

	CriarItem(ctx context.Context, listaID int64, usuarioID int32, params CriarItemParams) (*ListaItem, error)
	AdicionarItensLote(ctx context.Context, listaID int64, usuarioID int32, itens []CriarItemParams) (AdicionarItensLoteResultado, error)
	BuscarItemPorID(ctx context.Context, id int64, listaID int64, usuarioID int32) (*ListaItem, error)
	ListarItensPorLista(ctx context.Context, listaID int64, usuarioID int32) ([]*ListaItem, error)
	ListarTodosItensDoUsuario(ctx context.Context, usuarioID int32) ([]*ListaItem, error)
	ExcluirItemERecompactar(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error
	ReordenarItens(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*ListaItem, error)
	AssociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*ListaItem, error)
	DesassociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*ListaItem, error)
	SincronizarFranquia(ctx context.Context, listaID int64, usuarioID int32, novosItens []CriarItemParams) (int, error)

	ListarJogosZeradosUsuario(ctx context.Context, usuarioID int32) ([]*JogoZeradoResumo, error)
	BuscarJogoZeradoUsuario(ctx context.Context, id int32, usuarioID int32) (*JogoZeradoResumo, error)
}

type SQLListasRepository struct {
	pool    DBTXPool
	queries *db.Queries
}

func NewListasRepository(pool DBTXPool, queries *db.Queries) *SQLListasRepository {
	return &SQLListasRepository{pool: pool, queries: queries}
}

func (r *SQLListasRepository) Criar(ctx context.Context, params CriarListaParams) (*Lista, error) {
	dbParams := db.CriarListaParams{
		UsuarioID: params.UsuarioID,
		Tipo:      db.ListaTipo(params.Tipo),
		Nome:      params.Nome,
	}
	if params.Descricao != nil {
		dbParams.Descricao = pgtype.Text{String: *params.Descricao, Valid: true}
	}
	if params.RegraTipo != nil {
		dbParams.RegraTipo = db.NullListaRegraTipo{ListaRegraTipo: db.ListaRegraTipo(*params.RegraTipo), Valid: true}
	}
	if params.RegraValor != nil {
		dbParams.RegraValor = pgtype.Text{String: *params.RegraValor, Valid: true}
	}
	if params.RegraIgdbID != nil {
		dbParams.RegraIgdbID = pgtype.Int4{Int32: *params.RegraIgdbID, Valid: true}
	}

	row, err := r.queries.CriarLista(ctx, dbParams)
	if err != nil {
		return nil, fmt.Errorf("criar lista: %w", err)
	}

	return mapearLista(row), nil
}

func (r *SQLListasRepository) CriarComItens(ctx context.Context, params CriarListaComItensParams) (*Lista, []*ListaItem, error) {
	if r.pool == nil {
		return nil, nil, errors.New("pool nao configurado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("iniciar transacao: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := r.queries.WithTx(tx)

	dbParams := db.CriarListaParams{
		UsuarioID: params.Lista.UsuarioID,
		Tipo:      db.ListaTipo(params.Lista.Tipo),
		Nome:      params.Lista.Nome,
	}
	if params.Lista.Descricao != nil {
		dbParams.Descricao = pgtype.Text{String: *params.Lista.Descricao, Valid: true}
	}
	if params.Lista.RegraTipo != nil {
		dbParams.RegraTipo = db.NullListaRegraTipo{ListaRegraTipo: db.ListaRegraTipo(*params.Lista.RegraTipo), Valid: true}
	}
	if params.Lista.RegraValor != nil {
		dbParams.RegraValor = pgtype.Text{String: *params.Lista.RegraValor, Valid: true}
	}
	if params.Lista.RegraIgdbID != nil {
		dbParams.RegraIgdbID = pgtype.Int4{Int32: *params.Lista.RegraIgdbID, Valid: true}
	}

	rowLista, err := qtx.CriarLista(ctx, dbParams)
	if err != nil {
		return nil, nil, fmt.Errorf("criar lista: %w", err)
	}

	itensCriados := make([]*ListaItem, 0, len(params.Itens))
	for _, it := range params.Itens {
		itParam := db.CriarItemListaParams{
			ListaID: rowLista.ID,
			Nome:    it.Nome,
			Posicao: int32(it.Posicao),
		}
		if it.IgdbID != nil {
			itParam.IgdbID = pgtype.Int4{Int32: *it.IgdbID, Valid: true}
		}
		if it.Console != nil {
			itParam.Console = pgtype.Text{String: *it.Console, Valid: true}
		}
		if it.IgdbCapaURL != nil {
			itParam.IgdbCapaUrl = pgtype.Text{String: *it.IgdbCapaURL, Valid: true}
		}
		if it.AnoLancamento != nil {
			itParam.AnoLancamento = pgtype.Int4{Int32: int32(*it.AnoLancamento), Valid: true}
		}
		if it.JogoZeradoID != nil {
			itParam.JogoZeradoID = pgtype.Int4{Int32: *it.JogoZeradoID, Valid: true}
		}

		itemRow, err := qtx.CriarItemLista(ctx, itParam)
		if err != nil {
			return nil, nil, fmt.Errorf("criar item lista: %w", err)
		}
		itensCriados = append(itensCriados, mapearListaItem(itemRow))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit transacao: %w", err)
	}

	return mapearLista(rowLista), itensCriados, nil
}

func (r *SQLListasRepository) BuscarPorID(ctx context.Context, id int64, usuarioID int32) (*Lista, error) {
	row, err := r.queries.BuscarListaPorID(ctx, db.BuscarListaPorIDParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("buscar lista por id: %w", err)
	}
	return mapearLista(row), nil
}

func (r *SQLListasRepository) ListarPorUsuario(ctx context.Context, usuarioID int32) ([]*Lista, error) {
	rows, err := r.queries.ListarListasPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar listas por usuario: %w", err)
	}

	res := make([]*Lista, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapearLista(row))
	}
	return res, nil
}

func (r *SQLListasRepository) Atualizar(ctx context.Context, params AtualizarListaParams) (*Lista, error) {
	dbParams := db.AtualizarListaParams{
		ID:        params.ID,
		UsuarioID: params.UsuarioID,
		Nome:      params.Nome,
	}
	if params.Descricao != nil {
		dbParams.Descricao = pgtype.Text{String: *params.Descricao, Valid: true}
	}

	row, err := r.queries.AtualizarLista(ctx, dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("atualizar lista: %w", err)
	}
	return mapearLista(row), nil
}

func (r *SQLListasRepository) Excluir(ctx context.Context, id int64, usuarioID int32) error {
	rowsAffected, err := r.queries.ExcluirLista(ctx, db.ExcluirListaParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		return fmt.Errorf("excluir lista: %w", err)
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *SQLListasRepository) CriarItem(ctx context.Context, listaID int64, usuarioID int32, params CriarItemParams) (*ListaItem, error) {
	_, err := r.BuscarPorID(ctx, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	maxPos, err := r.queries.ObterMaxPosicaoItem(ctx, listaID)
	if err != nil {
		return nil, fmt.Errorf("obter max posicao: %w", err)
	}

	itParam := db.CriarItemListaParams{
		ListaID: listaID,
		Nome:    params.Nome,
		Posicao: maxPos + 1,
	}
	if params.IgdbID != nil {
		itParam.IgdbID = pgtype.Int4{Int32: *params.IgdbID, Valid: true}
	}
	if params.Console != nil {
		itParam.Console = pgtype.Text{String: *params.Console, Valid: true}
	}
	if params.IgdbCapaURL != nil {
		itParam.IgdbCapaUrl = pgtype.Text{String: *params.IgdbCapaURL, Valid: true}
	}
	if params.AnoLancamento != nil {
		itParam.AnoLancamento = pgtype.Int4{Int32: int32(*params.AnoLancamento), Valid: true}
	}
	if params.JogoZeradoID != nil {
		itParam.JogoZeradoID = pgtype.Int4{Int32: *params.JogoZeradoID, Valid: true}
	}

	row, err := r.queries.CriarItemLista(ctx, itParam)
	if err != nil {
		return nil, fmt.Errorf("criar item lista: %w", err)
	}
	return mapearListaItem(row), nil
}

func (r *SQLListasRepository) AdicionarItensLote(ctx context.Context, listaID int64, usuarioID int32, itens []CriarItemParams) (AdicionarItensLoteResultado, error) {
	if r.pool == nil {
		return AdicionarItensLoteResultado{}, errors.New("pool nao configurado")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return AdicionarItensLoteResultado{}, fmt.Errorf("iniciar transacao: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := r.queries.WithTx(tx)
	lista, err := qtx.BuscarListaPorIDParaUpdate(ctx, db.BuscarListaPorIDParaUpdateParams{ID: listaID, UsuarioID: usuarioID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdicionarItensLoteResultado{}, pgx.ErrNoRows
		}
		return AdicionarItensLoteResultado{}, fmt.Errorf("buscar lista para lote: %w", err)
	}
	existentes, err := qtx.ListarItensPorLista(ctx, lista.ID)
	if err != nil {
		return AdicionarItensLoteResultado{}, fmt.Errorf("listar itens para lote: %w", err)
	}
	ids := make(map[int32]struct{}, len(existentes))
	for _, item := range existentes {
		if item.IgdbID.Valid {
			ids[item.IgdbID.Int32] = struct{}{}
		}
	}
	resultado := AdicionarItensLoteResultado{}
	posicao := int32(0)
	for _, item := range existentes {
		if item.Posicao > posicao {
			posicao = item.Posicao
		}
	}
	for _, item := range itens {
		if _, found := ids[*item.IgdbID]; found {
			resultado.JaExistentes++
			continue
		}
		ids[*item.IgdbID] = struct{}{}
		posicao++
		_, err = qtx.CriarItemLista(ctx, toDBCriarItem(listaID, posicao, item))
		if err != nil {
			return AdicionarItensLoteResultado{}, fmt.Errorf("criar item em lote: %w", err)
		}
		resultado.Adicionados++
	}
	if err := tx.Commit(ctx); err != nil {
		return AdicionarItensLoteResultado{}, fmt.Errorf("commit lote: %w", err)
	}
	return resultado, nil
}

func toDBCriarItem(listaID int64, posicao int32, item CriarItemParams) db.CriarItemListaParams {
	params := db.CriarItemListaParams{ListaID: listaID, Posicao: posicao, Nome: item.Nome}
	if item.IgdbID != nil {
		params.IgdbID = pgtype.Int4{Int32: *item.IgdbID, Valid: true}
	}
	if item.Console != nil {
		params.Console = pgtype.Text{String: *item.Console, Valid: true}
	}
	if item.IgdbCapaURL != nil {
		params.IgdbCapaUrl = pgtype.Text{String: *item.IgdbCapaURL, Valid: true}
	}
	if item.AnoLancamento != nil {
		params.AnoLancamento = pgtype.Int4{Int32: int32(*item.AnoLancamento), Valid: true}
	}
	if item.JogoZeradoID != nil {
		params.JogoZeradoID = pgtype.Int4{Int32: *item.JogoZeradoID, Valid: true}
	}
	return params
}

func (r *SQLListasRepository) BuscarItemPorID(ctx context.Context, id int64, listaID int64, usuarioID int32) (*ListaItem, error) {
	row, err := r.queries.BuscarItemPorID(ctx, db.BuscarItemPorIDParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("buscar item por id: %w", err)
	}
	if row.ListaID != listaID {
		return nil, pgx.ErrNoRows
	}
	return mapearListaItem(row), nil
}

func (r *SQLListasRepository) ListarItensPorLista(ctx context.Context, listaID int64, usuarioID int32) ([]*ListaItem, error) {
	_, err := r.BuscarPorID(ctx, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListarItensPorLista(ctx, listaID)
	if err != nil {
		return nil, fmt.Errorf("listar itens por lista: %w", err)
	}

	res := make([]*ListaItem, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapearListaItem(row))
	}
	return res, nil
}

func (r *SQLListasRepository) ListarTodosItensDoUsuario(ctx context.Context, usuarioID int32) ([]*ListaItem, error) {
	rows, err := r.queries.ListarTodosItensDoUsuario(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar todos itens do usuario: %w", err)
	}

	res := make([]*ListaItem, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapearListaItem(row))
	}
	return res, nil
}

func (r *SQLListasRepository) ExcluirItemERecompactar(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error {
	if r.pool == nil {
		return errors.New("pool nao configurado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciar transacao: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := r.queries.WithTx(tx)

	_, err = qtx.BuscarItemPorID(ctx, db.BuscarItemPorIDParams{
		ID:        itemID,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("verificar item: %w", err)
	}

	rowsAffected, err := qtx.ExcluirItem(ctx, db.ExcluirItemParams{
		ID:      itemID,
		ListaID: listaID,
	})
	if err != nil {
		return fmt.Errorf("excluir item: %w", err)
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}

	restantes, err := qtx.ListarItensPorLista(ctx, listaID)
	if err != nil {
		return fmt.Errorf("listar itens restantes: %w", err)
	}

	for i, it := range restantes {
		novaPos := int32(i + 1)
		if it.Posicao != novaPos {
			if err := qtx.AtualizarPosicaoItem(ctx, db.AtualizarPosicaoItemParams{
				ID:      it.ID,
				ListaID: listaID,
				Posicao: novaPos,
			}); err != nil {
				return fmt.Errorf("recompactar posicao item %d: %w", it.ID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transacao: %w", err)
	}

	return nil
}

func (r *SQLListasRepository) ReordenarItens(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*ListaItem, error) {
	if r.pool == nil {
		return nil, errors.New("pool nao configurado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("iniciar transacao: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := r.queries.WithTx(tx)

	_, err = qtx.BuscarListaPorIDParaUpdate(ctx, db.BuscarListaPorIDParaUpdateParams{
		ID:        listaID,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("bloquear lista: %w", err)
	}

	itensAtuais, err := qtx.ListarItensPorLista(ctx, listaID)
	if err != nil {
		return nil, fmt.Errorf("listar itens atuais: %w", err)
	}

	if len(itemIDs) != len(itensAtuais) {
		return nil, errors.New("quantidade de itens divergente")
	}

	atuaisMap := make(map[int64]bool, len(itensAtuais))
	for _, it := range itensAtuais {
		atuaisMap[it.ID] = true
	}

	vistos := make(map[int64]bool, len(itemIDs))
	for _, id := range itemIDs {
		if !atuaisMap[id] || vistos[id] {
			return nil, errors.New("item id invalido ou duplicado na reordenacao")
		}
		vistos[id] = true
	}

	for i, id := range itemIDs {
		novaPos := int32(i + 1)
		if err := qtx.AtualizarPosicaoItem(ctx, db.AtualizarPosicaoItemParams{
			ID:      id,
			ListaID: listaID,
			Posicao: novaPos,
		}); err != nil {
			return nil, fmt.Errorf("atualizar posicao item %d: %w", id, err)
		}
	}

	itensOrdenados, err := qtx.ListarItensPorLista(ctx, listaID)
	if err != nil {
		return nil, fmt.Errorf("listar itens ordenados: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transacao: %w", err)
	}

	res := make([]*ListaItem, 0, len(itensOrdenados))
	for _, row := range itensOrdenados {
		res = append(res, mapearListaItem(row))
	}
	return res, nil
}

func (r *SQLListasRepository) AssociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*ListaItem, error) {
	_, err := r.BuscarItemPorID(ctx, itemID, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.AssociarJogoZeradoItem(ctx, db.AssociarJogoZeradoItemParams{
		ID:           itemID,
		ListaID:      listaID,
		JogoZeradoID: pgtype.Int4{Int32: jogoZeradoID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("associar jogo zerado item: %w", err)
	}
	return mapearListaItem(row), nil
}

func (r *SQLListasRepository) DesassociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*ListaItem, error) {
	_, err := r.BuscarItemPorID(ctx, itemID, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.DesassociarJogoZeradoItem(ctx, db.DesassociarJogoZeradoItemParams{
		ID:      itemID,
		ListaID: listaID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("desassociar jogo zerado item: %w", err)
	}
	return mapearListaItem(row), nil
}

func (r *SQLListasRepository) SincronizarFranquia(ctx context.Context, listaID int64, usuarioID int32, novosItens []CriarItemParams) (int, error) {
	if r.pool == nil {
		return 0, errors.New("pool nao configurado")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("iniciar transacao: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := r.queries.WithTx(tx)

	_, err = qtx.BuscarListaPorIDParaUpdate(ctx, db.BuscarListaPorIDParaUpdateParams{
		ID:        listaID,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, pgx.ErrNoRows
		}
		return 0, fmt.Errorf("bloquear lista: %w", err)
	}

	itensAtuais, err := qtx.ListarItensPorLista(ctx, listaID)
	if err != nil {
		return 0, fmt.Errorf("listar itens atuais: %w", err)
	}

	igdbExistentes := make(map[int32]bool, len(itensAtuais))
	maxPos := int32(0)
	for _, it := range itensAtuais {
		if it.IgdbID.Valid {
			igdbExistentes[it.IgdbID.Int32] = true
		}
		if it.Posicao > maxPos {
			maxPos = it.Posicao
		}
	}

	adicionados := 0
	for _, it := range novosItens {
		if it.IgdbID == nil || igdbExistentes[*it.IgdbID] {
			continue
		}

		maxPos++
		itParam := db.CriarItemListaParams{
			ListaID: listaID,
			Nome:    it.Nome,
			Posicao: maxPos,
		}
		if it.IgdbID != nil {
			itParam.IgdbID = pgtype.Int4{Int32: *it.IgdbID, Valid: true}
		}
		if it.Console != nil {
			itParam.Console = pgtype.Text{String: *it.Console, Valid: true}
		}
		if it.IgdbCapaURL != nil {
			itParam.IgdbCapaUrl = pgtype.Text{String: *it.IgdbCapaURL, Valid: true}
		}
		if it.AnoLancamento != nil {
			itParam.AnoLancamento = pgtype.Int4{Int32: int32(*it.AnoLancamento), Valid: true}
		}

		if _, err := qtx.CriarItemLista(ctx, itParam); err != nil {
			return 0, fmt.Errorf("inserir item sincronizado: %w", err)
		}
		igdbExistentes[*it.IgdbID] = true
		adicionados++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit transacao: %w", err)
	}

	return adicionados, nil
}

func (r *SQLListasRepository) ListarJogosZeradosUsuario(ctx context.Context, usuarioID int32) ([]*JogoZeradoResumo, error) {
	rows, err := r.queries.ListarJogosZeradosUsuarioParaMatching(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar jogos zerados para matching: %w", err)
	}

	res := make([]*JogoZeradoResumo, 0, len(rows))
	for _, row := range rows {
		item := &JogoZeradoResumo{
			ID:           row.ID,
			UsuarioID:    row.UsuarioID,
			Nome:         row.Nome,
			Console:      row.Console,
			FinalizadoEm: row.FinalizadoEm.Time,
			Nota:         row.Nota,
		}
		if row.IgdbID.Valid {
			item.IgdbID = &row.IgdbID.Int32
		}
		if row.Genero.Valid {
			item.Genero = row.Genero.String
		}
		if row.IgdbCapaUrl.Valid && row.IgdbCapaUrl.String != "" {
			item.IgdbCapaURL = &row.IgdbCapaUrl.String
		}
		res = append(res, item)
	}
	return res, nil
}

func (r *SQLListasRepository) BuscarJogoZeradoUsuario(ctx context.Context, id int32, usuarioID int32) (*JogoZeradoResumo, error) {
	row, err := r.queries.BuscarJogoZeradoDoUsuario(ctx, db.BuscarJogoZeradoDoUsuarioParams{
		ID:        id,
		UsuarioID: usuarioID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("buscar jogo zerado do usuario: %w", err)
	}

	item := &JogoZeradoResumo{
		ID:           row.ID,
		UsuarioID:    row.UsuarioID,
		Nome:         row.Nome,
		Console:      row.Console,
		FinalizadoEm: row.FinalizadoEm.Time,
		Nota:         row.Nota,
	}
	if row.IgdbID.Valid {
		item.IgdbID = &row.IgdbID.Int32
	}
	if row.Genero.Valid {
		item.Genero = row.Genero.String
	}
	if row.IgdbCapaUrl.Valid && row.IgdbCapaUrl.String != "" {
		item.IgdbCapaURL = &row.IgdbCapaUrl.String
	}
	return item, nil
}

func mapearLista(row db.Lista) *Lista {
	l := &Lista{
		ID:        row.ID,
		UsuarioID: row.UsuarioID,
		Tipo:      string(row.Tipo),
		Nome:      row.Nome,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
	if row.Descricao.Valid {
		l.Descricao = &row.Descricao.String
	}
	if row.RegraTipo.Valid {
		rt := string(row.RegraTipo.ListaRegraTipo)
		l.RegraTipo = &rt
	}
	if row.RegraValor.Valid {
		l.RegraValor = &row.RegraValor.String
	}
	if row.RegraIgdbID.Valid {
		l.RegraIgdbID = &row.RegraIgdbID.Int32
	}
	if row.Meta.Valid {
		m := int(row.Meta.Int32)
		l.Meta = &m
	}
	return l
}

func mapearListaItem(row db.ListaIten) *ListaItem {
	it := &ListaItem{
		ID:        row.ID,
		ListaID:   row.ListaID,
		Nome:      row.Nome,
		Posicao:   int(row.Posicao),
		Ignorado:  false,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.IgdbID.Valid {
		it.IgdbID = &row.IgdbID.Int32
	}
	if row.Console.Valid {
		it.Console = &row.Console.String
	}
	if row.IgdbCapaUrl.Valid {
		it.IgdbCapaURL = &row.IgdbCapaUrl.String
	}
	if row.AnoLancamento.Valid {
		ano := int(row.AnoLancamento.Int32)
		it.AnoLancamento = &ano
	}
	if row.JogoZeradoID.Valid {
		it.JogoZeradoID = &row.JogoZeradoID.Int32
	}
	return it
}
