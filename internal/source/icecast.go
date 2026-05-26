package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/porech/streaming-titles-aggregator/internal/config"
)

func init() {
	Register(icecast{})
}

type icecast struct{}

func (icecast) Kind() string { return "icecast" }

type icecastResponse struct {
	Icestats struct {
		Source json.RawMessage `json:"source"`
	} `json:"icestats"`
}

type sourceInfo struct {
	ListenURL string `json:"listenurl"`
	Title     string `json:"title"`
}

func (icecast) Fetch(ctx context.Context, cfg config.StreamConfig) (string, error) {
	u := cfg.BaseURL
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	u += "status-json.xsl"

	var raw icecastResponse
	if err := fetchJSON(ctx, u, cfg.UserAgent, &raw); err != nil {
		return "", fmt.Errorf("icecast: %w", err)
	}

	sources, err := parseSources(raw.Icestats.Source)
	if err != nil {
		return "", fmt.Errorf("icecast: %w", err)
	}

	mount := cfg.Mount
	for _, src := range sources {
		if matchMount(src.ListenURL, mount) {
			return src.Title, nil
		}
	}

	return "", fmt.Errorf("icecast: no source found for mount %q", mount)
}

func parseSources(raw json.RawMessage) ([]sourceInfo, error) {
	var sources []sourceInfo
	if err := json.Unmarshal(raw, &sources); err == nil {
		return sources, nil
	}

	var single sourceInfo
	if err := json.Unmarshal(raw, &single); err != nil {
		return nil, fmt.Errorf("cannot parse source field: %w", err)
	}

	return []sourceInfo{single}, nil
}

func matchMount(listenURL, mount string) bool {
	parsed, err := url.Parse(listenURL)
	if err != nil {
		return strings.Contains(listenURL, mount)
	}
	return parsed.Path == mount
}
