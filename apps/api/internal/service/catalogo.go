package service

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

var (
	ErrCatalogoParametroInvalido = errors.New("catalogo.parametro_invalido")
	ErrCatalogoPaginacaoInvalida = errors.New("catalogo.paginacao_invalida")
)

type CatalogoClient interface {
	FranchiseDetails(context.Context, int64) (*igdbclient.Franchise, error)
	GamesByFranchiseParaDesafio(context.Context, int64) ([]igdbclient.Game, error)
	CatalogGames(context.Context, string, string, int, int) ([]igdbclient.Game, error)
	CountGames(context.Context, string) (int, error)
	Genres(context.Context) ([]igdbclient.Genre, error)
}

type CatalogoOrigem string

const (
	CatalogoFranquia   CatalogoOrigem = "franquia"
	CatalogoPlataforma CatalogoOrigem = "plataforma"
	CatalogoGenero     CatalogoOrigem = "genero"
)

type CatalogoFiltro struct {
	Origem           CatalogoOrigem
	ID               int64
	GeneroID         *int64
	PlataformaID     *int64
	SomenteSugeridos bool
	Busca            string
	Ordenar          string
	Pagina           int
	PorPagina        int
	Agora            time.Time
}

type CatalogoItem struct {
	IGDBID        int32   `json:"igdb_id"`
	Nome          string  `json:"nome"`
	IGDBCapaURL   *string `json:"igdb_capa_url"`
	AnoLancamento *int    `json:"ano_lancamento"`
	Sugerido      bool    `json:"sugerido"`
	JaZerado      bool    `json:"ja_zerado"`
	JogoZeradoID  *int32  `json:"jogo_zerado_id"`
}

type CatalogoMeta struct {
	Pagina         int  `json:"pagina"`
	PorPagina      int  `json:"por_pagina"`
	Total          int  `json:"total"`
	TotalSugeridos *int `json:"total_sugeridos"`
	TotalTodos     *int `json:"total_todos"`
}

type CatalogoResultado struct {
	Itens []*CatalogoItem `json:"itens"`
	Meta  CatalogoMeta    `json:"meta"`
}

type GeneroCatalogo struct {
	ID   int64  `json:"id"`
	Nome string `json:"nome"`
}

type CatalogoService struct {
	client CatalogoClient
	cache  IGDBSnapshotRepository
	repo   repository.ListasRepository
}

func NewCatalogoService(client CatalogoClient, cache IGDBSnapshotRepository, repo repository.ListasRepository) *CatalogoService {
	return &CatalogoService{client: client, cache: cache, repo: repo}
}

func (s *CatalogoService) Jogos(ctx context.Context, usuarioID int32, filtro CatalogoFiltro) (*CatalogoResultado, error) {
	if filtro.Agora.IsZero() {
		filtro.Agora = time.Now()
	}
	if err := validarCatalogoFiltro(filtro); err != nil {
		return nil, err
	}
	if filtro.Origem == CatalogoFranquia {
		return s.jogosFranquia(ctx, usuarioID, filtro)
	}
	where := montarWhereCatalogo(filtro)
	games, err := s.client.CatalogGames(ctx, where, ordenarIGDB(filtro.Ordenar), filtro.PorPagina, (filtro.Pagina-1)*filtro.PorPagina)
	if err != nil {
		return nil, traduzErroIGDB(err)
	}
	total, err := s.client.CountGames(ctx, where)
	if err != nil {
		return nil, traduzErroIGDB(err)
	}
	return s.montarResultado(ctx, usuarioID, games, CatalogoMeta{Pagina: filtro.Pagina, PorPagina: filtro.PorPagina, Total: total})
}

func (s *CatalogoService) jogosFranquia(ctx context.Context, usuarioID int32, filtro CatalogoFiltro) (*CatalogoResultado, error) {
	franquia, err := s.client.FranchiseDetails(ctx, filtro.ID)
	if err != nil {
		return nil, traduzErroIGDB(err)
	}
	if franquia == nil {
		return nil, ErrListaFranquiaNaoEncontrada
	}
	var games []igdbclient.Game
	key := chaveSnapshotFranquiaDesafio(filtro.ID)
	found, err := s.cache.Buscar(ctx, key, &games)
	if err != nil {
		return nil, err
	}
	if !found {
		games, err = s.client.GamesByFranchiseParaDesafio(ctx, filtro.ID)
		if err != nil {
			return nil, traduzErroIGDB(err)
		}
		if err := s.cache.Salvar(ctx, key, games); err != nil {
			return nil, err
		}
	}
	games = filtrarJogosDesafioFranquia(filtro.ID, games, filtro.Agora)
	filtered := make([]igdbclient.Game, 0, len(games))
	for _, game := range games {
		if filtro.Busca == "" || strings.Contains(normalizeString(game.Name), normalizeString(filtro.Busca)) {
			filtered = append(filtered, game)
		}
	}
	totalTodos := len(filtered)
	totalSugeridos := 0
	for _, game := range filtered {
		if jogoSugerido(game, filtro.ID, franquia.Name) {
			totalSugeridos++
		}
	}
	if filtro.SomenteSugeridos {
		sugeridos := make([]igdbclient.Game, 0, totalSugeridos)
		for _, game := range filtered {
			if jogoSugerido(game, filtro.ID, franquia.Name) {
				sugeridos = append(sugeridos, game)
			}
		}
		filtered = sugeridos
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		return compararJogosCatalogo(filtered[i], filtered[j], filtro.Ordenar)
	})
	total := len(filtered)
	start := (filtro.Pagina - 1) * filtro.PorPagina
	if start > total {
		start = total
	}
	end := start + filtro.PorPagina
	if end > total {
		end = total
	}
	meta := CatalogoMeta{
		Pagina:         filtro.Pagina,
		PorPagina:      filtro.PorPagina,
		Total:          total,
		TotalSugeridos: &totalSugeridos,
		TotalTodos:     &totalTodos,
	}
	resultado, err := s.montarResultado(ctx, usuarioID, filtered[start:end], meta)
	if err != nil {
		return nil, err
	}
	for index, item := range resultado.Itens {
		item.Sugerido = jogoSugerido(filtered[start+index], filtro.ID, franquia.Name)
	}
	return resultado, nil
}

func (s *CatalogoService) Generos(ctx context.Context) ([]GeneroCatalogo, error) {
	var genres []igdbclient.Genre
	found, err := s.cache.Buscar(ctx, "genres", &genres)
	if err != nil {
		return nil, err
	}
	if !found {
		genres, err = s.client.Genres(ctx)
		if err != nil {
			return nil, traduzErroIGDB(err)
		}
		if err := s.cache.Salvar(ctx, "genres", genres); err != nil {
			return nil, err
		}
	}
	result := make([]GeneroCatalogo, 0, len(genres))
	for _, genre := range genres {
		result = append(result, GeneroCatalogo{ID: genre.ID, Nome: nomeGenero(genre)})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Nome < result[j].Nome })
	return result, nil
}

func (s *CatalogoService) montarResultado(ctx context.Context, usuarioID int32, games []igdbclient.Game, meta CatalogoMeta) (*CatalogoResultado, error) {
	jogosZerados, err := s.repo.ListarJogosZeradosUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	matching := make(map[int32]int32, len(jogosZerados))
	for _, jogo := range jogosZerados {
		if jogo.IgdbID != nil {
			matching[*jogo.IgdbID] = jogo.ID
		}
	}
	itens := make([]*CatalogoItem, 0, len(games))
	for _, game := range games {
		id := int32(game.ID)
		item := &CatalogoItem{IGDBID: id, Nome: game.Name, Sugerido: false}
		if game.Cover != nil {
			game.Cover.EnsureURL()
			if game.Cover.URL != "" {
				item.IGDBCapaURL = &game.Cover.URL
			}
		}
		if game.FirstReleaseDate != nil {
			year := time.Unix(*game.FirstReleaseDate, 0).UTC().Year()
			item.AnoLancamento = &year
		}
		if zeradoID, found := matching[id]; found {
			item.JaZerado = true
			item.JogoZeradoID = &zeradoID
		}
		itens = append(itens, item)
	}
	return &CatalogoResultado{Itens: itens, Meta: meta}, nil
}

func validarCatalogoFiltro(filtro CatalogoFiltro) error {
	if filtro.ID <= 0 || (filtro.Origem != CatalogoFranquia && filtro.Origem != CatalogoPlataforma && filtro.Origem != CatalogoGenero) {
		return ErrCatalogoParametroInvalido
	}
	if filtro.Origem != CatalogoPlataforma && filtro.GeneroID != nil || filtro.Origem != CatalogoGenero && filtro.PlataformaID != nil {
		return ErrCatalogoParametroInvalido
	}
	if filtro.SomenteSugeridos && filtro.Origem != CatalogoFranquia {
		return ErrCatalogoParametroInvalido
	}
	if filtro.Ordenar != "populares" && filtro.Ordenar != "lancamento" && filtro.Ordenar != "nome" {
		return ErrCatalogoParametroInvalido
	}
	if filtro.Pagina < 1 || filtro.PorPagina < 1 || filtro.PorPagina > 60 {
		return ErrCatalogoPaginacaoInvalida
	}
	if len([]rune(filtro.Busca)) > 100 {
		return ErrCatalogoParametroInvalido
	}
	return nil
}

func montarWhereCatalogo(filtro CatalogoFiltro) string {
	parts := []string{"game_type = (0, 4, 8, 9, 10)", "version_parent = null", "first_release_date != null", "first_release_date <= " + formatInt(filtro.Agora.Unix())}
	if filtro.Origem == CatalogoPlataforma {
		parts = append(parts, "platforms = ("+formatInt(filtro.ID)+")")
		if filtro.GeneroID != nil {
			parts = append(parts, "genres = ("+formatInt(*filtro.GeneroID)+")")
		}
	}
	if filtro.Origem == CatalogoGenero {
		parts = append(parts, "genres = ("+formatInt(filtro.ID)+")")
		if filtro.PlataformaID != nil {
			parts = append(parts, "platforms = ("+formatInt(*filtro.PlataformaID)+")")
		}
	}
	if strings.TrimSpace(filtro.Busca) != "" {
		parts = append(parts, `name ~ *"`+escapeCatalogTerm(filtro.Busca)+`"*`)
	}
	return strings.Join(parts, " & ")
}

func escapeCatalogTerm(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.NewReplacer(";", " ", "*", " ").Replace(value)
	return strings.Join(strings.Fields(value), " ")
}

func ordenarIGDB(ordenar string) string {
	if ordenar == "lancamento" {
		return "first_release_date asc"
	}
	if ordenar == "nome" {
		return "name asc"
	}
	return "total_rating_count desc"
}

func releaseDate(game igdbclient.Game) int64 {
	if game.FirstReleaseDate == nil {
		return 1<<62 - 1
	}
	return *game.FirstReleaseDate
}

func compararJogosCatalogo(left, right igdbclient.Game, ordenar string) bool {
	if ordenar == "lancamento" {
		return releaseDate(left) < releaseDate(right)
	}
	if ordenar == "nome" {
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	}
	leftRating, rightRating := ratingCount(left), ratingCount(right)
	if leftRating != rightRating {
		return leftRating > rightRating
	}
	return strings.ToLower(left.Name) < strings.ToLower(right.Name)
}

func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

var generosPTBR = map[string]string{
	"action": "Ação", "adventure": "Aventura", "arcade": "Arcade", "board-games": "Jogos de tabuleiro", "card-and-board-game": "Cartas e tabuleiro", "fighting": "Luta", "hack-and-slash-beat-em-up": "Beat 'em up", "indie": "Indie", "music": "Música", "pinball": "Pinball", "platform": "Plataforma", "point-and-click": "Point-and-click", "puzzle": "Quebra-cabeça", "quiz-trivia": "Quiz e trivia", "racing": "Corrida", "real-time-strategy-rts": "Estratégia em tempo real", "role-playing-rpg": "RPG", "shooter": "Tiro", "simulator": "Simulação", "sport": "Esporte", "strategy": "Estratégia", "tactical": "Tática", "turn-based-strategy-tbs": "Estratégia por turnos", "visual-novel": "Novela visual",
}

func nomeGenero(genre igdbclient.Genre) string {
	if nome, found := generosPTBR[genre.Slug]; found {
		return nome
	}
	return genre.Name
}
