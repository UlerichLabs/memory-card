package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/UlerichLabs/memory-card/apps/api/internal/igdbclient"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrTermoIGDBVazio        = errors.New("igdb.search.empty")
	ErrTermoIGDBCurto        = errors.New("igdb.search.too_short")
	ErrIGDBIndisponivel      = errors.New("igdb.unavailable")
	ErrIGDBRateLimit         = errors.New("igdb.rate_limited")
	ErrIGDBQueryInvalida     = errors.New("igdb.query.invalid")
	ErrJogoIGDBNaoEncontrado = errors.New("jogos.not_found")
)

var normTransformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

type IGDBClient interface {
	SearchGames(context.Context, string) ([]igdbclient.Game, error)
	GameDetails(context.Context, int64) (*igdbclient.Game, error)
	Platforms(context.Context) ([]igdbclient.Platform, error)
	GamesByPlatform(context.Context, int64) ([]igdbclient.Game, error)
	SearchFranchises(context.Context, string) ([]igdbclient.Franchise, error)
	FranchiseDetails(context.Context, int64) (*igdbclient.Franchise, error)
	GamesByFranchise(context.Context, int64) ([]igdbclient.Game, error)
}

type IGDBSnapshotRepository interface {
	Buscar(context.Context, string, any) (bool, error)
	Salvar(context.Context, string, any) error
}

type IGDBService struct {
	client IGDBClient
	cache  IGDBSnapshotRepository
}

func NewIGDBService(client IGDBClient, cache IGDBSnapshotRepository) *IGDBService {
	return &IGDBService{client: client, cache: cache}
}

func (svc *IGDBService) BuscarJogos(ctx context.Context, termo string) ([]igdbclient.Game, error) {
	termo = sanitizarTermoIGDB(termo)
	if termo == "" {
		return nil, ErrTermoIGDBVazio
	}
	if len([]rune(termo)) < 2 {
		return nil, ErrTermoIGDBCurto
	}
	games, err := svc.client.SearchGames(ctx, termo)
	if err != nil {
		return nil, traduzErroIGDB(err)
	}
	games = deduplicateGames(games)

	var filtered []igdbclient.Game
	for _, g := range games {
		if isAllowedGameType(g.GameType) {
			normalizePlatforms(&g.Platforms)
			filtered = append(filtered, g)
		}
	}
	finalGames := filtered

	normTerm := normalizeString(termo)
	termTokens := strings.Fields(normTerm)
	hasExactRated := false
	for _, game := range finalGames {
		if calculateTier(game.Name, normTerm, termTokens) == 1 && ratingCount(game) > 0 {
			hasExactRated = true
			break
		}
	}

	sort.SliceStable(finalGames, func(i, j int) bool {
		if hasExactRated {
			unratedA := ratingCount(finalGames[i]) == 0
			unratedB := ratingCount(finalGames[j]) == 0
			if unratedA != unratedB {
				return !unratedA
			}
		}
		tierA := calculateTier(finalGames[i].Name, normTerm, termTokens)
		tierB := calculateTier(finalGames[j].Name, normTerm, termTokens)
		if tierA != tierB {
			return tierA < tierB
		}
		ratingA := ratingCount(finalGames[i])
		ratingB := ratingCount(finalGames[j])
		if ratingA != ratingB {
			return ratingA > ratingB
		}
		hasDateA := finalGames[i].FirstReleaseDate != nil
		hasDateB := finalGames[j].FirstReleaseDate != nil
		if hasDateA && !hasDateB {
			return true
		}
		if !hasDateA && hasDateB {
			return false
		}
		if hasDateA && hasDateB {
			dateA := *finalGames[i].FirstReleaseDate
			dateB := *finalGames[j].FirstReleaseDate
			if dateA != dateB {
				return dateA < dateB
			}
		}
		return false
	})

	if len(finalGames) > 20 {
		finalGames = finalGames[:20]
	}

	return finalGames, nil
}

func isAllowedGameType(gt int) bool {
	switch gt {
	case igdbclient.GameTypeMainGame,
		igdbclient.GameTypeStandaloneExpansion,
		igdbclient.GameTypeRemake,
		igdbclient.GameTypeRemaster,
		igdbclient.GameTypeExpandedGame,
		igdbclient.GameTypePort,
		igdbclient.GameTypeFork:
		return true
	default:
		return false
	}
}

func deduplicateGames(games []igdbclient.Game) []igdbclient.Game {
	unique := make([]igdbclient.Game, 0, len(games))
	seen := make(map[int64]bool, len(games))
	for _, game := range games {
		if seen[game.ID] {
			continue
		}
		seen[game.ID] = true
		unique = append(unique, game)
	}
	return unique
}

func normalizePlatforms(platforms *[]igdbclient.Platform) {
	seen := make(map[string]bool, len(*platforms))
	result := make([]igdbclient.Platform, 0, len(*platforms))
	for _, platform := range *platforms {
		platform.Name = platformDisplayName(platform.Name)
		key := strings.ToLower(strings.TrimSpace(platform.Name))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, platform)
	}
	*platforms = result
}

func platformDisplayName(name string) string {
	switch name {
	case "PC (Microsoft Windows)":
		return "PC"
	case "Super Nintendo Entertainment System":
		return "Super Nintendo"
	case "Nintendo Entertainment System":
		return "NES"
	default:
		return name
	}
}

func normalizeString(s string) string {
	lower := strings.ToLower(s)
	clean, _, err := transform.String(normTransformer, lower)
	if err != nil {
		clean = lower
	}
	var sb strings.Builder
	for _, r := range clean {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

func calculateTier(name, normTerm string, termTokens []string) int {
	normName := normalizeString(name)
	if normName == normTerm {
		return 1
	}
	if strings.HasPrefix(normName, normTerm) && (len(normName) == len(normTerm) || normName[len(normTerm)] == ' ') {
		return 2
	}
	if len(termTokens) > 0 && containsAllTokens(normName, termTokens) {
		return 3
	}
	return 4
}

func containsAllTokens(normName string, tokens []string) bool {
	words := strings.Fields(normName)
	for i, token := range tokens {
		matched := false
		for _, word := range words {
			if (i == len(tokens)-1 && strings.HasPrefix(word, token)) || (i != len(tokens)-1 && strings.Contains(word, token)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func ratingCount(game igdbclient.Game) int {
	if game.TotalRatingCount == nil {
		return 0
	}
	return *game.TotalRatingCount
}

func traduzErroIGDB(err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w", ErrIGDBIndisponivel, err)
	}
	if errors.Is(err, igdbclient.ErrQueryInvalid) {
		return fmt.Errorf("%w: %w", ErrIGDBQueryInvalida, err)
	}
	if errors.Is(err, igdbclient.ErrRateLimited) {
		return fmt.Errorf("%w: %w", ErrIGDBRateLimit, err)
	}
	return fmt.Errorf("%w: %w", ErrIGDBIndisponivel, err)
}

func sanitizarTermoIGDB(value string) string {
	var sanitized strings.Builder
	for _, character := range value {
		switch character {
		case '"', '\\', '*', ';':
			sanitized.WriteRune(' ')
		default:
			sanitized.WriteRune(character)
		}
	}
	return strings.Join(strings.Fields(strings.TrimRight(strings.TrimSpace(sanitized.String()), ".")), " ")
}

func (svc *IGDBService) BuscarJogo(ctx context.Context, id int64) (*igdbclient.Game, error) {
	game, err := svc.client.GameDetails(ctx, id)
	if err != nil {
		return nil, traduzErroIGDB(err)
	}
	if game == nil {
		return nil, ErrJogoIGDBNaoEncontrado
	}
	normalizePlatforms(&game.Platforms)
	return game, nil
}

func (svc *IGDBService) ListarPlataformas(ctx context.Context) ([]igdbclient.Platform, error) {
	return svc.client.Platforms(ctx)
}

func (svc *IGDBService) BuscarFranquias(ctx context.Context, termo string) ([]igdbclient.Franchise, error) {
	if strings.TrimSpace(termo) == "" {
		return nil, ErrTermoIGDBVazio
	}
	return svc.client.SearchFranchises(ctx, termo)
}

func (svc *IGDBService) ObterFranquia(ctx context.Context, id int64) (*igdbclient.Franchise, error) {
	franchise, err := svc.client.FranchiseDetails(ctx, id)
	if err != nil {
		return nil, traduzErroIGDB(err)
	}
	return franchise, nil
}

func (svc *IGDBService) JogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.jogosSnapshot(ctx, fmt.Sprintf("platform:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByPlatform(ctx, id)
	})
}

func (svc *IGDBService) JogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.jogosSnapshot(ctx, fmt.Sprintf("franchise:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByFranchise(ctx, id)
	})
}

func (svc *IGDBService) AtualizarJogosDaPlataforma(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.atualizarSnapshot(ctx, fmt.Sprintf("platform:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByPlatform(ctx, id)
	})
}

func (svc *IGDBService) AtualizarJogosDaFranquia(ctx context.Context, id int64) ([]igdbclient.Game, error) {
	return svc.atualizarSnapshot(ctx, fmt.Sprintf("franchise:%d", id), func() ([]igdbclient.Game, error) {
		return svc.client.GamesByFranchise(ctx, id)
	})
}

func (svc *IGDBService) jogosSnapshot(ctx context.Context, key string, buscar func() ([]igdbclient.Game, error)) ([]igdbclient.Game, error) {
	var games []igdbclient.Game
	found, err := svc.cache.Buscar(ctx, key, &games)
	if err != nil {
		return nil, err
	}
	if found {
		return games, nil
	}
	return svc.atualizarSnapshot(ctx, key, buscar)
}

func (svc *IGDBService) atualizarSnapshot(ctx context.Context, key string, buscar func() ([]igdbclient.Game, error)) ([]igdbclient.Game, error) {
	games, err := buscar()
	if err != nil {
		return nil, err
	}
	if games == nil {
		games = []igdbclient.Game{}
	}
	if err := svc.cache.Salvar(ctx, key, games); err != nil {
		return nil, err
	}
	return games, nil
}
