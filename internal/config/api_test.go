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
