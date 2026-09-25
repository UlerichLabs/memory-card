package igdbclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrAuthentication = errors.New("igdb authentication failed")
	ErrRateLimited    = errors.New("igdb rate limit exceeded")
	ErrUnavailable    = errors.New("igdb unavailable")
)

type Config struct {
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
	TokenURL     string
	APIURL       string
	Now          func() time.Time
}

type Client struct {
	client       *http.Client
	clientID     string
	clientSecret string
	tokenURL     string
	apiURL       string
	now          func() time.Time
	tokenMu      sync.Mutex
	token        string
	tokenExpiry  time.Time
	rateMu       sync.Mutex
	nextRequest  time.Time
}

const (
	GameTypeMainGame            = 0
	GameTypeDLC                 = 1
	GameTypeExpansion           = 2
	GameTypeBundle              = 3
	GameTypeStandaloneExpansion = 4
	GameTypeMod                 = 5
	GameTypeEpisode             = 6
	GameTypeSeason              = 7
	GameTypeRemake              = 8
	GameTypeRemaster            = 9
	GameTypeExpandedGame        = 10
	GameTypePort                = 11
	GameTypeFork                = 12
	GameTypePack                = 13
	GameTypeUpdate              = 14
)

type Game struct {
	ID               int64       `json:"id"`
	Name             string      `json:"name"`
	Cover            *Image      `json:"cover,omitempty"`
	FirstReleaseDate *int64      `json:"first_release_date,omitempty"`
	Platforms        []Platform  `json:"platforms,omitempty"`
	Genres           []Genre     `json:"genres,omitempty"`
	Summary          string      `json:"summary,omitempty"`
	GameType         int         `json:"game_type"`
	VersionParent    *int64      `json:"version_parent,omitempty"`
	TotalRatingCount *int        `json:"total_rating_count,omitempty"`
	ParentGame       *ParentGame `json:"parent_game,omitempty"`
}

type ParentGame struct {
	ID               int64      `json:"id"`
	Name             string     `json:"name"`
	FirstReleaseDate *int64     `json:"first_release_date,omitempty"`
	Cover            *Image     `json:"cover,omitempty"`
	Platforms        []Platform `json:"platforms,omitempty"`
}

type Platform struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Genre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Franchise struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Image struct {
	ID      int64  `json:"id"`
	ImageID string `json:"image_id,omitempty"`
	URL     string `json:"url,omitempty"`
}

func (img *Image) EnsureURL() {
	if img != nil && img.URL == "" && img.ImageID != "" {
		img.URL = "//images.igdb.com/igdb/image/upload/t_thumb/" + img.ImageID + ".jpg"
	}
}

type accessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func New(cfg Config) *Client {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	tokenURL := cfg.TokenURL
	if tokenURL == "" {
		tokenURL = "https://id.twitch.tv/oauth2/token"
	}
	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = "https://api.igdb.com/v4"
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &Client{client: client, clientID: cfg.ClientID, clientSecret: cfg.ClientSecret, tokenURL: tokenURL, apiURL: strings.TrimRight(apiURL, "/"), now: now}
}

func (c *Client) SearchGames(ctx context.Context, query string) ([]Game, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Game{}, nil
	}
	body := fmt.Sprintf("fields id, name, first_release_date, summary, cover.image_id, genres.name, game_type, version_parent, total_rating_count, platforms.name, parent_game.id, parent_game.name, parent_game.first_release_date, parent_game.cover.image_id, parent_game.platforms.name; search %q; limit 50;", query)
	games, err := c.games(ctx, body)
	if err != nil {
		return nil, err
	}
	for i := range games {
		games[i].Cover.EnsureURL()
		if games[i].ParentGame != nil {
			games[i].ParentGame.Cover.EnsureURL()
		}
	}
	return games, nil
}

func (c *Client) GameDetails(ctx context.Context, id int64) (*Game, error) {
	games, err := c.games(ctx, fmt.Sprintf("fields id,name,cover.url,first_release_date,platforms.name,genres.name,summary; where id = %d; limit 1;", id))
	if err != nil || len(games) == 0 {
		return nil, err
	}
	games[0].Cover.EnsureURL()
	return &games[0], nil
}

func (c *Client) Platforms(ctx context.Context) ([]Platform, error) {
	var result []Platform
	if err := c.query(ctx, "platforms", "fields id,name; limit 500;", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GamesByPlatform(ctx context.Context, id int64) ([]Game, error) {
	return c.games(ctx, fmt.Sprintf("fields id,name,cover.url,first_release_date; where platforms = (%d); limit 500;", id))
}

func (c *Client) SearchFranchises(ctx context.Context, query string) ([]Franchise, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Franchise{}, nil
	}
	var result []Franchise
	if err := c.query(ctx, "franchises", fmt.Sprintf("fields id,name; where name ~ *%q*; limit 20;", query), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GamesByFranchise(ctx context.Context, id int64) ([]Game, error) {
	return c.games(ctx, fmt.Sprintf("fields id,name,cover.url,first_release_date; where franchises = (%d); limit 500;", id))
}

func (c *Client) games(ctx context.Context, query string) ([]Game, error) {
	var result []Game
	if err := c.query(ctx, "games", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) query(ctx context.Context, endpoint, body string, target any) error {
	if err := c.wait(ctx); err != nil {
		return fmt.Errorf("aguardar janela de requisicao IGDB: %w", err)
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/"+endpoint, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("montar requisicao IGDB: %w", err)
	}
	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "text/plain")
	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimited
	}
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return ErrAuthentication
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%w: status %d", ErrUnavailable, response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(target); err != nil {
		return fmt.Errorf("decodificar resposta IGDB: %w", ErrUnavailable)
	}
	return nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" && c.now().Add(30*time.Second).Before(c.tokenExpiry) {
		return c.token, nil
	}
	if c.clientID == "" || c.clientSecret == "" {
		return "", ErrAuthentication
	}
	form := url.Values{"client_id": {c.clientID}, "client_secret": {c.clientSecret}, "grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("montar autenticacao IGDB: %w", ErrAuthentication)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAuthentication, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", ErrAuthentication
	}
	var token accessToken
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&token); err != nil || token.AccessToken == "" || token.ExpiresIn <= 0 {
		return "", ErrAuthentication
	}
	c.token = token.AccessToken
	c.tokenExpiry = c.now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return c.token, nil
}

func (c *Client) wait(ctx context.Context) error {
	c.rateMu.Lock()
	now := time.Now()
	when := c.nextRequest
	if when.Before(now) {
		when = now
	}
	c.nextRequest = when.Add(250 * time.Millisecond)
	c.rateMu.Unlock()
	wait := time.Until(when)
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
