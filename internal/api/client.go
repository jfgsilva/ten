package api

import (
	"fmt"

	"github.com/jfgsilva/ten/internal/config"
)

// Call dispatches to the appropriate backend based on prompt.API.
func Call(cfg config.ApiConfig, p config.Prompt) (config.Message, error) {
	// Apply default model from api config if not set in prompt
	if p.Model == "" {
		p.Model = cfg.DefaultModel
	}

	switch p.API {
	case "openai", "azureopenai", "mistral", "groq", "cerebras":
		return callOpenAI(cfg, p)
	case "anthropic":
		return callAnthropic(cfg, p)
	case "ollama":
		return callOllama(cfg, p)
	case "gemini":
		return callGemini(cfg, p)
	default:
		return config.Message{}, fmt.Errorf("unknown api %q", p.API)
	}
}
