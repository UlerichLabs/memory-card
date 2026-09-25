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
	ErrTermoIGDBVazio   = errors.New("igdb.search.empty")
	ErrIGDBIndisponivel = errors.New("igdb.unavailable")
	ErrIGDBRateLimit    = errors.New("igdb.rate_limited")
)

var normTransformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

type IGDBClient interface {
	SearchGames(context.Context, string) ([]igdbclient.Game, error)
	GameDetails(context.Context, int64) (*igdbclient.Game, error)
	Platforms(context.Context) ([]igdbclient.Platform, error)
	GamesByPlatform(context.Context, int64) ([]igdbclient.Game, error)
	SearchFranchises(context.Context, string) ([]igdbclient.Franchise, error)
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
	termo = strings.TrimSpace(termo)
	if termo == "" {
		return nil, ErrTermoIGDBVazio
	}
	games, err := svc.client.SearchGames(ctx, termo)
	if err != nil {
		if errors.Is(err, igdbclient.ErrRateLimited) {
			return nil, fmt.Errorf("%w: %w", ErrIGDBRateLimit, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrIGDBIndisponivel, err)
	}

	var filtered []igdbclient.Game
	for _, g := range games {
		if isAllowedGameType(g.GameType) {
			filtered = append(filtered, g)
		}
	}

	gamesByID := make(map[int64]*igdbclient.Game, len(filtered))
	for i := range filtered {
		g := filtered[i]
		gamesByID[g.ID] = &g
	}

	syntheticParents := make(map[int64]*igdbclient.Game)
	mergedIDs := make(map[int64]bool)

	for _, g := range filtered {
		parentName := getParentName(g, gamesByID)
		if !shouldGroupIntoParent(g, parentName) {
			continue
		}
		targetID := resolveTargetParentID(getParentID(g), gamesByID)
		if targetID == 0 || targetID == g.ID {
			continue
		}
		targetGame, exists := gamesByID[targetID]
		if !exists {
			synth, synthExists := syntheticParents[targetID]
			if !synthExists {
				if g.ParentGame != nil && g.ParentGame.ID == targetID {
					p := &igdbclient.Game{
						ID:               g.ParentGame.ID,
						Name:             g.ParentGame.Name,
						FirstReleaseDate: g.ParentGame.FirstReleaseDate,
						Cover:            g.ParentGame.Cover,
						Platforms:        append([]igdbclient.Platform(nil), g.ParentGame.Platforms...),
						GameType:         igdbclient.GameTypeMainGame,
					}
					syntheticParents[targetID] = p
					targetGame = p
				}
			} else {
				targetGame = synth
			}
		}
		if targetGame != nil {
			mergePlatforms(&targetGame.Platforms, g.Platforms)
			if targetGame.TotalRatingCount == nil && g.TotalRatingCount != nil {
				targetGame.TotalRatingCount = g.TotalRatingCount
			}
			if targetGame.Summary == "" && g.Summary != "" {
				targetGame.Summary = g.Summary
			}
			if len(targetGame.Genres) == 0 && len(g.Genres) > 0 {
				targetGame.Genres = g.Genres
			}
			mergedIDs[g.ID] = true
		}
	}

	finalGames := make([]igdbclient.Game, 0, len(filtered))
	for _, g := range filtered {
		if !mergedIDs[g.ID] {
			finalGames = append(finalGames, *gamesByID[g.ID])
		}
	}
	for _, p := range syntheticParents {
		finalGames = append(finalGames, *p)
	}

	normTerm := normalizeString(termo)
	termTokens := strings.Fields(normTerm)

	sort.SliceStable(finalGames, func(i, j int) bool {
		tierA := calculateTier(finalGames[i].Name, normTerm, termTokens)
		tierB := calculateTier(finalGames[j].Name, normTerm, termTokens)
		if tierA != tierB {
			return tierA < tierB
		}
		ratingA := 0
		if finalGames[i].TotalRatingCount != nil {
			ratingA = *finalGames[i].TotalRatingCount
		}
		ratingB := 0
		if finalGames[j].TotalRatingCount != nil {
			ratingB = *finalGames[j].TotalRatingCount
		}
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

	if len(finalGames) > 10 {
		finalGames = finalGames[:10]
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

func getParentID(g igdbclient.Game) int64 {
	if g.ParentGame != nil && g.ParentGame.ID > 0 {
		return g.ParentGame.ID
	}
	if g.VersionParent != nil && *g.VersionParent > 0 {
		return *g.VersionParent
	}
	return 0
}

func getParentName(g igdbclient.Game, gamesByID map[int64]*igdbclient.Game) string {
	if g.ParentGame != nil && g.ParentGame.Name != "" {
		return g.ParentGame.Name
	}
	if pid := getParentID(g); pid > 0 {
		if p, ok := gamesByID[pid]; ok {
			return p.Name
		}
	}
	return ""
}

func shouldGroupIntoParent(g igdbclient.Game, parentName string) bool {
	hasParent := (g.ParentGame != nil && g.ParentGame.ID > 0) || (g.VersionParent != nil && *g.VersionParent > 0)
	if !hasParent {
		return false
	}
	if g.GameType == igdbclient.GameTypeRemake {
		return false
	}
	if g.GameType == igdbclient.GameTypeStandaloneExpansion {
		return false
	}
	if g.GameType == igdbclient.GameTypeRemaster {
		return parentName != "" && normalizeString(g.Name) == normalizeString(parentName)
	}
	return true
}

func resolveTargetParentID(startID int64, gamesByID map[int64]*igdbclient.Game) int64 {
	curr := startID
	visited := map[int64]bool{curr: true}
	for {
		g, exists := gamesByID[curr]
		if !exists {
			return curr
		}
		pName := getParentName(*g, gamesByID)
		if !shouldGroupIntoParent(*g, pName) {
			return curr
		}
		nextID := getParentID(*g)
		if nextID == 0 || visited[nextID] {
			return curr
		}
		visited[nextID] = true
		curr = nextID
	}
}

func mergePlatforms(target *[]igdbclient.Platform, extra []igdbclient.Platform) {
	seenID := make(map[int64]bool, len(*target))
	seenName := make(map[string]bool, len(*target))
	for _, p := range *target {
		if p.ID > 0 {
			seenID[p.ID] = true
		}
		if p.Name != "" {
			seenName[strings.ToLower(strings.TrimSpace(p.Name))] = true
		}
	}
	for _, p := range extra {
		norm := strings.ToLower(strings.TrimSpace(p.Name))
		if (p.ID > 0 && seenID[p.ID]) || (norm != "" && seenName[norm]) {
			continue
		}
		if p.ID > 0 {
			seenID[p.ID] = true
		}
		if norm != "" {
			seenName[norm] = true
		}
		*target = append(*target, p)
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
	if strings.HasPrefix(normName, normTerm) {
		return 2
	}
	if len(termTokens) > 0 && containsAllTokens(normName, termTokens) {
		return 3
	}
	return 4
}

func containsAllTokens(normName string, tokens []string) bool {
	for _, t := range tokens {
		if !strings.Contains(normName, t) {
			return false
		}
	}
	return true
}

func (svc *IGDBService) BuscarJogo(ctx context.Context, id int64) (*igdbclient.Game, error) {
	return svc.client.GameDetails(ctx, id)
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
