package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jfgsilva/ten/internal/config"
)

func TestCallOllama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"message": map[string]string{"role": "assistant", "content": "hello from ollama"},
		}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	cfg := config.ApiConfig{URL: srv.URL, TimeoutSeconds: 5}
	p := config.Prompt{
		API:   "ollama",
		Model: "phi3",
		Messages: []config.Message{
			{Role: "user", Content: "hi"},
		},
	}

	msg, err := callOllama(cfg, p)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "hello from ollama" {
		t.Errorf("want 'hello from ollama', got %q", msg.Content)
	}
}
