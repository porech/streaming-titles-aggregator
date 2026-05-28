package server

import (
	"context"
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
	sync.RWMutex
	cfg  config.Config
	path string
}

type Server struct {
	rc   *reloadableConfig
	addr string
	srv  *http.Server
	srvMu sync.Mutex
}

func Run(configPath string) error {
	cfg, addr, _, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	s := &Server{
		rc:   &reloadableConfig{cfg: cfg, path: configPath},
		addr: addr,
	}

	go s.serve()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)
	for range sig {
		s.reload()
	}

	return nil
}

func (s *Server) serve() {
	for {
		s.srvMu.Lock()
		addr := s.addr
		mux := http.NewServeMux()
		mux.HandleFunc("GET /title/", handleTitle(s.rc))
		srv := &http.Server{Addr: addr, Handler: mux}
		s.srv = srv
		s.srvMu.Unlock()

		log.Printf("listening on %s", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}
}

func (s *Server) reload() {
	cfg, addr, _, err := config.LoadConfig(s.rc.path)
	if err != nil {
		log.Printf("reload failed: %v", err)
		return
	}

	s.rc.Lock()
	s.rc.cfg = cfg
	s.rc.Unlock()

	s.srvMu.Lock()
	oldAddr := s.addr
	s.addr = addr
	srv := s.srv
	s.srvMu.Unlock()

	if addr == oldAddr {
		log.Println("config reloaded")
		return
	}

	log.Printf("address changed from %s to %s, restarting...", oldAddr, addr)
	if srv != nil {
		go srv.Shutdown(context.Background())
	}
}

func handleTitle(rc *reloadableConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/title/"), ".json")
		if name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing stream name"})
			return
		}

		rc.RLock()
		sc, ok := rc.cfg[name]
		rc.RUnlock()
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
