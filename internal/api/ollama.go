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

type ollamaRequest struct {
	Model       string           `json:"model"`
	Messages    []config.Message `json:"messages"`
	Temperature *float64         `json:"temperature,omitempty"`
	Stream      bool             `json:"stream"`
}

type ollamaResponse struct {
	Message config.Message `json:"message"`
}

func callOllama(cfg config.ApiConfig, p config.Prompt) (config.Message, error) {
	req := ollamaRequest{
		Model:       p.Model,
		Messages:    p.Messages,
		Temperature: p.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return config.Message{}, err
	}

	timeout := 180 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	client := &http.Client{Timeout: timeout}

	httpReq, err := http.NewRequest("POST", cfg.URL, bytes.NewReader(body))
	if err != nil {
		return config.Message{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return config.Message{}, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return config.Message{}, err
	}

	var result ollamaResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return config.Message{}, fmt.Errorf("decoding response: %w\nbody: %s", err, data)
	}
	if result.Message.Content == "" {
		return config.Message{}, fmt.Errorf("empty message in response: %s", data)
	}
	return config.Message{Role: "assistant", Content: result.Message.Content}, nil
}
