package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
)

type igdbMock struct {
	platformGamesCalls  int
	franchiseGamesCalls int
}

func (m *igdbMock) SearchGames(context.Context, string) ([]igdbclient.Game, error) {
	return []igdbclient.Game{}, nil
}

func (m *igdbMock) GameDetails(context.Context, int64) (*igdbclient.Game, error) {
	return nil, nil
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

func (m *igdbMock) GamesByFranchise(context.Context, int64) ([]igdbclient.Game, error) {
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
		name      string
		gameType  int
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

func TestBuscarJogos_AgrupamentoChronoTrigger(t *testing.T) {
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
					ParentGame:       &igdbclient.ParentGame{ID: 1802, Name: "Chrono Trigger"},
				},
				{
					ID:               20398,
					Name:             "Chrono Trigger",
					GameType:         igdbclient.GameTypePort,
					FirstReleaseDate: &dateDS,
					Platforms:        []igdbclient.Platform{{ID: 20, Name: "Nintendo DS"}},
					ParentGame:       &igdbclient.ParentGame{ID: 263446, Name: "Chrono Trigger"},
				},
				{
					ID:               206320,
					Name:             "Chrono Trigger",
					GameType:         igdbclient.GameTypePort,
					FirstReleaseDate: &dateSteam,
					Platforms:        []igdbclient.Platform{{ID: 6, Name: "PC (Microsoft Windows)"}},
					ParentGame:       &igdbclient.ParentGame{ID: 20398, Name: "Chrono Trigger"},
				},
			}, nil
		},
	}

	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	games, err := svc.BuscarJogos(context.Background(), "Chrono Trigger")
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 {
		t.Fatalf("esperava 1 único resultado para Chrono Trigger, obteve %d", len(games))
	}
	g := games[0]
	if g.ID != 1802 {
		t.Fatalf("id esperado 1802, obteve %d", g.ID)
	}
	if len(g.Platforms) != 4 {
		t.Fatalf("esperava 4 plataformas unificadas, obteve %d: %+v", len(g.Platforms), g.Platforms)
	}
	hasSNES := false
	hasPS := false
	for _, p := range g.Platforms {
		if p.Name == "Super Nintendo Entertainment System" {
			hasSNES = true
		}
		if p.Name == "PlayStation" {
			hasPS = true
		}
	}
	if !hasSNES || !hasPS {
		t.Fatalf("plataformas esperadas SNES e PlayStation não encontradas: %+v", g.Platforms)
	}
}

func TestBuscarJogos_AgrupamentoPaiAusente(t *testing.T) {
	dateParent := int64(1000000000)
	datePort := int64(1100000000)
	cover := &igdbclient.Image{ID: 1, ImageID: "co1111", URL: "//images.igdb.com/igdb/image/upload/t_thumb/co1111.jpg"}

	mock := &customSearchMock{
		searchFn: func(ctx context.Context, query string) ([]igdbclient.Game, error) {
			return []igdbclient.Game{
				{
					ID:               501,
					Name:             "Persona 3 Portable",
					GameType:         igdbclient.GameTypePort,
					FirstReleaseDate: &datePort,
					Platforms:        []igdbclient.Platform{{ID: 38, Name: "PlayStation Portable"}},
					ParentGame: &igdbclient.ParentGame{
						ID:               101,
						Name:             "Persona 3",
						FirstReleaseDate: &dateParent,
						Cover:            cover,
						Platforms:        []igdbclient.Platform{{ID: 8, Name: "PlayStation 2"}},
					},
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
		t.Fatalf("esperava 1 jogo pai sintetizado, obteve %d", len(games))
	}
	g := games[0]
	if g.ID != 101 || g.Name != "Persona 3" {
		t.Fatalf("dados do pai incorretos: %+v", g)
	}
	if len(g.Platforms) != 2 {
		t.Fatalf("esperava união das 2 plataformas, obteve %d: %+v", len(g.Platforms), g.Platforms)
	}
}

func TestBuscarJogos_RemakeNaoAgrupado(t *testing.T) {
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
					ParentGame:       &igdbclient.ParentGame{ID: 39, Name: "Mafia"},
				},
				{
					ID:         392531,
					Name:       "Mafia: Definitive Edition - Chicago Outfit Pack",
					GameType:   igdbclient.GameTypePack,
					ParentGame: &igdbclient.ParentGame{ID: 134070, Name: "Mafia: Definitive Edition"},
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

func TestBuscarJogos_Limite10(t *testing.T) {
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
	if len(games) != 10 {
		t.Fatalf("esperava exatamente 10 resultados no limite, obteve %d", len(games))
	}
}

func TestBuscarJogos_ErrosTraduzidos(t *testing.T) {
	tests := []struct {
		name       string
		clientErr  error
		targetErr  error
	}{
		{"rate limit", igdbclient.ErrRateLimited, ErrIGDBRateLimit},
		{"unavailable", igdbclient.ErrUnavailable, ErrIGDBIndisponivel},
		{"authentication", igdbclient.ErrAuthentication, ErrIGDBIndisponivel},
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

func TestBuscarJogos_QueryRepassadaAoClient(t *testing.T) {
	mock := &customSearchMock{}
	svc := NewIGDBService(mock, &cacheMock{items: make(map[string]any)})
	_, _ = svc.BuscarJogos(context.Background(), "  Super Mario  ")
	if mock.lastQuery != "Super Mario" {
		t.Fatalf("esperava consulta 'Super Mario', obteve %q", mock.lastQuery)
	}
}
