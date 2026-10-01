package config

import (
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestLoadAuth_Cenarios(t *testing.T) {
	for _, tc := range []struct {
		name, secret, access, refresh string
		valid                         bool
	}{
		{"valido", uuid.NewString(), "15m", "168h", true},
		{"segredo ausente", "", "15m", "168h", false},
		{"access ausente", uuid.NewString(), "", "168h", false},
		{"refresh ausente", uuid.NewString(), "15m", "", false},
		{"access zero", uuid.NewString(), "0s", "168h", false},
		{"refresh curto", uuid.NewString(), "15m", "1m", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tc.secret)
			t.Setenv("JWT_ACCESS_TTL", tc.access)
			t.Setenv("JWT_REFRESH_TTL", tc.refresh)
			cfg, err := LoadAuth()
			if (err == nil) != tc.valid {
				t.Fatalf("erro inesperado: %v", err)
			}
			if tc.valid && (cfg.AccessTTL != 15*time.Minute || cfg.RefreshTTL != 168*time.Hour || cfg.Secret != tc.secret) {
				t.Fatal("config incorreta")
			}
		})
	}
}
