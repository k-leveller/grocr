package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	BaseURL              string `json:"base_url"`
	APIKey               string `json:"api_key"`
	DisplayNameUserfield string `json:"display_name_userfield"`
	TLSSkipVerify        bool   `json:"tls_skip_verify"`
	Language             string `json:"language"`
}

func Load() (*Config, error) {
	// Priority: env vars > config file
	baseURL := os.Getenv("GROCY_URL")
	apiKey := os.Getenv("GROCY_API_KEY")

	if baseURL != "" && apiKey != "" {
		return &Config{BaseURL: baseURL, APIKey: apiKey}, nil
	}

	paths := []string{
		filepath.Join(os.Getenv("HOME"), ".config", "grocr", "config.json"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		if cfg.BaseURL != "" && cfg.APIKey != "" {
			return &cfg, nil
		}
	}

	return nil, fmt.Errorf("no Grocy config found; set GROCY_URL and GROCY_API_KEY env vars or create ~/.config/grocr/config.json")
}
