package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/porech/streaming-titles-aggregator/internal/config"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type Source interface {
	Kind() string
	Fetch(context.Context, config.StreamConfig) (string, error)
}

var sources = make(map[string]Source)

func Register(s Source) {
	k := s.Kind()
	if _, exists := sources[k]; exists {
		panic(fmt.Sprintf("source %q already registered", k))
	}
	sources[k] = s
}

func Get(kind string) (Source, bool) {
	s, ok := sources[kind]
	return s, ok
}

func fetchJSON(ctx context.Context, url, userAgent string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}
