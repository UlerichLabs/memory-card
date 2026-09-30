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
	ErrListaFranquiaSemJogos          = errors.New("listas.franquia_sem_jogos")
	ErrListaRestauracaoNaoPermitida   = errors.New("listas.restauracao_nao_permitida")
	ErrListaCampoImutavel             = errors.New("listas.campo_imutavel")
	ErrListaNaoEncontrada             = errors.New("listas.nao_encontrada")
	ErrListaItemNaoEncontrado         = errors.New("listas.item_nao_encontrado")
	ErrListaItensNaoPermitidos        = errors.New("listas.itens_nao_permitidos")
	ErrListaItemDuplicado             = errors.New("listas.item_duplicado")
	ErrListaOrdemInvalida             = errors.New("listas.ordem_invalida")
	ErrListaSincronizacaoNaoPermitida = errors.New("listas.sincronizacao_nao_permitida")
	ErrListaIDInvalido                = errors.New("listas.id_invalido")
	ErrListaCampoNaoPermitido         = errors.New("listas.campo_nao_permitido")
	ErrListaDesafioSemJogos           = errors.New("listas.desafio_sem_jogos")
	ErrListaItensDemais               = errors.New("listas.itens_demais")
)

type ListasIGDBService interface {
	ObterFranquia(ctx context.Context, id int64) (*igdbclient.Franchise, error)
	JogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error)
	AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error)
	JogosDaFranquiaParaDesafio(ctx context.Context, id int64) ([]igdbclient.Game, error)
	AtualizarJogosDaFranquiaParaDesafio(ctx context.Context, id int64) ([]igdbclient.Game, error)
}

type RegraDetalhe struct {
	Tipo   string  `json:"tipo"`
	Valor  *string `json:"valor"`
	IgdbID *int32  `json:"igdb_id"`
}

type OrigemDetalhe struct {
	Tipo   string `json:"tipo"`
	IgdbID *int32 `json:"igdb_id"`
	Nome   string `json:"nome"`
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
	Ignorado      bool             `json:"-"`
	JogoZerado    *JogoZeradoMatch `json:"jogo_zerado"`
}

type ListaResumo struct {
	ID             int64             `json:"id"`
	Tipo           string            `json:"tipo"`
	Nome           string            `json:"nome"`
	Descricao      *string           `json:"descricao"`
	Regra          *RegraDetalhe     `json:"-"`
	Origem         *OrigemDetalhe    `json:"origem"`
	Meta           *int              `json:"-"`
	TotalItens     int               `json:"total_itens"`
	TotalIgnorados int               `json:"-"`
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

type PreviaFranquiaInfo struct {
	IgdbID int64  `json:"igdb_id"`
	Nome   string `json:"nome"`
}

type PreviaJogoItem struct {
	IgdbID        int32   `json:"igdb_id"`
	Nome          string  `json:"nome"`
	IgdbCapaURL   *string `json:"igdb_capa_url"`
	AnoLancamento *int    `json:"ano_lancamento"`
	Tipo          string  `json:"tipo"`
	Sugerido      bool    `json:"sugerido"`
	JaZerado      bool    `json:"ja_zerado"`
	JogoZeradoID  *int32  `json:"jogo_zerado_id"`
}

type PreviaDesafioResultado struct {
	Franquia       PreviaFranquiaInfo `json:"franquia"`
	Total          int                `json:"total"`
	TotalSugeridos int                `json:"total_sugeridos"`
	Jogos          []*PreviaJogoItem  `json:"jogos"`
}

type CriarListaRegraInput struct {
	Tipo             string  `json:"tipo"`
	Valor            string  `json:"valor"`
	IgdbID           *int32  `json:"igdb_id"`
	IgdbIDsIgnorados []int32 `json:"igdb_ids_ignorados,omitempty"`
}

type CriarListaOrigemInput struct {
	Tipo   string
	IgdbID int32
	Nome   string
}

type CriarListaItemInput struct {
	IgdbID        int32
	Nome          string
	IgdbCapaURL   *string
	AnoLancamento *int
}

type CriarListaInput struct {
	UsuarioID int32
	Tipo      string
	Nome      string
	Descricao *string
	Regra     *CriarListaRegraInput
	Meta      *int
	Origem    *CriarListaOrigemInput
	Itens     []CriarListaItemInput
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

type AdicionarItensLoteInput struct {
	ListaID   int64
	UsuarioID int32
	Itens     []CriarListaItemInput
}

type AdicionarItensLoteResultado struct {
	Adicionados  int `json:"adicionados"`
	JaExistentes int `json:"ja_existentes"`
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
	if input.Origem != nil || len(input.Itens) > 0 {
		return s.criarListaComJogos(ctx, input)
	}
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

		games, err := s.igdb.JogosDaFranquiaParaDesafio(ctx, int64(*input.Regra.IgdbID))
		if err != nil {
			return nil, err
		}

		filtered := filtrarJogosDesafioFranquia(int64(*input.Regra.IgdbID), games, time.Now())
		if len(filtered) == 0 {
			return nil, ErrListaFranquiaSemJogos
		}

		sort.SliceStable(filtered, func(i, j int) bool {
			dateA := int64(1<<62 - 1)
			if filtered[i].FirstReleaseDate != nil {
				dateA = *filtered[i].FirstReleaseDate
			}
			dateB := int64(1<<62 - 1)
			if filtered[j].FirstReleaseDate != nil {
				dateB = *filtered[j].FirstReleaseDate
			}
			if dateA != dateB {
				return dateA < dateB
			}
			return filtered[i].ID < filtered[j].ID
		})

		ignoradoSet := make(map[int32]bool)
		if input.Regra != nil {
			for _, id := range input.Regra.IgdbIDsIgnorados {
				ignoradoSet[id] = true
			}
		}

		seenIDs := make(map[int32]bool, len(filtered))
		itensParams := make([]repository.CriarItemParams, 0, len(filtered))
		pos := 1
		nonIgnoredCount := 0

		for _, g := range filtered {
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

			isIgnorado := ignoradoSet[gid]
			if !isIgnorado {
				nonIgnoredCount++
			}

			itensParams = append(itensParams, repository.CriarItemParams{
				IgdbID:        &gid,
				Nome:          g.Name,
				IgdbCapaURL:   capaURL,
				AnoLancamento: ano,
				Posicao:       pos,
				Ignorado:      isIgnorado,
			})
			pos++
		}

		if nonIgnoredCount == 0 {
			return nil, ErrListaFranquiaSemJogos
		}

		if input.Meta != nil {
			if *input.Meta < 1 || *input.Meta > nonIgnoredCount {
				return nil, ErrListaMetaInvalida
			}
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

func (s *ListasService) criarListaComJogos(ctx context.Context, input CriarListaInput) (*ListaDetalhada, error) {
	nome := strings.TrimSpace(input.Nome)
	if nome == "" {
		return nil, ErrListaNomeObrigatorio
	}
	if len([]rune(nome)) > 100 {
		return nil, ErrListaNomeInvalido
	}
	if input.Tipo != "fila" && input.Tipo != "desafio" {
		return nil, ErrListaTipoInvalido
	}
	if input.Meta != nil || input.Regra != nil {
		return nil, ErrListaCampoNaoPermitido
	}
	var descricao *string
	if input.Descricao != nil {
		valor := strings.TrimSpace(*input.Descricao)
		if len([]rune(valor)) > 200 {
			return nil, ErrListaDescricaoMuitoLonga
		}
		if valor != "" {
			descricao = &valor
		}
	}
	if input.Tipo == "desafio" {
		if input.Origem == nil || !origemValida(input.Origem) {
			return nil, ErrListaRegraInvalida
		}
		if len(input.Itens) == 0 {
			return nil, ErrListaDesafioSemJogos
		}
	}
	if len(input.Itens) > 1000 {
		return nil, ErrListaItensDemais
	}
	itens := make([]repository.CriarItemParams, 0, len(input.Itens))
	ids := make(map[int32]struct{}, len(input.Itens))
	for _, item := range input.Itens {
		if item.IgdbID <= 0 {
			return nil, ErrListaRegraInvalida
		}
		itemNome := strings.TrimSpace(item.Nome)
		if itemNome == "" {
			return nil, ErrListaNomeObrigatorio
		}
		if len([]rune(itemNome)) > 200 {
			return nil, ErrListaNomeInvalido
		}
		if _, found := ids[item.IgdbID]; found {
			continue
		}
		ids[item.IgdbID] = struct{}{}
		itens = append(itens, repository.CriarItemParams{IgdbID: &item.IgdbID, Nome: itemNome, IgdbCapaURL: item.IgdbCapaURL, AnoLancamento: item.AnoLancamento, Posicao: len(itens) + 1})
	}
	if input.Tipo == "desafio" && len(itens) == 0 {
		return nil, ErrListaDesafioSemJogos
	}
	var regraTipo, regraValor string
	var regraID *int32
	if input.Origem != nil {
		regraTipo, regraValor, regraID = input.Origem.Tipo, strings.TrimSpace(input.Origem.Nome), &input.Origem.IgdbID
	}
	params := repository.CriarListaComItensParams{
		Lista: repository.CriarListaParams{UsuarioID: input.UsuarioID, Tipo: input.Tipo, Nome: nome, Descricao: descricao, RegraTipo: optionalString(regraTipo), RegraValor: optionalString(regraValor), RegraIgdbID: regraID},
		Itens: itens,
	}
	if len(itens) == 0 {
		lista, err := s.repo.Criar(ctx, params.Lista)
		if err != nil {
			return nil, err
		}
		return s.ObterLista(ctx, lista.ID, input.UsuarioID)
	}
	lista, _, err := s.repo.CriarComItens(ctx, params)
	if err != nil {
		return nil, err
	}
	return s.ObterLista(ctx, lista.ID, input.UsuarioID)
}

func origemValida(origem *CriarListaOrigemInput) bool {
	if origem.IgdbID <= 0 || strings.TrimSpace(origem.Nome) == "" || len([]rune(strings.TrimSpace(origem.Nome))) > 150 {
		return false
	}
	return origem.Tipo == "franquia" || origem.Tipo == "plataforma" || origem.Tipo == "genero"
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
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

	dbItens, err := s.repo.ListarItensPorLista(ctx, id, usuarioID)
	if err != nil {
		return nil, err
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
		listaItens := itensPorLista[l.ID]
		resumo, _ := s.calcularLista(l, jogos, listaItens)
		res = append(res, resumo)
	}

	return res, nil
}

func (s *ListasService) AtualizarLista(ctx context.Context, input AtualizarListaInput) (*ListaDetalhada, error) {
	if input.Tipo != nil || input.Regra != nil || input.Meta != nil {
		return nil, ErrListaCampoNaoPermitido
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

	_, err = s.repo.Atualizar(ctx, repository.AtualizarListaParams{
		ID:        input.ID,
		UsuarioID: input.UsuarioID,
		Nome:      nome,
		Descricao: descPtr,
		Meta:      nil,
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
	lista, err := s.repo.BuscarPorID(ctx, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrListaNaoEncontrada
		}
		return err
	}
	if lista == nil {
		return ErrListaNaoEncontrada
	}

	item, err := s.repo.BuscarItemPorID(ctx, itemID, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrListaItemNaoEncontrado
		}
		return err
	}
	if item == nil {
		return ErrListaItemNaoEncontrado
	}

	err = s.repo.ExcluirItemERecompactar(ctx, itemID, listaID, usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrListaItemNaoEncontrado
		}
		return err
	}
	return nil
}

func (s *ListasService) AdicionarItensLote(ctx context.Context, input AdicionarItensLoteInput) (*AdicionarItensLoteResultado, error) {
	if len(input.Itens) < 1 || len(input.Itens) > 500 {
		return nil, ErrListaItensDemais
	}
	params := make([]repository.CriarItemParams, 0, len(input.Itens))
	ids := make(map[int32]struct{}, len(input.Itens))
	for _, item := range input.Itens {
		if item.IgdbID <= 0 {
			return nil, ErrListaRegraInvalida
		}
		nome := strings.TrimSpace(item.Nome)
		if nome == "" {
			return nil, ErrListaNomeObrigatorio
		}
		if len([]rune(nome)) > 200 {
			return nil, ErrListaNomeInvalido
		}
		if _, found := ids[item.IgdbID]; found {
			continue
		}
		ids[item.IgdbID] = struct{}{}
		params = append(params, repository.CriarItemParams{IgdbID: &item.IgdbID, Nome: nome, IgdbCapaURL: item.IgdbCapaURL, AnoLancamento: item.AnoLancamento})
	}
	resultado, err := s.repo.AdicionarItensLote(ctx, input.ListaID, input.UsuarioID, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrListaNaoEncontrada
		}
		return nil, err
	}
	return &AdicionarItensLoteResultado{Adicionados: resultado.Adicionados, JaExistentes: resultado.JaExistentes + len(input.Itens) - len(params)}, nil
}

func (s *ListasService) RestaurarItem(ctx context.Context, itemID int64, listaID int64, usuarioID int32) (*ListaItemDetalhe, error) {
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

	if lista.Tipo != "desafio" || lista.RegraTipo == nil || *lista.RegraTipo != "franquia" {
		return nil, ErrListaRestauracaoNaoPermitida
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

	updatedItem := item
	if legacyRepo, ok := s.repo.(interface {
		DefinirIgnoradoItem(context.Context, int64, int64, int32, bool) (*repository.ListaItem, error)
	}); ok {
		var err error
		updatedItem, err = legacyRepo.DefinirIgnoradoItem(ctx, itemID, listaID, usuarioID, false)
		if err != nil {
			return nil, err
		}
	}

	jogos, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	return s.construirItemDetalhe(updatedItem, jogos), nil
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

	games, err := s.igdb.AtualizarJogosDaFranquiaParaDesafio(ctx, int64(*lista.RegraIgdbID))
	if err != nil {
		return nil, err
	}

	filtered := filtrarJogosDesafioFranquia(int64(*lista.RegraIgdbID), games, time.Now())

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
	for _, g := range filtered {
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

	seenIDs := make(map[int32]bool, len(novosJogos))
	novosParams := make([]repository.CriarItemParams, 0, len(novosJogos))
	for _, g := range novosJogos {
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
		novosParams = append(novosParams, repository.CriarItemParams{
			IgdbID:        &gid,
			Nome:          g.Name,
			IgdbCapaURL:   capaURL,
			AnoLancamento: ano,
			Ignorado:      false,
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

func (s *ListasService) PreviaDesafioFranquia(ctx context.Context, franquiaID int64, usuarioID int32) (*PreviaDesafioResultado, error) {
	if franquiaID <= 0 {
		return nil, ErrListaIDInvalido
	}

	franchise, err := s.igdb.ObterFranquia(ctx, franquiaID)
	if err != nil {
		if errors.Is(err, ErrJogoIGDBNaoEncontrado) {
			return nil, ErrListaFranquiaNaoEncontrada
		}
		return nil, err
	}
	if franchise == nil {
		return nil, ErrListaFranquiaNaoEncontrada
	}

	games, err := s.igdb.JogosDaFranquiaParaDesafio(ctx, franquiaID)
	if err != nil {
		return nil, err
	}

	filtered := filtrarJogosDesafioFranquia(franquiaID, games, time.Now())
	if len(filtered) == 0 {
		return nil, ErrListaFranquiaSemJogos
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		dateA := int64(1<<62 - 1)
		if filtered[i].FirstReleaseDate != nil {
			dateA = *filtered[i].FirstReleaseDate
		}
		dateB := int64(1<<62 - 1)
		if filtered[j].FirstReleaseDate != nil {
			dateB = *filtered[j].FirstReleaseDate
		}
		if dateA != dateB {
			return dateA < dateB
		}
		return filtered[i].ID < filtered[j].ID
	})

	jogosZerados, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	matchByIgdbID := make(map[int32]int32)
	for _, j := range jogosZerados {
		if j.IgdbID != nil {
			if _, exists := matchByIgdbID[*j.IgdbID]; !exists {
				matchByIgdbID[*j.IgdbID] = j.ID
			}
		}
	}

	seenIDs := make(map[int32]bool, len(filtered))
	previaItens := make([]*PreviaJogoItem, 0, len(filtered))
	totalSugeridos := 0
	for _, g := range filtered {
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

		var jogoZeradoID *int32
		jaZerado := false
		if zid, ok := matchByIgdbID[gid]; ok {
			jaZerado = true
			jogoZeradoID = &zid
		}

		sugerido := jogoSugerido(g, franquiaID, franchise.Name)
		if sugerido {
			totalSugeridos++
		}

		previaItens = append(previaItens, &PreviaJogoItem{
			IgdbID:        gid,
			Nome:          g.Name,
			IgdbCapaURL:   capaURL,
			AnoLancamento: ano,
			Tipo:          tipoJogo(g.GameType),
			Sugerido:      sugerido,
			JaZerado:      jaZerado,
			JogoZeradoID:  jogoZeradoID,
		})
	}

	return &PreviaDesafioResultado{
		Franquia: PreviaFranquiaInfo{
			IgdbID: franchise.ID,
			Nome:   franchise.Name,
		},
		Total:          len(previaItens),
		TotalSugeridos: totalSugeridos,
		Jogos:          previaItens,
	}, nil
}

func (s *ListasService) calcularLista(lista *repository.Lista, jogos []*repository.JogoZeradoResumo, itens []*repository.ListaItem) (*ListaResumo, []*ListaItemDetalhe) {
	var regra *RegraDetalhe
	var origem *OrigemDetalhe
	if lista.RegraTipo != nil {
		regra = &RegraDetalhe{
			Tipo:   *lista.RegraTipo,
			Valor:  lista.RegraValor,
			IgdbID: lista.RegraIgdbID,
		}
		origem = &OrigemDetalhe{Tipo: *lista.RegraTipo, IgdbID: lista.RegraIgdbID}
		if lista.RegraValor != nil {
			origem.Nome = *lista.RegraValor
		}
	}

	if lista.Tipo == "desafio" && len(itens) == 0 && lista.RegraTipo != nil && (*lista.RegraTipo == "plataforma" || *lista.RegraTipo == "genero") {
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
			Origem:         origem,
			Meta:           lista.Meta,
			TotalItens:     len(matchedJogos),
			TotalIgnorados: 0,
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
	totalItens := 0
	totalIgnorados := 0

	for _, it := range itens {
		d := s.construirItemDetalhe(it, jogos)
		detalhesItens = append(detalhesItens, d)
		if it.Ignorado {
			totalIgnorados++
			continue
		}
		totalItens++
		if d.Zerado {
			feitos++
			if d.JogoZerado != nil {
				matchedFinalizados = append(matchedFinalizados, d.JogoZerado.FinalizadoEm)
			}
		}
	}
	if lista.Tipo == "desafio" {
		ordenarItensDesafio(detalhesItens)
	}

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
		Origem:         origem,
		Meta:           lista.Meta,
		TotalItens:     totalItens,
		TotalIgnorados: totalIgnorados,
		ItensPendentes: itensPendentes,
		Progresso:      progresso,
		CreatedAt:      lista.CreatedAt,
		UpdatedAt:      lista.UpdatedAt,
	}

	return resumo, detalhesItens
}

func ordenarItensDesafio(itens []*ListaItemDetalhe) {
	sort.SliceStable(itens, func(i, j int) bool {
		itemA, itemB := itens[i], itens[j]
		if itemA.AnoLancamento == nil && itemB.AnoLancamento != nil {
			return false
		}
		if itemA.AnoLancamento != nil && itemB.AnoLancamento == nil {
			return true
		}
		if itemA.AnoLancamento != nil && itemB.AnoLancamento != nil && *itemA.AnoLancamento != *itemB.AnoLancamento {
			return *itemA.AnoLancamento < *itemB.AnoLancamento
		}
		nomeA := normalizeLowerUnaccent(itemA.Nome)
		nomeB := normalizeLowerUnaccent(itemB.Nome)
		if nomeA != nomeB {
			return nomeA < nomeB
		}
		return itemA.ID < itemB.ID
	})
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
		Ignorado:      item.Ignorado,
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
