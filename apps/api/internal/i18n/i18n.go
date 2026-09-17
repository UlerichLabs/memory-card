package i18n

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed *.json
var localesFS embed.FS

var (
	ptBRMessages = map[string]string{}
	enMessages   = map[string]string{}
)

func init() {
	loadLocale("pt-BR.json", ptBRMessages)
	loadLocale("en.json", enMessages)
}

func loadLocale(filename string, target map[string]string) {
	data, err := localesFS.ReadFile(filename)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &target)
}

func T(lang, key string) string {
	normalized := strings.ToLower(strings.TrimSpace(lang))
	if strings.HasPrefix(normalized, "en") {
		if val, ok := enMessages[key]; ok {
			return val
		}
	}

	if val, ok := ptBRMessages[key]; ok {
		return val
	}

	return key
}
