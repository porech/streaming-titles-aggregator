package source

import (
	"context"
	"fmt"

	"github.com/porech/streaming-titles-aggregator/internal/config"
)

func init() {
	Register(shoutcast{})
}

type shoutcast struct{}

func (shoutcast) Kind() string { return "shoutcast" }

func (shoutcast) Fetch(ctx context.Context, cfg config.StreamConfig) (string, error) {
	url := fmt.Sprintf("%s/stats?sid=%d&json=1", cfg.BaseURL, *cfg.SID)

	var resp struct {
		SongTitle string `json:"songtitle"`
	}
	if err := fetchJSON(ctx, url, cfg.UserAgent, &resp); err != nil {
		return "", fmt.Errorf("shoutcast: %w", err)
	}

	return resp.SongTitle, nil
}
