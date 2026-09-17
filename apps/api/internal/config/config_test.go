package config

import (
	"strings"
	"testing"
)

func TestLoad_ValoresValidos(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantConfig Config
	}{
		{
			name: "apenas DATABASE_URL definida usa porta padrao 8080",
			env: map[string]string{
				"DATABASE_URL":       "postgres://user:pass@localhost:5432/memorycard?sslmode=disable",
				"PORT":               "",
				"JWT_SECRET":         "",
				"IGDB_CLIENT_ID":     "",
				"IGDB_CLIENT_SECRET": "",
			},
			wantConfig: Config{
				DatabaseURL:      "postgres://user:pass@localhost:5432/memorycard?sslmode=disable",
				Port:             "8080",
				JWTSecret:        "",
				IGDBClientID:     "",
				IGDBClientSecret: "",
			},
		},
		{
			name: "todas as variaveis definidas com porta customizada",
			env: map[string]string{
				"DATABASE_URL":       "postgres://user:pass@localhost:5432/memorycard?sslmode=disable",
				"PORT":               "3000",
				"JWT_SECRET":         "supersecretkey123",
				"IGDB_CLIENT_ID":     "client_123",
				"IGDB_CLIENT_SECRET": "secret_456",
			},
			wantConfig: Config{
				DatabaseURL:      "postgres://user:pass@localhost:5432/memorycard?sslmode=disable",
				Port:             "3000",
				JWTSecret:        "supersecretkey123",
				IGDBClientID:     "client_123",
				IGDBClientSecret: "secret_456",
			},
		},
		{
			name: "limite inferior de porta valida 1",
			env: map[string]string{
				"DATABASE_URL": "postgres://user:pass@localhost:5432/memorycard",
				"PORT":         "1",
			},
			wantConfig: Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/memorycard",
				Port:        "1",
			},
		},
		{
			name: "limite superior de porta valida 65535",
			env: map[string]string{
				"DATABASE_URL": "postgres://user:pass@localhost:5432/memorycard",
				"PORT":         "65535",
			},
			wantConfig: Config{
				DatabaseURL: "postgres://user:pass@localhost:5432/memorycard",
				Port:        "65535",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			got, err := Load()
			if err != nil {
				t.Fatalf("Load() retornou erro inesperado: %v", err)
			}

			if got != tc.wantConfig {
				t.Errorf("Load() = %+v; esperado %+v", got, tc.wantConfig)
			}
		})
	}
}

func TestLoad_ValoresAusentes(t *testing.T) {
	t.Run("DATABASE_URL ausente retorna erro", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		t.Setenv("PORT", "8080")

		_, err := Load()
		if err == nil {
			t.Fatal("Load() deveria falhar quando DATABASE_URL estiver ausente")
		}

		if !strings.Contains(err.Error(), "DATABASE_URL é obrigatória") {
			t.Errorf("mensagem de erro = %q; esperava conter %q", err.Error(), "DATABASE_URL é obrigatória")
		}
	})
}

func TestLoad_ValoresInvalidos(t *testing.T) {
	tests := []struct {
		name          string
		port          string
		wantErrSubstr string
	}{
		{
			name:          "porta com caracteres nao numericos",
			port:          "abc",
			wantErrSubstr: "interpretar PORT",
		},
		{
			name:          "porta com espacos",
			port:          "80 80",
			wantErrSubstr: "interpretar PORT",
		},
		{
			name:          "porta abaixo do limite minimo zero",
			port:          "0",
			wantErrSubstr: "PORT deve estar entre 1 e 65535",
		},
		{
			name:          "porta negativa",
			port:          "-8080",
			wantErrSubstr: "PORT deve estar entre 1 e 65535",
		},
		{
			name:          "porta acima do limite maximo",
			port:          "65536",
			wantErrSubstr: "PORT deve estar entre 1 e 65535",
		},
		{
			name:          "porta muito alta",
			port:          "99999",
			wantErrSubstr: "PORT deve estar entre 1 e 65535",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://localhost:5432/memorycard")
			t.Setenv("PORT", tc.port)

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() deveria falhar para PORT=%q", tc.port)
			}

			if !strings.Contains(err.Error(), tc.wantErrSubstr) {
				t.Errorf("mensagem de erro = %q; esperava conter %q", err.Error(), tc.wantErrSubstr)
			}
		})
	}
}
