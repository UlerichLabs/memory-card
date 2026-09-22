package igdbclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestClientSearchGamesOAuthAndCache(t *testing.T) {
	tokenRequests := 0
	apiRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth":
			tokenRequests++
			if err := r.ParseForm(); err != nil || r.Form.Get("client_id") != "client" || r.Form.Get("client_secret") != "secret" {
				t.Fatal("oauth credentials missing")
			}
			_ = json.NewEncoder(w).Encode(accessToken{AccessToken: "private-token", ExpiresIn: 3600})
		case "/v4/games":
			apiRequests++
			if r.Header.Get("Authorization") != "Bearer private-token" || r.Header.Get("Client-ID") != "client" {
				t.Fatal("igdb authorization headers missing")
			}
			_, _ = fmt.Fprint(w, `[{"id":1,"name":"Game","cover":{"url":"//cover"}}]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := New(Config{ClientID: "client", ClientSecret: "secret", HTTPClient: server.Client(), TokenURL: server.URL + "/oauth", APIURL: server.URL + "/v4"})
	for range 2 {
		games, err := client.SearchGames(context.Background(), "Game")
		if err != nil || len(games) != 1 || games[0].Name != "Game" || games[0].Cover.URL != "//cover" {
			t.Fatalf("games=%+v err=%v", games, err)
		}
	}
	if tokenRequests != 1 || apiRequests != 2 {
		t.Fatalf("token requests=%d api requests=%d", tokenRequests, apiRequests)
	}
}

func TestClientSearchWithoutResults(t *testing.T) {
	server := mockIGDB(t, http.StatusOK, "[]")
	defer server.Close()
	client := New(Config{ClientID: "client", ClientSecret: "secret", HTTPClient: server.Client(), TokenURL: server.URL + "/oauth", APIURL: server.URL + "/v4"})
	games, err := client.SearchGames(context.Background(), "absent")
	if err != nil || len(games) != 0 {
		t.Fatalf("games=%v err=%v", games, err)
	}
}

func TestClientRenewsTokenBeforeExpiration(t *testing.T) {
	tokenRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth" {
			tokenRequests++
			_ = json.NewEncoder(w).Encode(accessToken{AccessToken: fmt.Sprintf("token-%d", tokenRequests), ExpiresIn: 3600})
			return
		}
		_, _ = fmt.Fprint(w, `[]`)
	}))
	defer server.Close()
	now := time.Now()
	client := New(Config{ClientID: "client", ClientSecret: "secret", HTTPClient: server.Client(), TokenURL: server.URL + "/oauth", APIURL: server.URL + "/v4", Now: func() time.Time { return now }})
	if _, err := client.SearchGames(context.Background(), "first"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(59*time.Minute + 40*time.Second)
	if _, err := client.SearchGames(context.Background(), "second"); err != nil {
		t.Fatal(err)
	}
	if tokenRequests != 2 {
		t.Fatalf("token requests=%d", tokenRequests)
	}
}

func TestClientAuthenticationFailureIsSanitized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "secret detail", http.StatusUnauthorized) }))
	defer server.Close()
	client := New(Config{ClientID: "client", ClientSecret: "secret", HTTPClient: server.Client(), TokenURL: server.URL + "/oauth", APIURL: server.URL + "/v4"})
	_, err := client.SearchGames(context.Background(), "game")
	if !errors.Is(err, ErrAuthentication) {
		t.Fatalf("error=%v", err)
	}
	if err.Error() == "" || contains(err.Error(), "secret detail") || contains(err.Error(), "secret") {
		t.Fatalf("authentication error leaked details: %v", err)
	}
}

func TestClientRateLimit(t *testing.T) {
	server := mockIGDB(t, http.StatusTooManyRequests, "{}")
	defer server.Close()
	client := New(Config{ClientID: "client", ClientSecret: "secret", HTTPClient: server.Client(), TokenURL: server.URL + "/oauth", APIURL: server.URL + "/v4"})
	_, err := client.SearchGames(context.Background(), "game")
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("error=%v", err)
	}
}

func TestClientThrottlesConcurrentRequests(t *testing.T) {
	var mu sync.Mutex
	var starts []time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth" {
			_ = json.NewEncoder(w).Encode(accessToken{AccessToken: "token", ExpiresIn: 3600})
			return
		}
		mu.Lock()
		starts = append(starts, time.Now())
		mu.Unlock()
		_, _ = fmt.Fprint(w, `[]`)
	}))
	defer server.Close()
	client := New(Config{ClientID: "client", ClientSecret: "secret", HTTPClient: server.Client(), TokenURL: server.URL + "/oauth", APIURL: server.URL + "/v4"})
	var group sync.WaitGroup
	for range 5 {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := client.SearchGames(context.Background(), "game"); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	mu.Lock()
	defer mu.Unlock()
	if len(starts) != 5 {
		t.Fatalf("request count=%d", len(starts))
	}
	for i := 1; i < len(starts); i++ {
		if starts[i].Sub(starts[i-1]) < 240*time.Millisecond {
			t.Fatalf("requests started too close: %v", starts[i].Sub(starts[i-1]))
		}
	}
}

func mockIGDB(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth" {
			_ = json.NewEncoder(w).Encode(accessToken{AccessToken: "token", ExpiresIn: 3600})
			return
		}
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, body)
	}))
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}

func TestClientSearchFranchisesQuery(t *testing.T) {
	var requestBody string
	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth":
			_ = json.NewEncoder(w).Encode(accessToken{AccessToken: "token", ExpiresIn: 3600})
		case "/v4/franchises":
			requestPath = r.URL.Path
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ler body: %v", err)
			}
			requestBody = string(bodyBytes)
			_, _ = fmt.Fprint(w, `[{"id":596,"name":"The Legend of Zelda"}]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(Config{
		ClientID:     "client",
		ClientSecret: "secret",
		HTTPClient:   server.Client(),
		TokenURL:     server.URL + "/oauth",
		APIURL:       server.URL + "/v4",
	})

	franchises, err := client.SearchFranchises(context.Background(), "zelda")
	if err != nil {
		t.Fatalf("SearchFranchises: %v", err)
	}
	if len(franchises) != 1 || franchises[0].Name != "The Legend of Zelda" {
		t.Fatalf("franchises=%+v", franchises)
	}
	if requestPath != "/v4/franchises" {
		t.Fatalf("path=%s, esperava /v4/franchises", requestPath)
	}
	expectedClause := `where name ~ *"zelda"*;`
	if !contains(requestBody, expectedClause) {
		t.Fatalf("body=%q não contém a cláusula esperada %q", requestBody, expectedClause)
	}
	if contains(requestBody, "search") {
		t.Fatalf("body=%q não deveria conter 'search'", requestBody)
	}
}

func TestClientSearchGamesQuery(t *testing.T) {
	var requestBody string
	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth":
			_ = json.NewEncoder(w).Encode(accessToken{AccessToken: "token", ExpiresIn: 3600})
		case "/v4/games":
			requestPath = r.URL.Path
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ler body: %v", err)
			}
			requestBody = string(bodyBytes)
			_, _ = fmt.Fprint(w, `[{"id":1022,"name":"The Legend of Zelda","cover":{"id":86202,"url":"//cover.jpg"},"first_release_date":509328000,"summary":"Action RPG"}]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := New(Config{
		ClientID:     "client",
		ClientSecret: "secret",
		HTTPClient:   server.Client(),
		TokenURL:     server.URL + "/oauth",
		APIURL:       server.URL + "/v4",
	})

	games, err := client.SearchGames(context.Background(), "zelda")
	if err != nil {
		t.Fatalf("SearchGames: %v", err)
	}
	if len(games) != 1 || games[0].Name != "The Legend of Zelda" || games[0].Summary != "Action RPG" {
		t.Fatalf("games=%+v", games)
	}
	if requestPath != "/v4/games" {
		t.Fatalf("path=%s, esperava /v4/games", requestPath)
	}
	expectedClause := `fields id,name,cover.url,first_release_date,summary; search "zelda"; limit 20;`
	if requestBody != expectedClause {
		t.Fatalf("body=%q, esperava %q", requestBody, expectedClause)
	}
}


