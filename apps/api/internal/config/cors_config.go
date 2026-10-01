package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// LoadCORS lê as origens HTTP permitidas, separadas por vírgula.
func LoadCORS() ([]string, error) {
	var origins []string
	for _, value := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		origin := strings.TrimSpace(value)
		if origin == "" {
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil {
			return nil, fmt.Errorf("interpretar CORS_ALLOWED_ORIGINS: %w", err)
		}
		if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || strings.Contains(parsed.Host, "*") || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS deve conter apenas origens HTTP(S), sem caminho ou wildcard")
		}
		origins = append(origins, origin)
	}
	return origins, nil
}
