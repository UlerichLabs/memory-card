package i18n

import (
	"testing"
)

func TestT_Traducao(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		key      string
		expected string
	}{
		{
			name:     "pt-BR default",
			lang:     "",
			key:      "auth.register.email_taken",
			expected: "Este email já está em uso.",
		},
		{
			name:     "pt-BR explicito",
			lang:     "pt-BR",
			key:      "auth.register.weak_password",
			expected: "A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.",
		},
		{
			name:     "en explicito",
			lang:     "en-US",
			key:      "auth.register.invalid_email",
			expected: "Please enter a valid email address.",
		},
		{
			name:     "chave inexistente retorna a propria chave",
			lang:     "en",
			key:      "chave.inexistente",
			expected: "chave.inexistente",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := T(tc.lang, tc.key)
			if got != tc.expected {
				t.Errorf("T(%q, %q) = %q, esperado %q", tc.lang, tc.key, got, tc.expected)
			}
		})
	}
}
