package config

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDefaultPrompt(t *testing.T) {
	p := DefaultPrompt()
	if p.API != "ollama" {
		t.Errorf("want ollama, got %s", p.API)
	}
	if len(p.Messages) != 1 {
		t.Errorf("want 1 message, got %d", len(p.Messages))
	}
	if p.Messages[0].Role != "system" {
		t.Errorf("want system role, got %s", p.Messages[0].Role)
	}
	if p.CharLimit == nil || *p.CharLimit != 50000 {
		t.Errorf("want CharLimit=50000")
	}
}

func TestEmptyPrompt(t *testing.T) {
	p := EmptyPrompt()
	if p.API != "ollama" {
		t.Errorf("want ollama, got %s", p.API)
	}
	if len(p.Messages) != 0 {
		t.Errorf("want 0 messages, got %d", len(p.Messages))
	}
	if p.CharLimit == nil || *p.CharLimit != 50000 {
		t.Errorf("want CharLimit=50000")
	}
}

func TestGenerateDefaultPrompts(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	if err := GenerateDefaultPrompts(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(PromptsPath()); err != nil {
		t.Fatal("prompts file not created:", err)
	}
	prompts, err := LoadPrompts()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := prompts["default"]; !ok {
		t.Error("missing 'default' key")
	}
	if _, ok := prompts["empty"]; !ok {
		t.Error("missing 'empty' key")
	}
}

func TestLoadPrompts(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	if err := GenerateDefaultPrompts(); err != nil {
		t.Fatal(err)
	}
	prompts, err := LoadPrompts()
	if err != nil {
		t.Fatal(err)
	}
	if prompts["default"].API != "ollama" {
		t.Errorf("want ollama, got %s", prompts["default"].API)
	}
}

func TestLoadPromptsNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	_, err := LoadPrompts()
	if err == nil {
		t.Error("expected error when prompts file missing")
	}
}

func TestSaveAndLoadConversation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	p := Prompt{
		API:   "openai",
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
			{Role: "assistant", Content: "hi there"},
		},
	}
	if err := SaveConversation(p); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConversation()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.API != "openai" {
		t.Errorf("want openai, got %s", loaded.API)
	}
	if loaded.Model != "gpt-4" {
		t.Errorf("want gpt-4, got %s", loaded.Model)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("want 2 messages, got %d", len(loaded.Messages))
	}
	if loaded.Messages[0].Content != "hello" {
		t.Errorf("want 'hello', got %q", loaded.Messages[0].Content)
	}
	if loaded.Messages[1].Content != "hi there" {
		t.Errorf("want 'hi there', got %q", loaded.Messages[1].Content)
	}
}

func TestLoadConversationNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	_, err := LoadConversation()
	if err == nil {
		t.Error("expected error when conversation file missing")
	}
}

func TestEnsureConfigFilesCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	if err := EnsureConfigFiles(false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(PromptsPath()); err != nil {
		t.Error("prompts file not created:", err)
	}
	if _, err := os.Stat(APIConfigPath()); err != nil {
		t.Error("api config file not created:", err)
	}
}

func TestEnsureConfigFilesIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	if err := EnsureConfigFiles(false); err != nil {
		t.Fatal("first call failed:", err)
	}
	if err := EnsureConfigFiles(false); err != nil {
		t.Fatal("second call failed:", err)
	}
}

func TestEnsureConfigFilesInteractiveMode(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(tenConfigEnvVar, dir)

	// Redirect os.Stderr to capture output
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	callErr := EnsureConfigFiles(true)

	w.Close() //nolint:errcheck
	os.Stderr = oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	r.Close()        //nolint:errcheck

	if callErr != nil {
		t.Fatal(callErr)
	}
	if !strings.Contains(buf.String(), "Prompt config not found") {
		t.Errorf("expected stderr to contain 'Prompt config not found', got: %q", buf.String())
	}
}
