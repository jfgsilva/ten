package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jfgsilva/ten/internal/config"
)

type openAIRequest struct {
	Model       string           `json:"model"`
	Messages    []config.Message `json:"messages"`
	Temperature *float64         `json:"temperature,omitempty"`
	Stream      bool             `json:"stream"`
}

type openAIResponse struct {
	Choices []struct {
		Message config.Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func callOpenAI(cfg config.ApiConfig, p config.Prompt) (config.Message, error) {
	req := openAIRequest{
		Model:       p.Model,
		Messages:    p.Messages,
		Temperature: p.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return config.Message{}, err
	}

	timeout := 30 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	client := &http.Client{Timeout: timeout}

	httpReq, err := http.NewRequest("POST", cfg.URL, bytes.NewReader(body))
	if err != nil {
		return config.Message{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	apiKey := cfg.GetAPIKey()
	switch p.API {
	case "azureopenai":
		httpReq.Header.Set("api-key", apiKey)
	default:
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if p.API == "cerebras" {
		httpReq.Header.Set("User-Agent", "CUSTOM_NAME/1.0")
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return config.Message{}, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return config.Message{}, err
	}

	var result openAIResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return config.Message{}, fmt.Errorf("decoding response: %w\nbody: %s", err, data)
	}
	if result.Error != nil {
		return config.Message{}, fmt.Errorf("api error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return config.Message{}, fmt.Errorf("empty choices in response: %s", data)
	}
	return config.Message{Role: "assistant", Content: result.Choices[0].Message.Content}, nil
}
