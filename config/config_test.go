package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_envVars(t *testing.T) {
	t.Setenv("GROCY_URL", "http://grocy.local")
	t.Setenv("GROCY_API_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "http://grocy.local" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "http://grocy.local")
	}
	if cfg.APIKey != "secret" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "secret")
	}
}

func TestLoad_envVarsPartial(t *testing.T) {
	// Only URL set, not key — should not short-circuit, falls through to file search.
	// Since there's no config file in the temp HOME, it should return an error.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GROCY_URL", "http://grocy.local")
	t.Setenv("GROCY_API_KEY", "") // clear

	_, err := Load()
	if err == nil {
		t.Error("expected error when only URL is set (no key, no file), got nil")
	}
}

func TestLoad_configFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GROCY_URL", "")
	t.Setenv("GROCY_API_KEY", "")

	cfgDir := filepath.Join(tmp, ".config", "grocr")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	want := Config{
		BaseURL:              "https://my-grocy.example.com",
		APIKey:               "file-key-123",
		DisplayNameUserfield: "display",
		TLSSkipVerify:        true,
	}
	data, _ := json.Marshal(want)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != want.BaseURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, want.BaseURL)
	}
	if cfg.APIKey != want.APIKey {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, want.APIKey)
	}
	if cfg.DisplayNameUserfield != want.DisplayNameUserfield {
		t.Errorf("DisplayNameUserfield = %q, want %q", cfg.DisplayNameUserfield, want.DisplayNameUserfield)
	}
	if !cfg.TLSSkipVerify {
		t.Error("expected TLSSkipVerify = true")
	}
}

func TestLoad_noConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GROCY_URL", "")
	t.Setenv("GROCY_API_KEY", "")

	_, err := Load()
	if err == nil {
		t.Error("expected error when no config is available, got nil")
	}
}

func TestLoad_invalidJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GROCY_URL", "")
	t.Setenv("GROCY_API_KEY", "")

	cfgDir := filepath.Join(tmp, ".config", "grocr")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte("{invalid json"), 0o600)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoad_missingFields(t *testing.T) {
	// Valid JSON but missing base_url or api_key — should not match.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GROCY_URL", "")
	t.Setenv("GROCY_API_KEY", "")

	cfgDir := filepath.Join(tmp, ".config", "grocr")
	os.MkdirAll(cfgDir, 0o755)
	data, _ := json.Marshal(Config{BaseURL: "http://grocy.local"}) // no APIKey
	os.WriteFile(filepath.Join(cfgDir, "config.json"), data, 0o600)

	_, err := Load()
	if err == nil {
		t.Error("expected error when APIKey is missing, got nil")
	}
}
