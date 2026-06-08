package config

import (
	"os"
	"strings"
)

type Config struct {
	ListenAddr  string
	DatabaseURL string
	APITokens   []string
	AgentToken  string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		ListenAddr:  getEnv("INFRABIOS_ADDR", ":8080"),
		DatabaseURL: getEnv("INFRABIOS_DATABASE_URL", "postgres://infrabios:infrabios@localhost:5432/infrabios?sslmode=disable"),
		APITokens:   splitTokens(getEnv("INFRABIOS_API_TOKENS", "changeme")),
		AgentToken:  getEnv("INFRABIOS_AGENT_TOKEN", "agent-changeme"),
		LogLevel:    getEnv("INFRABIOS_LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitTokens(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
