package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const defaultUserAgent = "streaming-titles-aggregator/1.0"

type StreamConfig struct {
	Kind      string `json:"kind"`
	BaseURL   string `json:"base_url,omitempty"`
	SID       *int   `json:"sid,omitempty"`
	Mount     string `json:"mount,omitempty"`
	Port      *int   `json:"port,omitempty"`
	StreamID  string `json:"stream_id,omitempty"`
	StreamURL string `json:"stream_url,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

type Config map[string]StreamConfig

type configFile struct {
	UserAgent string                  `json:"user_agent"`
	Streams   map[string]StreamConfig `json:"streams"`
}

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	defer f.Close()

	var raw configFile
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	cfg := make(Config, len(raw.Streams))
	for name, sc := range raw.Streams {
		if sc.UserAgent == "" {
			sc.UserAgent = raw.UserAgent
		}
		if sc.UserAgent == "" {
			sc.UserAgent = defaultUserAgent
		}

		if err := sc.Validate(); err != nil {
			return nil, fmt.Errorf("config %q: %w", name, err)
		}
		cfg[name] = sc
	}

	return cfg, nil
}

func (sc StreamConfig) Validate() error {
	if sc.Kind == "" {
		return fmt.Errorf("kind is required")
	}

	switch sc.Kind {
	case "icy":
		if sc.StreamURL == "" {
			return fmt.Errorf("stream_url is required for kind %q", sc.Kind)
		}
	default:
		if sc.BaseURL == "" {
			return fmt.Errorf("base_url is required for kind %q", sc.Kind)
		}
	}

	switch sc.Kind {
	case "shoutcast":
		if sc.SID == nil {
			return fmt.Errorf("sid is required for kind %q", sc.Kind)
		}
	case "icecast":
		if sc.Mount == "" {
			return fmt.Errorf("mount is required for kind %q", sc.Kind)
		}
	case "indiplay":
		if sc.StreamID == "" && sc.Mount == "" {
			return fmt.Errorf("stream_id or mount is required for kind %q", sc.Kind)
		}
	case "icy":
		// url already validated above
	default:
		return fmt.Errorf("unknown kind: %q", sc.Kind)
	}
	return nil
}
