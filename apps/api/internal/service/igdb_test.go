package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
)

type igdbMock struct {
	platformGamesCalls  int
	franchiseGamesCalls int
	detail              *igdbclient.Game
	detailErr           error
}

func (m *igdbMock) SearchGames(context.Context, string) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *igdbMock) GameDetails(context.Context, int64) (*igdbclient.Game, error) {
	return m.detail, m.detailErr
}

func (m *igdbMock) Platforms(context.Context) ([]igdbclient.Platform, error) {
	return []igdbclient.Platform{}, nil
}

func (m *igdbMock) GamesByPlatform(context.Context, int64) ([]igdbclient.Game, error) {
	m.platformGamesCalls++
	return []igdbclient.Game{{ID: 7, Name: "Cached"}}, nil
}

func (m *igdbMock) AtualizarJogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return m.GamesByPlatform(ctx, id)
}

func (m *igdbMock) SearchFranchises(context.Context, string) ([]igdbclient.Franchise, error) {
	return []igdbclient.Franchise{}, nil
}

func (m *igdbMock) FranchiseDetails(context.Context, int64) (*igdbclient.Franchise, error) {
	return &igdbclient.Franchise{ID: 1, Name: "Test Franchise"}, nil
}

func (m *igdbMock) GamesByFranchise(context.Context, int64) ([]igdbclient.Game, error) {
	m.franchiseGamesCalls++
	return []igdbclient.Game{{ID: 8, Name: "Franchise"}}, nil
}

func (m *igdbMock) GamesByFranchiseParaDesafio(context.Context, int64) ([]igdbclient.Game, error) {
	m.franchiseGamesCalls++
	return []igdbclient.Game{{ID: 8, Name: "Franchise"}}, nil
}

func (m *igdbMock) AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return m.GamesByFranchise(ctx, id)
}

type cacheMock struct {
	items map[string]any
}

func (m *cacheMock) Buscar(_ context.Context, key string, destino any) (bool, error) {
	item, found := m.items[key]
	if found {
		games := destino.(*[]igdbclient.Game)
		*games = item.([]igdbclient.Game)
	}
	return found, nil
}

func (m *cacheMock) Salvar(_ context.Context, key string, valor any) error {
	m.items[key] = valor
	return nil
}

func TestIGDBServiceSnapshotsPlatformCatalog(t *testing.T) {
	client := &igdbMock{}
	cache := &cacheMock{items: make(map[string]any)}
	service := NewIGDBService(client, cache)
	for range 2 {
		games, err := service.JogosDaPlataforma(context.Background(), 4)
		if err != nil || len(games) != 1 || games[0].Name != "Cached" {
			t.Fatalf("games=%v err=%v", games, err)
		}
	}
	if client.platformGamesCalls != 1 {
		t.Fatalf("IGDB calls=%d", client.platformGamesCalls)
	}
}

func TestIGDBServiceManualRefreshReplacesSnapshot(t *testing.T) {
	client := &igdbMock{}
	cache := &cacheMock{items: make(map[string]any)}
	service := NewIGDBService(client, cache)
	if _, err := service.JogosDaPlataforma(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AtualizarJogosDaPlataforma(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	if client.platformGamesCalls != 2 {
		t.Fatalf("IGDB calls=%d", client.platformGamesCalls)
	}
}

func TestIGDBServiceRejectsEmptySearch(t *testing.T) {
	service := NewIGDBService(&igdbMock{}, &cacheMock{items: make(map[string]any)})
	if _, err := service.BuscarJogos(context.Background(), "  "); !errors.Is(err, ErrTermoIGDBVazio) {
		t.Fatalf("error=%v", err)
	}
	if _, err := service.BuscarFranquias(context.Background(), ""); !errors.Is(err, ErrTermoIGDBVazio) {
		t.Fatalf("error=%v", err)
	}
}

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"acentos", "Pokémon: Let's Go", "pokemon let s go"},
		{"pontuação e espaços", "  Chrono   Trigger+  ", "chrono trigger"},
		{"caixa alta e caracteres especiais", "São Paulo - 2024!", "sao paulo 2024"},
		{"simbolos diversos", "Crash Bandicoot: Warped! (1998)", "crash bandicoot warped 1998"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeString(tc.input); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

type customSearchMock struct {
	igdbMock
	searchFn  func(ctx context.Context, query string) ([]igdbclient.Game, error)
	lastQuery string
}

func (m *customSearchMock) SearchGames(ctx context.Context, q string) ([]igdbclient.Game, error) {
	m.lastQuery = q
	if m.searchFn != nil {
		return m.searchFn(ctx, q)
	}
	return nil, nil
}

func TestBuscarJogos_Filtros(t *testing.T) {
	tests := []struct {
		name       string
		gameType   int
		shouldKeep bool
	}{
		{"main game", igdbclient.GameTypeMainGame, true},
		{"standalone expansion", igdbclient.GameTypeStandaloneExpansion, true},
		{"remake", igdbclient.GameTypeRemake, true},
		{"remaster", igdbclient.GameTypeRemaster, true},
		{"expanded game", igdbclient.GameTypeExpandedGame, true},
		{"port", igdbclient.GameTypePort, true},
		{"fork", igdbclient.GameTypeFork, true},
		{"dlc", igdbclient.GameTypeDLC, false},
		{"expansion", igdbclient.GameTypeExpansion, false},
		{"bundle", igdbclient.GameTypeBundle, false},
		{"mod", igdbclient.GameTypeMod, false},
		{"episode", igdbclient.GameTypeEpisode, false},
		{"season", igdbclient.GameTypeSeason, false},
		{"pack", igdbclient.GameTypePack, false},
		{"update", igdbclient.GameTypeUpdate, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &customSearchMock{
				searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
					return []igdbclient.Game{{ID: 10, Name: "Test Game", GameType: tc.gameType}}, nil
				},
			}
			svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
			games, err := svc.BuscarJogos(context.Background(), "test")
			if err != nil {
				t.Fatal(err)
			}
			if tc.shouldKeep && len(games) != 1 {
				t.Fatalf("esperava manter jogo do tipo %d, mas foi descartado", tc.gameType)
			}
			if !tc.shouldKeep && len(games) != 0 {
				t.Fatalf("esperava descartar jogo do tipo %d, mas foi retornado", tc.gameType)
			}
		})
	}
}

func TestBuscarJogos_VersoesChronoTriggerDatas(t *testing.T) {
	dateSNES := int64(794880000)
	datePS1 := int64(943920000)
	dateDS := int64(1227139200)
	dateSteam := int64(1519689600)
	mock := &customSearchMock{
		searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{
					ID:               1802,
					Name:             "Chrono Trigger",
					GameType:         igdbclient.GameTypeMainGame,
					FirstReleaseDate: &dateSNES,
					Platforms:        []igdbclient.Platform{{ID: 19, Name: "Super Nintendo Entertainment System"}},
				},
				{
					ID:               263446,
					Name:             "Chrono Trigger",
					GameType:         igdbclient.GameTypeExpandedGame,
					FirstReleaseDate: &datePS1,
					Platforms:        []igdbclient.Platform{{ID: 7, Name: "PlayStation"}},
				},
				{
					ID:               20398,
					Name:             "Chrono Trigger",
					GameType:         igdbclient.GameTypePort,
					FirstReleaseDate: &dateDS,
					Platforms:        []igdbclient.Platform{{ID: 20, Name: "Nintendo DS"}},
				},
				{
					ID:               206320,
					Name:             "Chrono Trigger",
					GameType:         igdbclient.GameTypePort,
					FirstReleaseDate: &dateSteam,
					Platforms:        []igdbclient.Platform{{ID: 6, Name: "PC (Microsoft Windows)"}},
				},
			}, nil
		},
	}

	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	games, err := svc.BuscarJogos(context.Background(), "Chrono Trigger")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 4 {
		t.Fatalf("esperava 4 versões para Chrono Trigger, obteve %d", len(games))
	}
	g := games[0]
	if g.ID != 1802 {
		t.Fatalf("id esperado 1802, obteve %d", g.ID)
	}
	if len(g.Platforms) != 1 {
		t.Fatalf("esperava plataforma própria, obteve %d: %+v", len(g.Platforms), g.Platforms)
	}
	hasSNES := false
	for _, p := range g.Platforms {
		if p.Name == "Super Nintendo" {
			hasSNES = true
		}
	}
	if !hasSNES {
		t.Fatalf("plataforma esperada não encontrada: %+v", g.Platforms)
	}
}

func TestBuscarJogos_VersaoSemPai(t *testing.T) {
	datePort := int64(1100000000)

	mock := &customSearchMock{
		searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{
					ID:               501,
					Name:             "Persona 3",
					GameType:         igdbclient.GameTypePort,
					FirstReleaseDate: &datePort,
					Platforms:        []igdbclient.Platform{{ID: 38, Name: "PlayStation Portable"}},
				},
			}, nil
		},
	}

	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	games, err := svc.BuscarJogos(context.Background(), "Persona 3")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 {
		t.Fatalf("esperava 1 jogo, obteve %d", len(games))
	}
	g := games[0]
	if g.ID != 501 || g.Name != "Persona 3" {
		t.Fatalf("dados incorretos: %+v", g)
	}
	if len(g.Platforms) != 1 {
		t.Fatalf("esperava plataforma própria, obteve %d: %+v", len(g.Platforms), g.Platforms)
	}
}

func TestBuscarJogos_RemakePermaneceSeparado(t *testing.T) {
	dateOriginal := int64(1030579200)
	dateRemake := int64(1600992000)
	mock := &customSearchMock{
		searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{
					ID:               39,
					Name:             "Mafia",
					GameType:         igdbclient.GameTypeMainGame,
					FirstReleaseDate: &dateOriginal,
					Platforms:        []igdbclient.Platform{{ID: 6, Name: "PC"}},
				},
				{
					ID:               134070,
					Name:             "Mafia: Definitive Edition",
					GameType:         igdbclient.GameTypeRemake,
					FirstReleaseDate: &dateRemake,
					Platforms:        []igdbclient.Platform{{ID: 48, Name: "PlayStation 4"}},
				},
				{
					ID:       392531,
					Name:     "Mafia: Definitive Edition - Chicago Outfit Pack",
					GameType: igdbclient.GameTypePack,
				},
			}, nil
		},
	}

	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	games, err := svc.BuscarJogos(context.Background(), "Mafia")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 2 {
		t.Fatalf("esperava 2 jogos (Mafia original e Remake), sem DLC, obteve %d", len(games))
	}
	ids := map[int64]bool{games[0].ID: true, games[1].ID: true}
	if !ids[39] || !ids[134070] {
		t.Fatalf("esperava IDs 39 e 134070, obteve %+v", games)
	}
}

func TestBuscarJogos_RankingETiers(t *testing.T) {
	ratingHigh := 1000
	ratingLow := 100
	dateOld := int64(1000000000)
	dateNew := int64(1500000000)

	mock := &customSearchMock{
		searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{ID: 4, Name: "Chrono Ressurection", TotalRatingCount: &ratingHigh, FirstReleaseDate: &dateNew},
				{ID: 3, Name: "Super Chrono Trigger World", TotalRatingCount: &ratingHigh, FirstReleaseDate: &dateNew},
				{ID: 2, Name: "Chrono Trigger: Jet Bike Special", TotalRatingCount: &ratingHigh, FirstReleaseDate: &dateNew},
				{ID: 1, Name: "Chrono Trigger", TotalRatingCount: &ratingLow, FirstReleaseDate: &dateOld},
			}, nil
		},
	}

	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	games, err := svc.BuscarJogos(context.Background(), "Chrono Trigger")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 4 {
		t.Fatalf("esperava 4 jogos, obteve %d", len(games))
	}
	if games[0].ID != 1 {
		t.Fatalf("tier 1 (exato) esperado no topo: obteve id %d (%s)", games[0].ID, games[0].Name)
	}
	if games[1].ID != 2 {
		t.Fatalf("tier 2 (prefixo) esperado em 2o: obteve id %d (%s)", games[1].ID, games[1].Name)
	}
	if games[2].ID != 3 {
		t.Fatalf("tier 3 (todos tokens) esperado em 3o: obteve id %d (%s)", games[2].ID, games[2].Name)
	}
	if games[3].ID != 4 {
		t.Fatalf("tier 4 (resto) esperado em 4o: obteve id %d (%s)", games[3].ID, games[3].Name)
	}
}

func TestBuscarJogos_DesempateRatingEData(t *testing.T) {
	rating500 := 500
	rating100 := 100
	date2000 := int64(946684800)
	date2010 := int64(1262304000)

	t.Run("desempate por rating desc", func(t *testing.T) {
		mock := &customSearchMock{
			searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
				return []igdbclient.Game{
					{ID: 1, Name: "Game", TotalRatingCount: &rating100},
					{ID: 2, Name: "Game", TotalRatingCount: &rating500},
				}, nil
			},
		}
		svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
		games, err := svc.BuscarJogos(context.Background(), "Game")
		if err != nil || len(games) != 2 || games[0].ID != 2 {
			t.Fatalf("esperava ID 2 com maior rating em 1o lugar: %+v", games)
		}
	})

	t.Run("desempate por data asc quando rating empata", func(t *testing.T) {
		mock := &customSearchMock{
			searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
				return []igdbclient.Game{
					{ID: 1, Name: "Game", TotalRatingCount: &rating500, FirstReleaseDate: &date2010},
					{ID: 2, Name: "Game", TotalRatingCount: &rating500, FirstReleaseDate: &date2000},
				}, nil
			},
		}
		svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
		games, err := svc.BuscarJogos(context.Background(), "Game")
		if err != nil || len(games) != 2 || games[0].ID != 2 {
			t.Fatalf("esperava ID 2 com data mais antiga em 1o lugar: %+v", games)
		}
	})
}

func TestBuscarJogos_Limite20(t *testing.T) {
	mock := &customSearchMock{
		searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
			games := make([]igdbclient.Game, 20)
			for i := range 20 {
				games[i] = igdbclient.Game{ID: int64(i + 1), Name: fmt.Sprintf("Game %d", i+1)}
			}
			return games, nil
		},
	}
	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	games, err := svc.BuscarJogos(context.Background(), "Game")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 20 {
		t.Fatalf("esperava exatamente 20 resultados no limite, obteve %d", len(games))
	}
}

func TestBuscarJogos_ErrosTraduzidos(t *testing.T) {
	tests := []struct {
		name      string
		clientErr error
		targetErr error
	}{
		{"rate limit", igdbclient.ErrRateLimited, ErrIGDBRateLimit},
		{"unavailable", igdbclient.ErrUnavailable, ErrIGDBIndisponivel},
		{"query invalid", igdbclient.ErrQueryInvalid, ErrIGDBQueryInvalida},
		{"authentication", igdbclient.ErrAuthentication, ErrIGDBIndisponivel},
		{"canceled", context.Canceled, context.Canceled},
		{"deadline exceeded", context.DeadlineExceeded, ErrIGDBIndisponivel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &customSearchMock{
				searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
					return nil, tc.clientErr
				},
			}
			svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
			_, err := svc.BuscarJogos(context.Background(), "term")
			if !errors.Is(err, tc.targetErr) {
				t.Fatalf("esperava erro %v, obteve %v", tc.targetErr, err)
			}
		})
	}
}

func TestBuscarJogos_SanitizaTermoECurto(t *testing.T) {
	mock := &customSearchMock{}
	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	_, err := svc.BuscarJogos(context.Background(), ` zelda\"*; `)
	if err != nil {
		t.Fatal(err)
	}
	if mock.lastQuery != "zelda" {
		t.Fatalf("termo sanitizado inesperado: %q", mock.lastQuery)
	}
	_, err = svc.BuscarJogos(context.Background(), ` a* `)
	if !errors.Is(err, ErrTermoIGDBCurto) {
		t.Fatalf("esperava termo curto, obteve %v", err)
	}
}

func TestBuscarJogos_QueryRepassadaAoClient(t *testing.T) {
	mock := &customSearchMock{}
	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	_, _ = svc.BuscarJogos(context.Background(), "  Super Mario  ")
	if mock.lastQuery != "Super Mario" {
		t.Fatalf("esperava consulta 'Super Mario', obteve %q", mock.lastQuery)
	}
}

func TestBuscarJogos_BuscaParcialEDeduplicacao(t *testing.T) {
	mock := &customSearchMock{
		searchFn: func(context.Context, string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{ID: 1, Name: "Chrono Trigger"},
				{ID: 1, Name: "Chrono Trigger"},
				{ID: 2, Name: "Chrono Cross"},
			}, nil
		},
	}

	games, err := NewIGDBService(mock, &cacheMock{items: make(map[string]any)}).BuscarJogos(context.Background(), "Chrono Tr")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 2 || games[0].ID != 1 {
		t.Fatalf("games=%+v, esperava Chrono Trigger no topo e ids únicos", games)
	}
	if tier := calculateTier("Chrono Trigger", normalizeString("Chrono Tr"), strings.Fields(normalizeString("Chrono Tr"))); tier != 3 {
		t.Fatalf("tier=%d, esperava tier 3 para prefixo do último token", tier)
	}
}

func TestBuscarJogos_VersoesMantemPlataformasProprias(t *testing.T) {
	mock := &customSearchMock{
		searchFn: func(context.Context, string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{
					ID:               1,
					Name:             "Chrono Trigger",
					GameType:         igdbclient.GameTypeMainGame,
					Platforms:        []igdbclient.Platform{{ID: 1, Name: "Super Nintendo Entertainment System"}},
					TotalRatingCount: func() *int { value := 10; return &value }(),
				},
				{
					ID:        2,
					Name:      "Chrono Trigger",
					GameType:  igdbclient.GameTypePort,
					Platforms: []igdbclient.Platform{{ID: 2, Name: "PC (Microsoft Windows)"}},
				},
				{
					ID:        3,
					Name:      "Chrono Trigger: Character Library",
					GameType:  igdbclient.GameTypeExpandedGame,
					Platforms: []igdbclient.Platform{{ID: 3, Name: "Satellaview"}},
				},
			}, nil
		},
	}

	games, err := NewIGDBService(mock, &cacheMock{items: make(map[string]any)}).BuscarJogos(context.Background(), "Chrono Trigger")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 3 {
		t.Fatalf("games=%+v, esperava versões separadas", games)
	}
	if len(games[0].Platforms) != 1 || games[0].Platforms[0].Name != "Super Nintendo" {
		t.Fatalf("plataformas da primeira versão=%+v", games[0].Platforms)
	}
	if len(games[1].Platforms) != 1 || games[1].Platforms[0].Name != "PC" {
		t.Fatalf("plataformas da segunda versão=%+v", games[1].Platforms)
	}
	if games[2].Name != "Chrono Trigger: Character Library" || len(games[2].Platforms) != 1 || games[2].Platforms[0].Name != "Satellaview" {
		t.Fatalf("subtítulo=%+v", games[2])
	}
}

func TestBuscarJogos_VersoesChronoTrigger(t *testing.T) {
	dates := []int64{794880000, 943920000, 1227139200, 1519689600}
	games := []igdbclient.Game{
		{ID: 1802, Name: "Chrono Trigger", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &dates[0], Platforms: []igdbclient.Platform{{Name: "Super Nintendo Entertainment System"}}},
		{ID: 263446, Name: "Chrono Trigger", GameType: igdbclient.GameTypeExpandedGame, FirstReleaseDate: &dates[1], Platforms: []igdbclient.Platform{{Name: "PlayStation"}}},
		{ID: 20398, Name: "Chrono Trigger", GameType: igdbclient.GameTypePort, FirstReleaseDate: &dates[2], Platforms: []igdbclient.Platform{{Name: "Nintendo DS"}}},
		{ID: 206320, Name: "Chrono Trigger", GameType: igdbclient.GameTypePort, FirstReleaseDate: &dates[3], Platforms: []igdbclient.Platform{{Name: "PC (Microsoft Windows)"}}},
	}
	mock := &customSearchMock{searchFn: func(context.Context, string) ([]igdbclient.Game, error) { return games, nil }}
	result, err := NewIGDBService(mock, &cacheMock{items: make(map[string]any)}).BuscarJogos(context.Background(), "Chrono Trigger")
	if err != nil || len(result) != 4 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	want := map[int64]string{1802: "Super Nintendo", 263446: "PlayStation", 20398: "Nintendo DS", 206320: "PC"}
	for _, game := range result {
		if len(game.Platforms) != 1 || game.Platforms[0].Name != want[game.ID] {
			t.Fatalf("versão %d plataformas=%+v", game.ID, game.Platforms)
		}
	}
}

func TestBuscarJogoDetalheMapeiaPlataformas(t *testing.T) {
	mock := &igdbMock{detail: &igdbclient.Game{ID: 1, Platforms: []igdbclient.Platform{{Name: "PC (Microsoft Windows)"}, {Name: "PC"}, {Name: "Nintendo Entertainment System"}}}}
	game, err := NewIGDBService(mock, &cacheMock{items: make(map[string]any)}).BuscarJogo(context.Background(), 1)
	if err != nil || len(game.Platforms) != 2 || game.Platforms[0].Name != "PC" || game.Platforms[1].Name != "NES" {
		t.Fatalf("game=%+v err=%v", game, err)
	}
}

func TestBuscarJogoDetalheNaoEncontrado(t *testing.T) {
	game, err := NewIGDBService(&igdbMock{}, &cacheMock{items: make(map[string]any)}).BuscarJogo(context.Background(), 999)
	if game != nil || !errors.Is(err, ErrJogoIGDBNaoEncontrado) {
		t.Fatalf("game=%+v err=%v", game, err)
	}
}

func TestBuscarJogos_MapeiaPlataformasEDeduplicaAposMapeamento(t *testing.T) {
	mock := &customSearchMock{
		searchFn: func(context.Context, string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{{
				ID:       1,
				Name:     "Game",
				GameType: igdbclient.GameTypeMainGame,
				Platforms: []igdbclient.Platform{
					{ID: 1, Name: "PC (Microsoft Windows)"},
					{ID: 2, Name: "PC"},
					{ID: 3, Name: "Super Nintendo Entertainment System"},
					{ID: 4, Name: "Nintendo Entertainment System"},
					{ID: 5, Name: "Satellaview"},
				},
			}}, nil
		},
	}

	games, err := NewIGDBService(mock, &cacheMock{items: make(map[string]any)}).BuscarJogos(context.Background(), "Game")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"PC", "Super Nintendo", "NES", "Satellaview"}
	if len(games) != 1 || len(games[0].Platforms) != len(want) {
		t.Fatalf("games=%+v", games)
	}
	for i, platform := range games[0].Platforms {
		if platform.Name != want[i] {
			t.Fatalf("platform[%d]=%q, want %q", i, platform.Name, want[i])
		}
	}
}

func TestBuscarJogos_RebaixaSemAvaliacao(t *testing.T) {
	rating := 10
	mock := &customSearchMock{
		searchFn: func(context.Context, string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{ID: 1, Name: "Chrono Trigger", TotalRatingCount: &rating},
				{ID: 2, Name: "Chrono Trigger Character Library"},
				{ID: 3, Name: "Chrono Trigger Music Library"},
			}, nil
		},
	}

	games, err := NewIGDBService(mock, &cacheMock{items: make(map[string]any)}).BuscarJogos(context.Background(), "Chrono Trigger")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 3 || games[0].ID != 1 || games[1].ID != 2 || games[2].ID != 3 {
		t.Fatalf("games=%+v, esperava resultados sem avaliação no fim", games)
	}
}

func TestFiltrarJogosDesafioFranquia(t *testing.T) {
	refTime := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	pastDate := refTime.Add(-24 * time.Hour).Unix()
	futureDate := refTime.Add(24 * time.Hour).Unix()
	parentID := int64(999)
	franchiseID := int64(596)
	otherFranchiseID := int64(845)

	tests := []struct {
		name       string
		game       igdbclient.Game
		shouldKeep bool
	}{
		{
			name:       "main game valido",
			game:       igdbclient.Game{ID: 1, Name: "Zelda 1", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: true,
		},
		{
			name:       "remake valido",
			game:       igdbclient.Game{ID: 2, Name: "Link's Awakening Remake", GameType: igdbclient.GameTypeRemake, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: true,
		},
		{
			name:       "remaster valido",
			game:       igdbclient.Game{ID: 3, Name: "Wind Waker HD", GameType: igdbclient.GameTypeRemaster, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: true,
		},
		{
			name:       "expanded game valido",
			game:       igdbclient.Game{ID: 4, Name: "Expanded Zelda", GameType: igdbclient.GameTypeExpandedGame, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: true,
		},
		{
			name:       "port valido",
			game:       igdbclient.Game{ID: 5, Name: "Zelda Port", GameType: igdbclient.GameTypePort, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: true,
		},
		{
			name:       "fork valido",
			game:       igdbclient.Game{ID: 6, Name: "Zelda Fork", GameType: igdbclient.GameTypeFork, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: true,
		},
		{
			name:       "update valido",
			game:       igdbclient.Game{ID: 7, Name: "Zelda Update", GameType: igdbclient.GameTypeUpdate, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: true,
		},
		{
			name:       "dlc descartado",
			game:       igdbclient.Game{ID: 8, Name: "Zelda DLC", GameType: igdbclient.GameTypeDLC, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "expansion descartada",
			game:       igdbclient.Game{ID: 9, Name: "Zelda Expansion", GameType: igdbclient.GameTypeExpansion, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "bundle descartado",
			game:       igdbclient.Game{ID: 10, Name: "Zelda Bundle", GameType: igdbclient.GameTypeBundle, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "standalone expansion descartada",
			game:       igdbclient.Game{ID: 11, Name: "Zelda Standalone", GameType: igdbclient.GameTypeStandaloneExpansion, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "mod descartado",
			game:       igdbclient.Game{ID: 12, Name: "Zelda Mod", GameType: igdbclient.GameTypeMod, FirstReleaseDate: &pastDate, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "version parent descartado",
			game:       igdbclient.Game{ID: 13, Name: "Zelda Special Edition", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &pastDate, VersionParent: &parentID, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "data de lancamento nula descartada",
			game:       igdbclient.Game{ID: 14, Name: "Zelda Unreleased", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: nil, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "data futura descartada",
			game:       igdbclient.Game{ID: 15, Name: "Zelda Future", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &futureDate, Franchise: &franchiseID},
			shouldKeep: false,
		},
		{
			name:       "outra franquia descartada",
			game:       igdbclient.Game{ID: 16, Name: "Mario Game", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &pastDate, Franchise: &otherFranchiseID},
			shouldKeep: false,
		},
		{
			name:       "franchises contem franquia valida mantido",
			game:       igdbclient.Game{ID: 17, Name: "Zelda Crossover", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &pastDate, Franchises: []int64{franchiseID, otherFranchiseID}},
			shouldKeep: true,
		},
		{
			name:       "franchises nao contem franquia descartado",
			game:       igdbclient.Game{ID: 18, Name: "Other Crossover", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &pastDate, Franchises: []int64{otherFranchiseID}},
			shouldKeep: false,
		},
		{
			name:       "sem franchise nem franchises mantido",
			game:       igdbclient.Game{ID: 19, Name: "Zelda Unknown Franchise", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &pastDate},
			shouldKeep: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := FiltrarJogosDesafioFranquia(franchiseID, []igdbclient.Game{tc.game}, refTime)
			if tc.shouldKeep && len(res) != 1 {
				t.Fatalf("esperava manter o jogo, mas foi filtrado")
			}
			if !tc.shouldKeep && len(res) != 0 {
				t.Fatalf("esperava filtrar o jogo, mas foi mantido")
			}
		})
	}
}

func TestMapearTipoJogo(t *testing.T) {
	tests := []struct {
		gameType int
		want     string
	}{
		{igdbclient.GameTypeMainGame, "main_game"},
		{igdbclient.GameTypeRemake, "remake"},
		{igdbclient.GameTypeRemaster, "remaster"},
		{igdbclient.GameTypeExpandedGame, "expanded_game"},
		{igdbclient.GameTypePort, "port"},
		{igdbclient.GameTypeFork, "fork"},
		{igdbclient.GameTypeUpdate, "update"},
		{999, "outro"},
	}

	for _, tc := range tests {
		got := MapearTipoJogo(tc.gameType)
		if got != tc.want {
			t.Fatalf("MapearTipoJogo(%d) = %q, want %q", tc.gameType, got, tc.want)
		}
	}
}
