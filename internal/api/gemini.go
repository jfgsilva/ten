package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jfgsilva/ten/internal/config"
)

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func callGemini(cfg config.ApiConfig, p config.Prompt) (config.Message, error) {
	// Convert messages: system → prepend to first user message; assistant → "model"
	var systemParts []string
	var contents []geminiContent

	for _, m := range p.Messages {
		switch m.Role {
		case "system":
			systemParts = append(systemParts, m.Content)
		case "assistant":
			contents = append(contents, geminiContent{
				Role:  "model",
				Parts: []geminiPart{{Text: m.Content}},
			})
		default: // "user"
			text := m.Content
			if len(systemParts) > 0 {
				text = strings.Join(systemParts, "\n\n") + "\n\n" + text
				systemParts = nil
			}
			contents = append(contents, geminiContent{
				Role:  "user",
				Parts: []geminiPart{{Text: text}},
			})
		}
	}
	// If only system messages remain (no user message), attach them
	if len(systemParts) > 0 {
		contents = append(contents, geminiContent{
			Role:  "user",
			Parts: []geminiPart{{Text: strings.Join(systemParts, "\n\n")}},
		})
	}

	req := geminiRequest{Contents: contents}
	body, err := json.Marshal(req)
	if err != nil {
		return config.Message{}, err
	}

	timeout := 30 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Endpoint: baseURL/{model}:generateContent?key=apikey
	baseURL := cfg.URL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta/models"
	}
	endpoint := fmt.Sprintf("%s/%s:generateContent?key=%s", baseURL, p.Model, cfg.GetAPIKey())

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return config.Message{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return config.Message{}, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return config.Message{}, err
	}

	var result geminiResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return config.Message{}, fmt.Errorf("decoding response: %w\nbody: %s", err, data)
	}
	if result.Error != nil {
		return config.Message{}, fmt.Errorf("api error: %s", result.Error.Message)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return config.Message{}, fmt.Errorf("empty candidates in response: %s", data)
	}
	return config.Message{Role: "assistant", Content: result.Candidates[0].Content.Parts[0].Text}, nil
}
