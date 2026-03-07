package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jfgsilva/ten/internal/config"
)

func TestCallGemini(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "key=gemini-key") {
			t.Errorf("missing key query param: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.Path, "gemini-1.5-flash") {
			t.Errorf("model not in path: %s", r.URL.Path)
		}
		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]string{
							{"text": "hello from gemini"},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "gemini-key", URL: srv.URL}
	p := config.Prompt{
		API:   "gemini",
		Model: "gemini-1.5-flash",
		Messages: []config.Message{
			{Role: "system", Content: "be helpful"},
			{Role: "user", Content: "hi"},
		},
	}

	msg, err := callGemini(cfg, p)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "hello from gemini" {
		t.Errorf("want 'hello from gemini', got %q", msg.Content)
	}
}

func TestGeminiSystemMessagePrependedToUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req geminiRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Contents) != 1 {
			t.Errorf("want 1 content, got %d", len(req.Contents))
		}
		if req.Contents[0].Role != "user" {
			t.Errorf("want role user, got %s", req.Contents[0].Role)
		}
		if !strings.Contains(req.Contents[0].Parts[0].Text, "system text") {
			t.Errorf("system text not prepended: %s", req.Contents[0].Parts[0].Text)
		}
		resp := map[string]any{
			"candidates": []map[string]any{
				{"content": map[string]any{"parts": []map[string]string{{"text": "ok"}}}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := config.ApiConfig{APIKey: "k", URL: srv.URL}
	p := config.Prompt{
		API:   "gemini",
		Model: "gemini-1.5-flash",
		Messages: []config.Message{
			{Role: "system", Content: "system text"},
			{Role: "user", Content: "user text"},
		},
	}
	if _, err := callGemini(cfg, p); err != nil {
		t.Fatal(err)
	}
}
