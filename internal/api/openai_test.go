package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jfgsilva/ten/internal/config"
)

func TestCallOpenAI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer testkey" {
			t.Errorf("missing/wrong Authorization header: %s", r.Header.Get("Authorization"))
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "hello from openai"}},
			},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "testkey", URL: srv.URL}
	p := config.Prompt{
		API:   "openai",
		Model: "gpt-4",
		Messages: []config.Message{
			{Role: "user", Content: "hi"},
		},
	}

	msg, err := callOpenAI(cfg, p)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "hello from openai" {
		t.Errorf("want 'hello from openai', got %q", msg.Content)
	}
}

func TestCallCerebrasUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "CUSTOM_NAME/1.0" {
			t.Errorf("missing User-Agent header: %s", r.Header.Get("User-Agent"))
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "key", URL: srv.URL}
	p := config.Prompt{
		API:   "cerebras",
		Model: "llama3",
		Messages: []config.Message{{Role: "user", Content: "hi"}},
	}
	if _, err := callOpenAI(cfg, p); err != nil {
		t.Fatal(err)
	}
}

func TestCallAzureOpenAI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") != "azurekey" {
			t.Errorf("missing api-key header: %s", r.Header.Get("api-key"))
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "azure response"}},
			},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "azurekey", URL: srv.URL}
	p := config.Prompt{
		API:   "azureopenai",
		Model: "gpt-4o",
		Messages: []config.Message{{Role: "user", Content: "hi"}},
	}
	msg, err := callOpenAI(cfg, p)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "azure response" {
		t.Errorf("want 'azure response', got %q", msg.Content)
	}
}
