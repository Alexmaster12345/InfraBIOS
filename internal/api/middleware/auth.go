package middleware

import (
	"net/http"
	"strings"

	"github.com/Alexmaster12345/infrabios/internal/api/response"
)

func BearerAuth(validTokens []string) func(http.Handler) http.Handler {
	set := make(map[string]struct{}, len(validTokens))
	for _, t := range validTokens {
		set[t] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				response.Error(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}
			if _, ok := set[token]; !ok {
				response.Error(w, http.StatusForbidden, "invalid or revoked token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func AgentAuth(agentToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				response.Error(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}
			if token != agentToken {
				response.Error(w, http.StatusForbidden, "invalid agent token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	parts := strings.SplitN(h, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
