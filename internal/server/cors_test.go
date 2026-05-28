package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/porech/streaming-titles-aggregator/internal/config"
)

func staticCORS(c *config.CORSConfig) func() *config.CORSConfig {
	return func() *config.CORSConfig { return c }
}

func nextOK() (http.Handler, *bool) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	return h, &called
}

func TestCORS_NilConfigIsNoOp(t *testing.T) {
	next, called := nextOK()
	h := corsMiddleware(staticCORS(nil), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !*called {
		t.Fatal("expected next to be called")
	}
	for _, k := range []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
		"Access-Control-Expose-Headers",
		"Vary",
	} {
		if v := rec.Header().Get(k); v != "" {
			t.Errorf("expected no %s header, got %q", k, v)
		}
	}
}

func TestCORS_OriginInList(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"https://a", "https://b"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}
	next, called := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	req.Header.Set("Origin", "https://b")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !*called {
		t.Fatal("expected next to be called")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://b" {
		t.Errorf("Allow-Origin = %q, want https://b", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORS_OriginNotInList(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"https://a"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}
	next, called := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	req.Header.Set("Origin", "https://evil")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !*called {
		t.Fatal("expected next to be called even when origin not allowed")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, want empty", got)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"https://a"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}
	next, called := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !*called {
		t.Fatal("expected next to be called")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, want empty", got)
	}
}

func TestCORS_WildcardWithoutCredentials(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}
	next, _ := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	req.Header.Set("Origin", "https://anything")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Allow-Origin = %q, want *", got)
	}
	if got := rec.Header().Get("Vary"); got != "" {
		t.Errorf("Vary = %q, want empty", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("Allow-Credentials = %q, want empty", got)
	}
}

func TestCORS_WildcardWithCredentialsEchoesOrigin(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowCredentials: true,
	}
	next, _ := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	req.Header.Set("Origin", "https://anything")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://anything" {
		t.Errorf("Allow-Origin = %q, want https://anything", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q, want true", got)
	}
}

func TestCORS_WildcardWithCredentialsNoOriginHeader(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowCredentials: true,
	}
	next, _ := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, want empty", got)
	}
}

func TestCORS_PreflightFull(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"https://a"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "X-Token"},
		MaxAge:         3600,
	}
	next, called := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodOptions, "/title/x.json", nil)
	req.Header.Set("Origin", "https://a")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if *called {
		t.Fatal("preflight must NOT invoke next")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, OPTIONS" {
		t.Errorf("Allow-Methods = %q, want %q", got, "GET, OPTIONS")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, X-Token" {
		t.Errorf("Allow-Headers = %q, want %q", got, "Content-Type, X-Token")
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got != "3600" {
		t.Errorf("Max-Age = %q, want 3600", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://a" {
		t.Errorf("Allow-Origin = %q, want https://a", got)
	}
}

func TestCORS_PreflightMaxAgeZeroOmitted(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"https://a"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		MaxAge:         0,
	}
	next, _ := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodOptions, "/title/x.json", nil)
	req.Header.Set("Origin", "https://a")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Max-Age"); got != "" {
		t.Errorf("Max-Age = %q, want empty when MaxAge=0", got)
	}
}

func TestCORS_OptionsWithoutPreflightHeaderFallsThrough(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"https://a"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}
	next, called := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodOptions, "/title/x.json", nil)
	req.Header.Set("Origin", "https://a")
	// no Access-Control-Request-Method => not a preflight
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !*called {
		t.Fatal("non-preflight OPTIONS should reach next")
	}
}

func TestCORS_ExposeHeaders(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"https://a"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		ExposedHeaders: []string{"X-Foo", "X-Bar"},
	}
	next, _ := nextOK()
	h := corsMiddleware(staticCORS(cfg), next)

	req := httptest.NewRequest(http.MethodGet, "/title/x.json", nil)
	req.Header.Set("Origin", "https://a")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Expose-Headers"); !strings.EqualFold(got, "X-Foo, X-Bar") {
		t.Errorf("Expose-Headers = %q, want %q", got, "X-Foo, X-Bar")
	}
}
