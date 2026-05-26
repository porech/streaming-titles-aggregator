package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type StreamConfig struct {
	Kind     string `json:"kind"`
	BaseURL  string `json:"base_url"`
	SID      *int   `json:"sid,omitempty"`
	Mount    string `json:"mount,omitempty"`
	Port     *int   `json:"port,omitempty"`
	StreamID string `json:"stream_id,omitempty"`
}

type Config map[string]StreamConfig

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	for name, sc := range cfg {
		if err := sc.Validate(); err != nil {
			return nil, fmt.Errorf("config %q: %w", name, err)
		}
	}

	return cfg, nil
}

func (sc StreamConfig) Validate() error {
	if sc.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if sc.BaseURL == "" {
		return fmt.Errorf("base_url is required")
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
	default:
		return fmt.Errorf("unknown kind: %q", sc.Kind)
	}
	return nil
}
