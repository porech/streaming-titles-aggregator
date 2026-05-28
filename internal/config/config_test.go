package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadConfig_CORSAbsent(t *testing.T) {
	path := writeConfig(t, `{"streams": {}}`)
	_, _, cors, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cors != nil {
		t.Fatalf("expected nil CORS, got %+v", cors)
	}
}

func TestLoadConfig_CORSParsed(t *testing.T) {
	path := writeConfig(t, `{
		"cors": {
			"allowed_origins": ["https://a", "https://b"],
			"allowed_methods": ["GET", "POST"],
			"allowed_headers": ["Content-Type"],
			"exposed_headers": ["X-Foo"],
			"allow_credentials": true,
			"max_age": 3600
		},
		"streams": {}
	}`)
	_, _, cors, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cors == nil {
		t.Fatal("expected non-nil CORS")
	}
	want := &CORSConfig{
		AllowedOrigins:   []string{"https://a", "https://b"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Foo"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	if !reflect.DeepEqual(cors, want) {
		t.Fatalf("CORS mismatch\n got: %+v\nwant: %+v", cors, want)
	}
}

func TestLoadConfig_CORSDefaultMethods(t *testing.T) {
	path := writeConfig(t, `{
		"cors": {"allowed_origins": ["*"]},
		"streams": {}
	}`)
	_, _, cors, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cors == nil {
		t.Fatal("expected non-nil CORS")
	}
	want := []string{"GET", "OPTIONS"}
	if !reflect.DeepEqual(cors.AllowedMethods, want) {
		t.Fatalf("AllowedMethods = %v, want %v", cors.AllowedMethods, want)
	}
}

func TestLoadConfig_CORSRequiresOrigins(t *testing.T) {
	path := writeConfig(t, `{
		"cors": {"max_age": 60},
		"streams": {}
	}`)
	_, _, _, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "allowed_origins") {
		t.Fatalf("error %q does not mention allowed_origins", err.Error())
	}
}

func TestLoadConfig_CORSRejectsNegativeMaxAge(t *testing.T) {
	path := writeConfig(t, `{
		"cors": {"allowed_origins": ["*"], "max_age": -1},
		"streams": {}
	}`)
	_, _, _, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "max_age") {
		t.Fatalf("error %q does not mention max_age", err.Error())
	}
}
