package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jfgsilva/ten/internal/config"
)

func TestCallAnthropic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "anthropic-key" {
			t.Errorf("missing x-api-key: %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}
		resp := map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": "hello from anthropic"},
			},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "anthropic-key", URL: srv.URL, Version: "2023-06-01"}
	p := config.Prompt{
		API:   "anthropic",
		Model: "claude-3-opus-20240229",
		Messages: []config.Message{
			{Role: "system", Content: "you are a cat"},
			{Role: "user", Content: "hi"},
		},
	}

	msg, err := callAnthropic(cfg, p)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "hello from anthropic" {
		t.Errorf("want 'hello from anthropic', got %q", msg.Content)
	}
}

func TestMergeMessages(t *testing.T) {
	messages := []config.Message{
		{Role: "system", Content: "sys msg"},
		{Role: "user", Content: "hello"},
		{Role: "user", Content: "world"},
		{Role: "assistant", Content: "response"},
	}
	merged := mergeMessages(messages)

	// system → user, merged with next user
	if len(merged) != 2 {
		t.Fatalf("want 2 merged messages, got %d", len(merged))
	}
	if merged[0].Role != "user" {
		t.Errorf("want user, got %s", merged[0].Role)
	}
	if merged[0].Content != "sys msg\n\nhello\n\nworld" {
		t.Errorf("unexpected merged content: %q", merged[0].Content)
	}
	if merged[1].Role != "assistant" {
		t.Errorf("want assistant, got %s", merged[1].Role)
	}
}
