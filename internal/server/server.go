package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/porech/streaming-titles-aggregator/internal/config"
	"github.com/porech/streaming-titles-aggregator/internal/source"
)

type reloadableConfig struct {
	mu   sync.RWMutex
	cfg  config.Config
	path string
}

func (rc *reloadableConfig) get(name string) (config.StreamConfig, bool) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	sc, ok := rc.cfg[name]
	return sc, ok
}

func (rc *reloadableConfig) reload() error {
	cfg, err := config.LoadConfig(rc.path)
	if err != nil {
		return err
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cfg = cfg
	return nil
}

func Run(addr, configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	rc := &reloadableConfig{cfg: cfg, path: configPath}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)
	go func() {
		for range sig {
			log.Println("reloading config...")
			if err := rc.reload(); err != nil {
				log.Printf("reload failed: %v", err)
			} else {
				log.Println("config reloaded")
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /title/", handleTitle(rc))

	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func handleTitle(rc *reloadableConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/title/"), ".json")
		if name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing stream name"})
			return
		}

		sc, ok := rc.get(name)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "stream not found: " + name})
			return
		}

		src, ok := source.Get(sc.Kind)
		if !ok {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unknown source kind: " + sc.Kind})
			return
		}

		title, err := src.Fetch(r.Context(), sc)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"name": name, "error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"title": title})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
