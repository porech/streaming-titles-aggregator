package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/porech/streaming-titles-aggregator/internal/config"
)

func init() {
	Register(icy{})
}

type icy struct{}

func (icy) Kind() string { return "icy" }

func (icy) Fetch(ctx context.Context, cfg config.StreamConfig) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.StreamURL, nil)
	if err != nil {
		return "", fmt.Errorf("icy: %w", err)
	}
	req.Header.Set("Icy-MetaData", "1")
	req.Header.Set("User-Agent", cfg.UserAgent)
	req.Header.Set("Accept-Encoding", "identity")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("icy: %w", err)
	}
	defer resp.Body.Close()

	metaIntStr := resp.Header.Get("icy-metaint")
	if metaIntStr == "" {
		return "", fmt.Errorf("icy: server did not send icy-metaint header")
	}

	var metaInt int
	if _, err := fmt.Sscanf(metaIntStr, "%d", &metaInt); err != nil || metaInt <= 0 {
		return "", fmt.Errorf("icy: invalid icy-metaint: %q", metaIntStr)
	}

	buf := make([]byte, metaInt)
	if _, err := io.ReadFull(resp.Body, buf); err != nil {
		return "", fmt.Errorf("icy: reading audio data: %w", err)
	}

	var metaLen [1]byte
	if _, err := io.ReadFull(resp.Body, metaLen[:]); err != nil {
		return "", fmt.Errorf("icy: reading metadata length: %w", err)
	}

	blocks := int(metaLen[0])
	metaBuf := make([]byte, blocks*16)
	if blocks > 0 {
		if _, err := io.ReadFull(resp.Body, metaBuf); err != nil {
			return "", fmt.Errorf("icy: reading metadata: %w", err)
		}
	}

	title := parseICYTitle(string(metaBuf))
	if title == "" {
		return "", fmt.Errorf("icy: no StreamTitle found in metadata")
	}

	return title, nil
}

func parseICYTitle(meta string) string {
	start := strings.Index(meta, "StreamTitle='")
	if start == -1 {
		return ""
	}
	start += len("StreamTitle='")
	end := strings.Index(meta[start:], "';")
	if end == -1 {
		return ""
	}
	return meta[start : start+end]
}
