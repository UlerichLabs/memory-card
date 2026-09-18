package config

import (
	"reflect"
	"testing"
)

func TestLoadCORS_Origens(t *testing.T) {
	for _, tc := range []struct {
		name, env string
		want      []string
		invalid   bool
	}{
		{"ausente", "", nil, false},
		{"multiplas", " http://localhost:5173, https://app.example.com ,", []string{"http://localhost:5173", "https://app.example.com"}, false},
		{"wildcard", "*", nil, true},
		{"subdominio wildcard", "https://*.example.com", nil, true},
		{"caminho", "http://localhost:5173/cadastro", nil, true},
		{"esquema invalido", "file://localhost", nil, true},
		{"sem host", "http://", nil, true},
		{"url invalida", "http://%", nil, true},
		{"credenciais", "https://user:pass@example.com", nil, true},
		{"query", "https://example.com?x=1", nil, true},
		{"fragmento", "https://example.com#x", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CORS_ALLOWED_ORIGINS", tc.env)
			got, err := LoadCORS()
			if (err != nil) != tc.invalid || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("origens = %v, erro = %v", got, err)
			}
		})
	}
}
