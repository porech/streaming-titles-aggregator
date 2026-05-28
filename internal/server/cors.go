package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/porech/streaming-titles-aggregator/internal/config"
)

func corsMiddleware(get func() *config.CORSConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := get()
		if cfg == nil {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		h := w.Header()

		allowOrigin, addVary := resolveAllowOrigin(cfg, origin)
		if allowOrigin != "" {
			h.Set("Access-Control-Allow-Origin", allowOrigin)
		}
		if addVary {
			h.Add("Vary", "Origin")
		}

		if cfg.AllowCredentials {
			h.Set("Access-Control-Allow-Credentials", "true")
		}
		if len(cfg.ExposedHeaders) > 0 {
			h.Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
		}

		if isPreflight(r) {
			if len(cfg.AllowedMethods) > 0 {
				h.Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			}
			if len(cfg.AllowedHeaders) > 0 {
				h.Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}
			if cfg.MaxAge > 0 {
				h.Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func resolveAllowOrigin(cfg *config.CORSConfig, origin string) (value string, vary bool) {
	if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
		if !cfg.AllowCredentials {
			return "*", false
		}
		if origin == "" {
			return "", false
		}
		return origin, true
	}

	if origin == "" {
		return "", false
	}
	for _, o := range cfg.AllowedOrigins {
		if o == origin {
			return origin, true
		}
	}
	return "", false
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}
