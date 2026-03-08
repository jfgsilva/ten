package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

// ApiConfig holds configuration for a single API provider.
type ApiConfig struct {
	APIKey         string `toml:"api_key"`
	APIKeyCommand  string `toml:"api_key_command"`
	URL            string `toml:"url"`
	DefaultModel   string `toml:"default_model"`
	Version        string `toml:"version"`         // Anthropic only
	TimeoutSeconds int    `toml:"timeout_seconds"` // 0 means use default (30s)
}

// GetAPIKey returns the API key, running APIKeyCommand if needed.
func (c ApiConfig) GetAPIKey() string {
	if c.APIKey != "" {
		return c.APIKey
	}
	if c.APIKeyCommand != "" {
		out, err := exec.Command("sh", "-c", c.APIKeyCommand).Output() // #nosec G204
		if err != nil {
			panic(fmt.Sprintf("failed to run api_key_command %q: %v", c.APIKeyCommand, err))
		}
		return strings.TrimSpace(string(out))
	}
	return ""
}

// LoadAPIConfigs reads .api_configs.toml and returns the map.
func LoadAPIConfigs() (map[string]ApiConfig, error) {
	var configs map[string]ApiConfig
	if _, err := toml.DecodeFile(APIConfigPath(), &configs); err != nil {
		return nil, fmt.Errorf("reading %s: %w", APIConfigPath(), err)
	}
	return configs, nil
}

// GetAPIConfig returns a single provider's config, panicking if not found.
func GetAPIConfig(api string) ApiConfig {
	configs, err := LoadAPIConfigs()
	if err != nil {
		panic(err.Error())
	}
	cfg, ok := configs[api]
	if !ok {
		keys := make([]string, 0, len(configs))
		for k := range configs {
			keys = append(keys, k)
		}
		panic(fmt.Sprintf("api %q not found, available: %v", api, keys))
	}
	return cfg
}

// defaultAPIConfigs returns the stub configs written on first run.
func defaultAPIConfigs() map[string]ApiConfig {
	return map[string]ApiConfig{
		"ollama": {
			URL:            "http://localhost:11434/api/chat",
			DefaultModel:   "phi3",
			TimeoutSeconds: 180,
		},
		"openai": {
			URL:          "https://api.openai.com/v1/chat/completions",
			DefaultModel: "gpt-4",
		},
		"azureopenai": {
			URL:          "https://your-azure-endpoint.azure.com/openai/deployments/your-deployment-id/chat/completions?api-version=2024-06-01",
			DefaultModel: "gpt-4o",
		},
		"mistral": {
			URL:          "https://api.mistral.ai/v1/chat/completions",
			DefaultModel: "mistral-medium",
		},
		"groq": {
			URL:          "https://api.groq.com/openai/v1/chat/completions",
			DefaultModel: "llama3-70b-8192",
		},
		"anthropic": {
			URL:          "https://api.anthropic.com/v1/messages",
			DefaultModel: "claude-3-opus-20240229",
			Version:      "2023-06-01",
		},
		"cerebras": {
			URL:          "https://api.cerebras.ai/v1/chat/completions",
			DefaultModel: "llama3.1-70b",
		},
		"gemini": {
			URL:          "https://generativelanguage.googleapis.com/v1beta/models",
			DefaultModel: "gemini-1.5-flash",
		},
	}
}

// GenerateDefaultAPIConfigs writes a stub .api_configs.toml.
func GenerateDefaultAPIConfigs() error {
	if err := os.MkdirAll(ConfigDir(), 0o750); err != nil {
		return err
	}

	f, err := os.Create(APIConfigPath())
	if err != nil {
		return err
	}
	defer f.Close()

	header := "# API config file — use api_key or api_key_command fields to set credentials\n" +
		"# Example: api_key_command = \"op read op://vault/item/field\"\n\n"
	if _, err := f.WriteString(header); err != nil {
		return err
	}

	return toml.NewEncoder(f).Encode(defaultAPIConfigs()) // #nosec G117 -- APIKey field is intentional in config file format
}
