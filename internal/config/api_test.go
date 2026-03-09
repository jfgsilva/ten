package config

import (
	"os"
	"testing"
)

func TestGetAPIKeyDirect(t *testing.T) {
	cfg := ApiConfig{APIKey: "testkey"}
	if got := cfg.GetAPIKey(); got != "testkey" {
		t.Fatalf("want testkey, got %s", got)
	}
}

func TestGetAPIKeyCommand(t *testing.T) {
	cfg := ApiConfig{APIKeyCommand: "echo secretkey"}
	if got := cfg.GetAPIKey(); got != "secretkey" {
		t.Fatalf("want secretkey, got %s", got)
	}
}

func TestGetAPIKeyEmpty(t *testing.T) {
	cfg := ApiConfig{}
	if got := cfg.GetAPIKey(); got != "" {
		t.Fatalf("want empty, got %s", got)
	}
}

func TestGetAPIConfigHappy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)
	if err := GenerateDefaultAPIConfigs(); err != nil {
		t.Fatal(err)
	}
	cfg := GetAPIConfig("openai")
	if cfg.URL == "" {
		t.Error("want non-empty URL")
	}
	if cfg.DefaultModel != "gpt-4" {
		t.Errorf("want gpt-4, got %s", cfg.DefaultModel)
	}
}

func TestGetAPIConfigUnknownPanics(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)
	if err := GenerateDefaultAPIConfigs(); err != nil {
		t.Fatal(err)
	}
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		GetAPIConfig("nonexistent")
	}()
	if !panicked {
		t.Error("expected panic for unknown api")
	}
}

func TestGenerateDefaultAPIConfigs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	if err := GenerateDefaultAPIConfigs(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(APIConfigPath()); err != nil {
		t.Fatal("api config file not created:", err)
	}

	configs, err := LoadAPIConfigs()
	if err != nil {
		t.Fatal(err)
	}
	for _, provider := range []string{"ollama", "openai", "anthropic", "gemini"} {
		if _, ok := configs[provider]; !ok {
			t.Errorf("missing provider %q in generated config", provider)
		}
	}
}
