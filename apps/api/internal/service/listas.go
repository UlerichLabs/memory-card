// Package service implementa a logica de negocio da aplicacao.
package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrListaNomeObrigatorio           = errors.New("listas.nome_obrigatorio")
	ErrListaNomeInvalido              = errors.New("listas.nome_invalido")
	ErrListaDescricaoMuitoLonga       = errors.New("listas.descricao_muito_longa")
	ErrListaTipoInvalido              = errors.New("listas.tipo_invalido")
	ErrListaRegraInvalida             = errors.New("listas.regra_invalida")
	ErrListaMetaInvalida              = errors.New("listas.meta_invalida")
	ErrListaFranquiaNaoEncontrada     = errors.New("listas.franquia_nao_encontrada")
	ErrListaCampoImutavel             = errors.New("listas.campo_imutavel")
	ErrListaNaoEncontrada             = errors.New("listas.nao_encontrada")
	ErrListaItemNaoEncontrado         = errors.New("listas.item_nao_encontrado")
	ErrListaItensNaoPermitidos        = errors.New("listas.itens_nao_permitidos")
	ErrListaItemDuplicado             = errors.New("listas.item_duplicado")
	ErrListaOrdemInvalida             = errors.New("listas.ordem_invalida")
	ErrListaSincronizacaoNaoPermitida = errors.New("listas.sincronizacao_nao_permitida")
	ErrListaIDInvalido                = errors.New("listas.id_invalido")
)

type ListasIGDBService interface {
	ObterFranquia(ctx context.Context, id int64) (*igdbclient.Franchise, error)
	JogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error)
	AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error)
}

type RegraDetalhe struct {
	Tipo   string  `json:"tipo"`
	Valor  *string `json:"valor"`
	IgdbID *int32  `json:"igdb_id"`
}

type ProgressoDetalhe struct {
	Feitos      int        `json:"feitos"`
	Meta        int        `json:"meta"`
	Percentual  int        `json:"percentual"`
	Concluido   bool       `json:"concluido"`
	ConcluidoEm *time.Time `json:"concluido_em"`
}

type JogoZeradoMatch struct {
	ID           int32     `json:"id"`
	Nota         int32     `json:"nota"`
	FinalizadoEm time.Time `json:"finalizado_em"`
}

type ListaItemDetalhe struct {
	ID            int64            `json:"id"`
	IgdbID        *int32           `json:"igdb_id"`
	Nome          string           `json:"nome"`
	Console       *string          `json:"console"`
	IgdbCapaURL   *string          `json:"igdb_capa_url"`
	AnoLancamento *int             `json:"ano_lancamento"`
	Posicao       int              `json:"posicao"`
	Origem        string           `json:"origem"`
	Zerado        bool             `json:"zerado"`
	JogoZerado    *JogoZeradoMatch `json:"jogo_zerado"`
}

type ListaResumo struct {
	ID             int64             `json:"id"`
	Tipo           string            `json:"tipo"`
	Nome           string            `json:"nome"`
	Descricao      *string           `json:"descricao"`
	Regra          *RegraDetalhe     `json:"regra"`
	Meta           *int              `json:"meta"`
	TotalItens     int               `json:"total_itens"`
	ItensPendentes int               `json:"itens_pendentes"`
	Progresso      *ProgressoDetalhe `json:"progresso"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type ListaDetalhada struct {
	ListaResumo
	Itens []*ListaItemDetalhe `json:"itens"`
}

type SincronizarResultado struct {
	ListaDetalhada
	Adicionados int `json:"adicionados"`
}

type CriarListaRegraInput struct {
	Tipo   string `json:"tipo"`
	Valor  string `json:"valor"`
	IgdbID *int32 `json:"igdb_id"`
}

type CriarListaInput struct {
	UsuarioID int32
	Tipo      string
	Nome      string
	Descricao *string
	Regra     *CriarListaRegraInput
	Meta      *int
}

type AtualizarListaInput struct {
	ID        int64
	UsuarioID int32
	Nome      *string
	Descricao *string
	Meta      *int
	Tipo      *string
	Regra     *CriarListaRegraInput
}

type AdicionarItemInput struct {
	ListaID       int64
	UsuarioID     int32
	IgdbID        *int32
	Nome          string
	Console       *string
	IgdbCapaURL   *string
	AnoLancamento *int
}

type ListasService struct {
	repo repository.ListasRepository
	igdb ListasIGDBService
}

func NewListasService(repo repository.ListasRepository, igdb ListasIGDBService) *ListasService {
	return &ListasService{repo: repo, igdb: igdb}
}

func normalizeLowerUnaccent(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	res, _, err := transform.String(t, s)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(s))
	}
	return strings.ToLower(strings.TrimSpace(res))
}

func (s *ListasService) CriarLista(ctx context.Context, input CriarListaInput) (*ListaDetalhada, error) {
	nome := strings.TrimSpace(input.Nome)
	if nome == "" {
		return nil, ErrListaNomeObrigatorio
	}
	if len([]rune(nome)) > 100 {
		return nil, ErrListaNomeInvalido
	}

	var descPtr *string
	if input.Descricao != nil {
		d := strings.TrimSpace(*input.Descricao)
		if len([]rune(d)) > 200 {
			return nil, ErrListaDescricaoMuitoLonga
		}
		if d != "" {
			descPtr = &d
		}
	}

	if input.Tipo != "fila" && input.Tipo != "desafio" {
		return nil, ErrListaTipoInvalido
	}

	if input.Tipo == "fila" {
		if input.Regra != nil || input.Meta != nil {
			return nil, ErrListaRegraInvalida
		}

		lista, err := s.repo.Criar(ctx, repository.CriarListaParams{
			UsuarioID: input.UsuarioID,
			Tipo:      "fila",
			Nome:      nome,
			Descricao: descPtr,
		})
		if err != nil {
			return nil, err
		}

		return s.ObterLista(ctx, lista.ID, input.UsuarioID)
	}

	if input.Regra == nil {
		return nil, ErrListaRegraInvalida
	}

	regraTipo := strings.TrimSpace(input.Regra.Tipo)
	if regraTipo != "franquia" && regraTipo != "plataforma" && regraTipo != "genero" && regraTipo != "manual" {
		return nil, ErrListaRegraInvalida
	}

	switch regraTipo {
	case "franquia":
		if input.Regra.IgdbID == nil || *input.Regra.IgdbID <= 0 {
			return nil, ErrListaRegraInvalida
		}
		if input.Meta != nil && (*input.Meta < 1 || *input.Meta > 10000) {
			return nil, ErrListaMetaInvalida
		}

		franchise, err := s.igdb.ObterFranquia(ctx, int64(*input.Regra.IgdbID))
		if err != nil {
			if errors.Is(err, ErrJogoIGDBNaoEncontrado) {
				return nil, ErrListaFranquiaNaoEncontrada
			}
			return nil, err
		}
		if franchise == nil {
			return nil, ErrListaFranquiaNaoEncontrada
		}

		games, err := s.igdb.JogosDaFranquia(ctx, int64(*input.Regra.IgdbID))
		if err != nil {
			return nil, err
		}

		sort.SliceStable(games, func(i, j int) bool {
			dateA := int64(1<<62 - 1)
			if games[i].FirstReleaseDate != nil {
				dateA = *games[i].FirstReleaseDate
			}
			dateB := int64(1<<62 - 1)
			if games[j].FirstReleaseDate != nil {
				dateB = *games[j].FirstReleaseDate
			}
			if dateA != dateB {
				return dateA < dateB
			}
			return games[i].ID < games[j].ID
		})

		seenIDs := make(map[int32]bool, len(games))
		itensParams := make([]repository.CriarItemParams, 0, len(games))
		pos := 1
		for _, g := range games {
			gid := int32(g.ID)
			if seenIDs[gid] {
				continue
			}
			seenIDs[gid] = true

			var capaURL *string
			if g.Cover != nil {
				g.Cover.EnsureURL()
				if g.Cover.URL != "" {
					c := g.Cover.URL
					capaURL = &c
				}
			}
			var ano *int
			if g.FirstReleaseDate != nil {
				y := time.Unix(*g.FirstReleaseDate, 0).UTC().Year()
				ano = &y
			}

			itensParams = append(itensParams, repository.CriarItemParams{
				IgdbID:        &gid,
				Nome:          g.Name,
				IgdbCapaURL:   capaURL,
				AnoLancamento: ano,
				Posicao:       pos,
			})
			pos++
		}

		lista, _, err := s.repo.CriarComItens(ctx, repository.CriarListaComItensParams{
			Lista: repository.CriarListaParams{
				UsuarioID:   input.UsuarioID,
				Tipo:        "desafio",
				Nome:        nome,
				Descricao:   descPtr,
				RegraTipo:   &regraTipo,
				RegraValor:  &franchise.Name,
				RegraIgdbID: input.Regra.IgdbID,
				Meta:        input.Meta,
			},
			Itens: itensParams,
		})
		if err != nil {
			return nil, err
		}

		return s.ObterLista(ctx, lista.ID, input.UsuarioID)

	case "plataforma", "genero":
		valor := strings.TrimSpace(input.Regra.Valor)
		if valor == "" || len([]rune(valor)) > 150 {
			return nil, ErrListaRegraInvalida
		}
		if input.Meta == nil || *input.Meta < 1 || *input.Meta > 10000 {
			return nil, ErrListaMetaInvalida
		}

		lista, err := s.repo.Criar(ctx, repository.CriarListaParams{
			UsuarioID:  input.UsuarioID,
			Tipo:       "desafio",
			Nome:       nome,
			Descricao:  descPtr,
			RegraTipo:  &regraTipo,
			RegraValor: &valor,
			Meta:       input.Meta,
		})
		if err != nil {
			return nil, err
		}

		return s.ObterLista(ctx, lista.ID, input.UsuarioID)

	case "manual":
		if input.Meta != nil && (*input.Meta < 1 || *input.Meta > 10000) {
			return nil, ErrListaMetaInvalida
		}

		lista, err := s.repo.Criar(ctx, repository.CriarListaParams{
			UsuarioID: input.UsuarioID,
			Tipo:      "desafio",
			Nome:      nome,
			Descricao: descPtr,
			RegraTipo: &regraTipo,
			Meta:      input.Meta,
		})
		if err != nil {
			return nil, err
		}

		return s.ObterLista(ctx, lista.ID, input.UsuarioID)
	}

	return nil, ErrListaRegraInvalida
}

func (s *ListasService) ObterLista(ctx context.Context, id int64, usuarioID int32) (*ListaDetalhada, error) {
	lista, err := s.repo.BuscarPorID(ctx, id, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	if lista == nil {
		return nil, ErrListaNaoEncontrada
	}

	jogos, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	var dbItens []*repository.ListaItem
	if lista.Tipo != "desafio" || lista.RegraTipo == nil || (*lista.RegraTipo != "plataforma" && *lista.RegraTipo != "genero") {
		dbItens, err = s.repo.ListarItensPorLista(ctx, id, usuarioID)
		if err != nil {
			return nil, err
		}
	}

	resumo, itens := s.calcularLista(lista, jogos, dbItens)

	return &ListaDetalhada{
		ListaResumo: *resumo,
		Itens:       itens,
	}, nil
}

func (s *ListasService) ListarListas(ctx context.Context, usuarioID int32) ([]*ListaResumo, error) {
	listas, err := s.repo.ListarPorUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	todosItens, err := s.repo.ListarTodosItensDoUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	jogos, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	itensPorLista := make(map[int64][]*repository.ListaItem)
	for _, it := range todosItens {
		itensPorLista[it.ListaID] = append(itensPorLista[it.ListaID], it)
	}

	res := make([]*ListaResumo, 0, len(listas))
	for _, l := range listas {
		var listaItens []*repository.ListaItem
		if l.Tipo != "desafio" || l.RegraTipo == nil || (*l.RegraTipo != "plataforma" && *l.RegraTipo != "genero") {
			listaItens = itensPorLista[l.ID]
		}
		resumo, _ := s.calcularLista(l, jogos, listaItens)
		res = append(res, resumo)
	}

	return res, nil
}

func (s *ListasService) AtualizarLista(ctx context.Context, input AtualizarListaInput) (*ListaDetalhada, error) {
	if input.Tipo != nil || input.Regra != nil {
		return nil, ErrListaCampoImutavel
	}

	lista, err := s.repo.BuscarPorID(ctx, input.ID, input.UsuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	if lista == nil {
		return nil, ErrListaNaoEncontrada
	}

	nome := lista.Nome
	if input.Nome != nil {
		n := strings.TrimSpace(*input.Nome)
		if n == "" {
			return nil, ErrListaNomeObrigatorio
		}
		if len([]rune(n)) > 100 {
			return nil, ErrListaNomeInvalido
		}
		nome = n
	}

	descPtr := lista.Descricao
	if input.Descricao != nil {
		d := strings.TrimSpace(*input.Descricao)
		if len([]rune(d)) > 200 {
			return nil, ErrListaDescricaoMuitoLonga
		}
		if d == "" {
			descPtr = nil
		} else {
			descPtr = &d
		}
	}

	metaPtr := lista.Meta
	if lista.Tipo == "fila" {
		if input.Meta != nil {
			return nil, ErrListaMetaInvalida
		}
		metaPtr = nil
	} else {
		if lista.RegraTipo != nil && (*lista.RegraTipo == "plataforma" || *lista.RegraTipo == "genero") {
			if input.Meta != nil {
				if *input.Meta < 1 || *input.Meta > 10000 {
					return nil, ErrListaMetaInvalida
				}
				metaPtr = input.Meta
			}
		} else {
			if input.Meta != nil {
				if *input.Meta < 1 || *input.Meta > 10000 {
					return nil, ErrListaMetaInvalida
				}
				metaPtr = input.Meta
			}
		}
	}

	_, err = s.repo.Atualizar(ctx, repository.AtualizarListaParams{
		ID:        input.ID,
		UsuarioID: input.UsuarioID,
		Nome:      nome,
		Descricao: descPtr,
		Meta:      metaPtr,
	})
	if err != nil {
		return nil, err
	}

	return s.ObterLista(ctx, input.ID, input.UsuarioID)
}

func (s *ListasService) ExcluirLista(ctx context.Context, id int64, usuarioID int32) error {
	err := s.repo.Excluir(ctx, id, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrListaNaoEncontrada
		}
		return err
	}
	return nil
}

func (s *ListasService) AdicionarItem(ctx context.Context, input AdicionarItemInput) (*ListaItemDetalhe, error) {
	lista, err := s.repo.BuscarPorID(ctx, input.ListaID, input.UsuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	if lista == nil {
		return nil, ErrListaNaoEncontrada
	}

	if lista.Tipo == "desafio" && lista.RegraTipo != nil && (*lista.RegraTipo == "plataforma" || *lista.RegraTipo == "genero") {
		return nil, ErrListaItensNaoPermitidos
	}

	nome := strings.TrimSpace(input.Nome)
	if nome == "" {
		return nil, ErrListaNomeObrigatorio
	}
	if len([]rune(nome)) > 200 {
		return nil, ErrListaNomeInvalido
	}

	item, err := s.repo.CriarItem(ctx, input.ListaID, input.UsuarioID, repository.CriarItemParams{
		IgdbID:        input.IgdbID,
		Nome:          nome,
		Console:       input.Console,
		IgdbCapaURL:   input.IgdbCapaURL,
		AnoLancamento: input.AnoLancamento,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation || pgErr.Code == "23505" {
				return nil, ErrListaItemDuplicado
			}
		}
		return nil, err
	}

	jogos, err := s.repo.ListarJogosZeradosUsuario(ctx, input.UsuarioID)
	if err != nil {
		return nil, err
	}

	return s.construirItemDetalhe(item, jogos), nil
}

func (s *ListasService) ExcluirItem(ctx context.Context, itemID int64, listaID int64, usuarioID int32) error {
	err := s.repo.ExcluirItemERecompactar(ctx, itemID, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrListaItemNaoEncontrado
		}
		return err
	}
	return nil
}

func (s *ListasService) ReordenarItens(ctx context.Context, listaID int64, usuarioID int32, itemIDs []int64) ([]*ListaItemDetalhe, error) {
	lista, err := s.repo.BuscarPorID(ctx, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	if lista == nil {
		return nil, ErrListaNaoEncontrada
	}

	itens, err := s.repo.ListarItensPorLista(ctx, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	if len(itemIDs) != len(itens) {
		return nil, ErrListaOrdemInvalida
	}

	validIDs := make(map[int64]bool, len(itens))
	for _, it := range itens {
		validIDs[it.ID] = true
	}

	seen := make(map[int64]bool, len(itemIDs))
	for _, id := range itemIDs {
		if !validIDs[id] || seen[id] {
			return nil, ErrListaOrdemInvalida
		}
		seen[id] = true
	}

	reordenados, err := s.repo.ReordenarItens(ctx, listaID, usuarioID, itemIDs)
	if err != nil {
		return nil, err
	}

	jogos, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	detalhes := make([]*ListaItemDetalhe, 0, len(reordenados))
	for _, it := range reordenados {
		detalhes = append(detalhes, s.construirItemDetalhe(it, jogos))
	}
	return detalhes, nil
}

func (s *ListasService) AssociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32, jogoZeradoID int32) (*ListaItemDetalhe, error) {
	lista, err := s.repo.BuscarPorID(ctx, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	if lista == nil {
		return nil, ErrListaNaoEncontrada
	}

	item, err := s.repo.BuscarItemPorID(ctx, itemID, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaItemNaoEncontrado
		}
		return nil, err
	}
	if item == nil {
		return nil, ErrListaItemNaoEncontrado
	}

	jogo, err := s.repo.BuscarJogoZeradoUsuario(ctx, jogoZeradoID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJogoNaoEncontrado
		}
		return nil, err
	}
	if jogo == nil {
		return nil, ErrJogoNaoEncontrado
	}

	itemAtualizado, err := s.repo.AssociarJogoZerado(ctx, itemID, listaID, usuarioID, jogoZeradoID)
	if err != nil {
		return nil, err
	}

	jogos, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	return s.construirItemDetalhe(itemAtualizado, jogos), nil
}

func (s *ListasService) DesassociarJogoZerado(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*ListaItemDetalhe, error) {
	lista, err := s.repo.BuscarPorID(ctx, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	if lista == nil {
		return nil, ErrListaNaoEncontrada
	}

	item, err := s.repo.BuscarItemPorID(ctx, itemID, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaItemNaoEncontrado
		}
		return nil, err
	}
	if item == nil {
		return nil, ErrListaItemNaoEncontrado
	}

	itemAtualizado, err := s.repo.DesassociarJogoZerado(ctx, itemID, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	jogos, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	return s.construirItemDetalhe(itemAtualizado, jogos), nil
}

func (s *ListasService) SincronizarFranquia(ctx context.Context, listaID int64, usuarioID int32) (*SincronizarResultado, error) {
	lista, err := s.repo.BuscarPorID(ctx, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	if lista == nil {
		return nil, ErrListaNaoEncontrada
	}

	if lista.Tipo != "desafio" || lista.RegraTipo == nil || *lista.RegraTipo != "franquia" || lista.RegraIgdbID == nil {
		return nil, ErrListaSincronizacaoNaoPermitida
	}

	games, err := s.igdb.AtualizarJogosDaFranquia(ctx, int64(*lista.RegraIgdbID))
	if err != nil {
		return nil, err
	}

	itensExistentes, err := s.repo.ListarItensPorLista(ctx, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	igdbIDsExistentes := make(map[int32]bool, len(itensExistentes))
	for _, it := range itensExistentes {
		if it.IgdbID != nil {
			igdbIDsExistentes[*it.IgdbID] = true
		}
	}

	novosJogos := make([]igdbclient.Game, 0)
	for _, g := range games {
		gid := int32(g.ID)
		if !igdbIDsExistentes[gid] {
			novosJogos = append(novosJogos, g)
		}
	}

	sort.SliceStable(novosJogos, func(i, j int) bool {
		dateA := int64(1<<62 - 1)
		if novosJogos[i].FirstReleaseDate != nil {
			dateA = *novosJogos[i].FirstReleaseDate
		}
		dateB := int64(1<<62 - 1)
		if novosJogos[j].FirstReleaseDate != nil {
			dateB = *novosJogos[j].FirstReleaseDate
		}
		if dateA != dateB {
			return dateA < dateB
		}
		return novosJogos[i].ID < novosJogos[j].ID
	})

	novosParams := make([]repository.CriarItemParams, 0, len(novosJogos))
	for _, g := range novosJogos {
		gid := int32(g.ID)
		var capaURL *string
		if g.Cover != nil {
			g.Cover.EnsureURL()
			if g.Cover.URL != "" {
				c := g.Cover.URL
				capaURL = &c
			}
		}
		var ano *int
		if g.FirstReleaseDate != nil {
			y := time.Unix(*g.FirstReleaseDate, 0).UTC().Year()
			ano = &y
		}
		novosParams = append(novosParams, repository.CriarItemParams{
			IgdbID:        &gid,
			Nome:          g.Name,
			IgdbCapaURL:   capaURL,
			AnoLancamento: ano,
		})
	}

	adicionados, err := s.repo.SincronizarFranquia(ctx, listaID, usuarioID, novosParams)
	if err != nil {
		return nil, err
	}

	listaCompleta, err := s.ObterLista(ctx, listaID, usuarioID)
	if err != nil {
		return nil, err
	}

	return &SincronizarResultado{
		ListaDetalhada: *listaCompleta,
		Adicionados:    adicionados,
	}, nil
}

func (s *ListasService) calcularLista(lista *repository.Lista, jogos []*repository.JogoZeradoResumo, itens []*repository.ListaItem) (*ListaResumo, []*ListaItemDetalhe) {
	var regra *RegraDetalhe
	if lista.RegraTipo != nil {
		regra = &RegraDetalhe{
			Tipo:   *lista.RegraTipo,
			Valor:  lista.RegraValor,
			IgdbID: lista.RegraIgdbID,
		}
	}

	if lista.Tipo == "desafio" && lista.RegraTipo != nil && (*lista.RegraTipo == "plataforma" || *lista.RegraTipo == "genero") {
		var matchedJogos []*repository.JogoZeradoResumo
		regraValorNorm := ""
		if lista.RegraValor != nil {
			regraValorNorm = normalizeLowerUnaccent(*lista.RegraValor)
		}

		for _, j := range jogos {
			if *lista.RegraTipo == "plataforma" {
				if normalizeLowerUnaccent(j.Console) == regraValorNorm {
					matchedJogos = append(matchedJogos, j)
				}
			} else if *lista.RegraTipo == "genero" {
				if strings.Contains(normalizeLowerUnaccent(j.Genero), regraValorNorm) {
					matchedJogos = append(matchedJogos, j)
				}
			}
		}

		feitos := len(matchedJogos)
		meta := 0
		if lista.Meta != nil {
			meta = *lista.Meta
		} else {
			meta = feitos
		}

		percentual := 0
		if meta > 0 {
			feitosLimitados := feitos
			if feitosLimitados > meta {
				feitosLimitados = meta
			}
			percentual = (feitosLimitados * 100) / meta
		}

		concluido := meta > 0 && feitos >= meta
		var concluidoEm *time.Time
		if concluido && len(matchedJogos) >= meta {
			t := matchedJogos[meta-1].FinalizadoEm
			concluidoEm = &t
		}

		progresso := &ProgressoDetalhe{
			Feitos:      feitos,
			Meta:        meta,
			Percentual:  percentual,
			Concluido:   concluido,
			ConcluidoEm: concluidoEm,
		}

		itensDetalhe := make([]*ListaItemDetalhe, 0, len(matchedJogos))
		for idx, j := range matchedJogos {
			consoleStr := j.Console
			itensDetalhe = append(itensDetalhe, &ListaItemDetalhe{
				ID:          int64(j.ID),
				IgdbID:      j.IgdbID,
				Nome:        j.Nome,
				Console:     &consoleStr,
				IgdbCapaURL: j.IgdbCapaURL,
				Posicao:     idx + 1,
				Origem:      "regra",
				Zerado:      true,
				JogoZerado: &JogoZeradoMatch{
					ID:           j.ID,
					Nota:         j.Nota,
					FinalizadoEm: j.FinalizadoEm,
				},
			})
		}

		resumo := &ListaResumo{
			ID:             lista.ID,
			Tipo:           lista.Tipo,
			Nome:           lista.Nome,
			Descricao:      lista.Descricao,
			Regra:          regra,
			Meta:           lista.Meta,
			TotalItens:     len(matchedJogos),
			ItensPendentes: 0,
			Progresso:      progresso,
			CreatedAt:      lista.CreatedAt,
			UpdatedAt:      lista.UpdatedAt,
		}

		return resumo, itensDetalhe
	}

	detalhesItens := make([]*ListaItemDetalhe, 0, len(itens))
	var matchedFinalizados []time.Time
	feitos := 0

	for _, it := range itens {
		d := s.construirItemDetalhe(it, jogos)
		detalhesItens = append(detalhesItens, d)
		if d.Zerado {
			feitos++
			if d.JogoZerado != nil {
				matchedFinalizados = append(matchedFinalizados, d.JogoZerado.FinalizadoEm)
			}
		}
	}

	totalItens := len(itens)
	var progresso *ProgressoDetalhe
	var itensPendentes int

	if lista.Tipo == "fila" {
		itensPendentes = totalItens - feitos
		progresso = nil
	} else {
		meta := totalItens
		if lista.Meta != nil {
			meta = *lista.Meta
		}
		itensPendentes = totalItens - feitos
		if itensPendentes < 0 {
			itensPendentes = 0
		}
		percentual := 0
		if meta > 0 {
			feitosLimitados := feitos
			if feitosLimitados > meta {
				feitosLimitados = meta
			}
			percentual = (feitosLimitados * 100) / meta
		}
		concluido := meta > 0 && feitos >= meta
		var concluidoEm *time.Time
		if concluido && len(matchedFinalizados) >= meta {
			sort.Slice(matchedFinalizados, func(i, j int) bool {
				return matchedFinalizados[i].Before(matchedFinalizados[j])
			})
			t := matchedFinalizados[meta-1]
			concluidoEm = &t
		}
		progresso = &ProgressoDetalhe{
			Feitos:      feitos,
			Meta:        meta,
			Percentual:  percentual,
			Concluido:   concluido,
			ConcluidoEm: concluidoEm,
		}
	}

	resumo := &ListaResumo{
		ID:             lista.ID,
		Tipo:           lista.Tipo,
		Nome:           lista.Nome,
		Descricao:      lista.Descricao,
		Regra:          regra,
		Meta:           lista.Meta,
		TotalItens:     totalItens,
		ItensPendentes: itensPendentes,
		Progresso:      progresso,
		CreatedAt:      lista.CreatedAt,
		UpdatedAt:      lista.UpdatedAt,
	}

	return resumo, detalhesItens
}

func (s *ListasService) construirItemDetalhe(item *repository.ListaItem, jogos []*repository.JogoZeradoResumo) *ListaItemDetalhe {
	var match *repository.JogoZeradoResumo

	if item.JogoZeradoID != nil {
		for _, j := range jogos {
			if j.ID == *item.JogoZeradoID {
				match = j
				break
			}
		}
	}

	if match == nil && item.IgdbID != nil {
		for _, j := range jogos {
			if j.IgdbID != nil && *j.IgdbID == *item.IgdbID {
				match = j
				break
			}
		}
	}

	if match == nil && item.IgdbID == nil {
		normItemNome := normalizeLowerUnaccent(item.Nome)
		for _, j := range jogos {
			if normalizeLowerUnaccent(j.Nome) == normItemNome {
				match = j
				break
			}
		}
	}

	res := &ListaItemDetalhe{
		ID:            item.ID,
		IgdbID:        item.IgdbID,
		Nome:          item.Nome,
		Console:       item.Console,
		IgdbCapaURL:   item.IgdbCapaURL,
		AnoLancamento: item.AnoLancamento,
		Posicao:       item.Posicao,
		Origem:        "item",
		Zerado:        match != nil,
	}

	if match != nil {
		res.JogoZerado = &JogoZeradoMatch{
			ID:           match.ID,
			Nota:         match.Nota,
			FinalizadoEm: match.FinalizadoEm,
		}
	}

	return res
}
