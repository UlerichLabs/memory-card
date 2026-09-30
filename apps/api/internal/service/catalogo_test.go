package service

import (
	"context"
	"testing"
	"time"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"github.com/UlerichLabs/memory-card/apps/api/internal/repository"
)

type catalogoClientMock struct {
	games               []igdbclient.Game
	franchiseGamesCalls int
	where               string
	sort                string
	limit               int
	offset              int
}

func (m *catalogoClientMock) FranchiseDetails(context.Context, int64) (*igdbclient.Franchise, error) {
	return &igdbclient.Franchise{ID: 596, Name: "The Legend of Zelda"}, nil
}
func (m *catalogoClientMock) GamesByFranchiseParaDesafio(context.Context, int64) ([]igdbclient.Game, error) {
	m.franchiseGamesCalls++
	return m.games, nil
}
func (m *catalogoClientMock) CatalogGames(_ context.Context, where, sort string, limit, offset int) ([]igdbclient.Game, error) {
	m.where, m.sort, m.limit, m.offset = where, sort, limit, offset
	return m.games, nil
}
func (m *catalogoClientMock) CountGames(context.Context, string) (int, error) { return 1742, nil }
func (m *catalogoClientMock) Genres(context.Context) ([]igdbclient.Genre, error) {
	return []igdbclient.Genre{{ID: 12, Name: "Role-playing (RPG)", Slug: "role-playing-rpg"}, {ID: 99, Name: "Unknown"}}, nil
}

type catalogoCacheMock struct{ values map[string]interface{} }

func (m *catalogoCacheMock) Buscar(_ context.Context, key string, destino interface{}) (bool, error) {
	value, found := m.values[key]
	if !found {
		return false, nil
	}
	switch target := destino.(type) {
	case *[]igdbclient.Game:
		*target = value.([]igdbclient.Game)
	case *[]igdbclient.Genre:
		*target = value.([]igdbclient.Genre)
	}
	return true, nil
}
func (m *catalogoCacheMock) Salvar(_ context.Context, key string, value interface{}) error {
	m.values[key] = value
	return nil
}

func TestCatalogoService_FranquiaFiltraSugereBuscaOrdenaEPagina(t *testing.T) {
	past := time.Now().Add(-time.Hour).Unix()
	client := &catalogoClientMock{games: []igdbclient.Game{
		{ID: 1, Name: "Zelda II: The Adventure of Link", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &past, TotalRatingCount: intPointer(10)},
		{ID: 2, Name: "Link: The Faces of Evil", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &past, TotalRatingCount: intPointer(20)},
		{ID: 3, Name: "Zeldaverse", GameType: igdbclient.GameTypeMainGame, FirstReleaseDate: &past, TotalRatingCount: intPointer(30)},
		{ID: 4, Name: "The Legend of Zelda: Breath of the Wild - Nintendo Switch 2 Edition", GameType: igdbclient.GameTypeExpandedGame, FirstReleaseDate: &past, TotalRatingCount: intPointer(40)},
	}}
	repo := &mockListasRepo{listarJogosZeradosUsuarioFn: func(context.Context, int32) ([]*repository.JogoZeradoResumo, error) { return nil, nil }}
	cache := &catalogoCacheMock{values: make(map[string]interface{})}
	svc := NewCatalogoService(client, cache, repo)
	result, err := svc.Jogos(context.Background(), 1, CatalogoFiltro{Origem: CatalogoFranquia, ID: 596, Ordenar: "populares", Pagina: 1, PorPagina: 3, Agora: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Meta.Total != 4 || result.Meta.TotalSugeridos == nil || *result.Meta.TotalSugeridos != 1 {
		t.Fatalf("meta inesperada: %+v", result.Meta)
	}
	if len(result.Itens) != 3 || result.Itens[0].Nome != "The Legend of Zelda: Breath of the Wild - Nintendo Switch 2 Edition" {
		t.Fatalf("ordenação/paginação inesperada: %+v", result.Itens)
	}
	if result.Itens[0].Sugerido || result.Itens[1].Sugerido {
		t.Fatalf("sugestões inesperadas: %+v", result.Itens)
	}
	if client.franchiseGamesCalls != 1 {
		t.Fatalf("chamadas ao IGDB=%d, esperado 1", client.franchiseGamesCalls)
	}
	if _, found := cache.values[chaveSnapshotFranquiaDesafio(596)]; !found {
		t.Fatalf("snapshot não gravado na chave v2: %+v", cache.values)
	}

	result, err = svc.Jogos(context.Background(), 1, CatalogoFiltro{Origem: CatalogoFranquia, ID: 596, Busca: "zelda", Ordenar: "nome", Pagina: 1, PorPagina: 60, Agora: time.Now()})
	if err != nil || result.Meta.Total != 3 {
		t.Fatalf("busca inesperada: result=%+v err=%v", result, err)
	}
}

func TestCatalogoService_PlataformaMontaFiltroEGenerosTraduzem(t *testing.T) {
	client := &catalogoClientMock{games: []igdbclient.Game{{ID: 10, Name: "Mario"}}}
	repo := &mockListasRepo{}
	svc := NewCatalogoService(client, &catalogoCacheMock{values: make(map[string]interface{})}, repo)
	genreID := int64(12)
	result, err := svc.Jogos(context.Background(), 1, CatalogoFiltro{Origem: CatalogoPlataforma, ID: 19, GeneroID: &genreID, Busca: `mario"`, Ordenar: "nome", Pagina: 2, PorPagina: 20, Agora: time.Unix(1700000000, 0)})
	if err != nil || result.Meta.Total != 1742 {
		t.Fatalf("catálogo de plataforma inesperado: result=%+v err=%v", result, err)
	}
	if client.limit != 20 || client.offset != 20 || client.sort != "name asc" || client.where == "" || client.where == "name ~ *\"mario\"*" {
		t.Fatalf("query do catálogo inesperada: where=%q sort=%q limit=%d offset=%d", client.where, client.sort, client.limit, client.offset)
	}
	genres, err := svc.Generos(context.Background())
	if err != nil || genres[0].Nome != "RPG" || genres[1].Nome != "Unknown" {
		t.Fatalf("gêneros inesperados: %+v err=%v", genres, err)
	}
}

func intPointer(value int) *int { return &value }
