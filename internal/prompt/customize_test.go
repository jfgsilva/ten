package prompt

import (
	"os"
	"testing"

	"github.com/jfgsilva/ten/internal/config"
)

func defaultPrompt() config.Prompt { return config.DefaultPrompt() }
func emptyPrompt() config.Prompt   { return config.EmptyPrompt() }

func TestCustomizeEmptyNoOverrides(t *testing.T) {
	p := defaultPrompt()
	customized := Customize(p, Params{}, "")

	last := customized.Messages[len(customized.Messages)-1]
	if last.Role != "user" {
		t.Errorf("last message role: want user, got %s", last.Role)
	}
	if !containsPlaceholder(last.Content) {
		t.Error("last message should contain placeholder")
	}
}

func TestCustomizeAPIOverride(t *testing.T) {
	p := emptyPrompt()
	customized := Customize(p, Params{API: "openai"}, "")
	if customized.API != "openai" {
		t.Errorf("want openai, got %s", customized.API)
	}
}

func TestCustomizeModelOverride(t *testing.T) {
	p := emptyPrompt()
	customized := Customize(p, Params{Model: "gpt-4o"}, "")
	if customized.Model != "gpt-4o" {
		t.Errorf("want gpt-4o, got %s", customized.Model)
	}
}

func TestCustomizeCommandInsertion(t *testing.T) {
	p := emptyPrompt()
	customized := Customize(p, Params{}, "test command")
	found := false
	for _, m := range customized.Messages {
		if indexOf(m.Content, "test command") >= 0 {
			found = true
		}
	}
	if !found {
		t.Error("custom text not found in messages")
	}
}

func TestCustomizeTemperatureOverride(t *testing.T) {
	p := emptyPrompt()
	temp := 42.0
	customized := Customize(p, Params{Temperature: &temp}, "")
	if customized.Temperature == nil || *customized.Temperature != 42.0 {
		t.Errorf("want temperature 42.0, got %v", customized.Temperature)
	}
}

func TestCustomizeTemperatureZero(t *testing.T) {
	p := emptyPrompt()
	temp := 0.0
	customized := Customize(p, Params{Temperature: &temp}, "")
	if customized.Temperature == nil || *customized.Temperature != 1e-13 {
		t.Errorf("want temperature 1e-13 for zero input, got %v", customized.Temperature)
	}
}

func TestCustomizeContextFile(t *testing.T) {
	p := emptyPrompt()
	f, err := os.CreateTemp("", "ctx*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("hello there")
	f.Close()

	customized := Customize(p, Params{Context: []string{f.Name()}}, "")
	if len(customized.Messages) < 2 {
		t.Fatal("expected at least 2 messages after context injection")
	}
	ctx := customized.Messages[0]
	if ctx.Role != "system" {
		t.Errorf("context message role: want system, got %s", ctx.Role)
	}
	if indexOf(ctx.Content, "hello there") < 0 {
		t.Errorf("context file content not found in message: %s", ctx.Content)
	}
}

func TestCustomizePlaceholderOnlyOnce(t *testing.T) {
	p := config.Prompt{
		API: "ollama",
		Messages: []config.Message{
			{Role: "user", Content: "do this " + config.PlaceholderToken},
		},
	}
	customized := Customize(p, Params{}, "override text")
	count := 0
	for _, m := range customized.Messages {
		s := m.Content
		for {
			idx := indexOf(s, config.PlaceholderToken)
			if idx < 0 {
				break
			}
			count++
			s = s[idx+len(config.PlaceholderToken):]
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 placeholder, found %d", count)
	}
}
