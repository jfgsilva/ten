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

type anthropicRequest struct {
	Model       string           `json:"model"`
	Messages    []config.Message `json:"messages"`
	Temperature *float64         `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens"`
	Stream      bool             `json:"stream"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// mergeMessages converts system→user and merges consecutive same-role messages.
func mergeMessages(messages []config.Message) []config.Message {
	merged := make([]config.Message, 0, len(messages))
	for _, m := range messages {
		msg := config.Message{Role: m.Role, Content: m.Content}
		if msg.Role == "system" {
			msg.Role = "user"
		}
		if len(merged) > 0 && merged[len(merged)-1].Role == msg.Role {
			merged[len(merged)-1].Content += "\n\n" + msg.Content
		} else {
			merged = append(merged, msg)
		}
	}
	return merged
}

func callAnthropic(cfg config.ApiConfig, p config.Prompt) (config.Message, error) {
	req := anthropicRequest{
		Model:       p.Model,
		Messages:    mergeMessages(p.Messages),
		Temperature: p.Temperature,
		MaxTokens:   4096,
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
	httpReq.Header.Set("x-api-key", cfg.GetAPIKey())
	version := cfg.Version
	if version == "" {
		version = "2023-06-01"
	}
	httpReq.Header.Set("anthropic-version", version)

	resp, err := client.Do(httpReq)
	if err != nil {
		return config.Message{}, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return config.Message{}, err
	}

	var result anthropicResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return config.Message{}, fmt.Errorf("decoding response: %w\nbody: %s", err, data)
	}
	if result.Error != nil {
		return config.Message{}, fmt.Errorf("api error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return config.Message{}, fmt.Errorf("empty content in response: %s", data)
	}
	return config.Message{Role: "assistant", Content: result.Content[0].Text}, nil
}
