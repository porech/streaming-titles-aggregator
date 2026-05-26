package source

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/porech/streaming-titles-aggregator/internal/config"
)

func init() {
	Register(indiplay{})
}

type indiplay struct{}

func (indiplay) Kind() string { return "indiplay" }

func (indiplay) Fetch(ctx context.Context, cfg config.StreamConfig) (string, error) {
	u := cfg.BaseURL
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	u += "metadata/title"

	q := url.Values{}
	if cfg.StreamID != "" {
		q.Set("id", cfg.StreamID)
	} else {
		q.Set("mount", cfg.Mount)
		if cfg.Port != nil {
			q.Set("port", fmt.Sprintf("%d", *cfg.Port))
		}
	}

	fullURL := u + "?" + q.Encode()

	var resp struct {
		StreamTitle string `json:"streamTitle"`
	}
	if err := fetchJSON(ctx, fullURL, &resp); err != nil {
		return "", fmt.Errorf("indiplay: %w", err)
	}

	return resp.StreamTitle, nil
}
