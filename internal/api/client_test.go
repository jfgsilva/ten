package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jfgsilva/ten/internal/config"
)

func openAIHandler(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": content}},
			},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}
}

func basePrompt(api string) config.Prompt {
	return config.Prompt{
		API:   api,
		Model: "test-model",
		Messages: []config.Message{
			{Role: "user", Content: "hi"},
		},
	}
}

func TestCallDispatchOpenAI(t *testing.T) {
	srv := httptest.NewServer(openAIHandler("openai response"))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "k", URL: srv.URL}
	msg, err := Call(cfg, basePrompt("openai"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "openai response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallDispatchAzureOpenAI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") == "" {
			t.Error("missing api-key header")
		}
		openAIHandler("azure response")(w, r)
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "azurekey", URL: srv.URL}
	msg, err := Call(cfg, basePrompt("azureopenai"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "azure response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallDispatchMistral(t *testing.T) {
	srv := httptest.NewServer(openAIHandler("mistral response"))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "k", URL: srv.URL}
	msg, err := Call(cfg, basePrompt("mistral"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "mistral response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallDispatchGroq(t *testing.T) {
	srv := httptest.NewServer(openAIHandler("groq response"))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "k", URL: srv.URL}
	msg, err := Call(cfg, basePrompt("groq"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "groq response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallDispatchCerebras(t *testing.T) {
	srv := httptest.NewServer(openAIHandler("cerebras response"))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "k", URL: srv.URL}
	msg, err := Call(cfg, basePrompt("cerebras"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "cerebras response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallDispatchAnthropic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": "anthropic response"},
			},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "k", URL: srv.URL, Version: "2023-06-01"}
	msg, err := Call(cfg, basePrompt("anthropic"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "anthropic response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallDispatchOllama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"message": map[string]string{"role": "assistant", "content": "ollama response"},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{URL: srv.URL}
	msg, err := Call(cfg, basePrompt("ollama"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "ollama response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallDispatchGemini(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "key=gemini-key") {
			t.Errorf("missing key param: %s", r.URL.RawQuery)
		}
		resp := map[string]any{
			"candidates": []map[string]any{
				{"content": map[string]any{"parts": []map[string]string{{"text": "gemini response"}}}},
			},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "gemini-key", URL: srv.URL}
	msg, err := Call(cfg, basePrompt("gemini"))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "gemini response" {
		t.Errorf("got %q", msg.Content)
	}
}

func TestCallUnknownAPIReturnsError(t *testing.T) {
	cfg := config.ApiConfig{}
	_, err := Call(cfg, basePrompt("unknownapi"))
	if err == nil {
		t.Fatal("expected error for unknown api")
	}
	if !strings.Contains(err.Error(), "unknown api") {
		t.Errorf("expected 'unknown api' in error, got: %v", err)
	}
}

func TestCallAppliesDefaultModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		if body["model"] != "gpt-4" {
			t.Errorf("want model=gpt-4, got %v", body["model"])
		}
		openAIHandler("ok")(w, r)
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "k", URL: srv.URL, DefaultModel: "gpt-4"}
	p := config.Prompt{
		API:      "openai",
		Model:    "", // empty — should fall back to cfg.DefaultModel
		Messages: []config.Message{{Role: "user", Content: "hi"}},
	}
	if _, err := Call(cfg, p); err != nil {
		t.Fatal(err)
	}
}
