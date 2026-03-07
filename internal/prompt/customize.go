package prompt

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jfgsilva/ten/internal/config"
)

// Params holds CLI overrides for prompt customization.
type Params struct {
	API         string
	Model       string
	Temperature *float64
	CharLimit   *int
	Context     []string // glob patterns
}

// Customize applies overrides, injects context files, and ensures the
// placeholder token exists in the last user message.
func Customize(p config.Prompt, params Params, customText string) config.Prompt {
	// Override parameters
	if params.API != "" {
		p.API = params.API
	}
	if params.Model != "" {
		p.Model = params.Model
	}
	if params.CharLimit != nil {
		p.CharLimit = params.CharLimit
	}

	// Load context files
	var contextContent string
	for _, pattern := range params.Context {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: invalid glob %q: %v\n", pattern, err)
			continue
		}
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not read %q: %v\n", path, err)
				continue
			}
			contextContent += fmt.Sprintf("%s:\n```\n%s\n```\n", path, string(data))
		}
	}
	if contextContent != "" {
		p.Messages = append(p.Messages, config.Message{
			Role:    "system",
			Content: "files content for context:\n\n" + contextContent,
		})
	}

	// If custom text provided, strip existing placeholders and add new user message
	if customText != "" {
		promptMessage := customText
		if !containsPlaceholder(promptMessage) {
			promptMessage += config.PlaceholderToken
		}
		for i := range p.Messages {
			p.Messages[i].Content = removePlaceholder(p.Messages[i].Content)
		}
		p.Messages = append(p.Messages, config.Message{Role: "user", Content: promptMessage})
	}

	// Ensure last message is a user message containing the placeholder
	var lastMessage config.Message
	if len(p.Messages) == 0 || p.Messages[len(p.Messages)-1].Role != "user" {
		lastMessage = config.Message{Role: "user", Content: config.PlaceholderToken}
	} else {
		lastMessage = p.Messages[len(p.Messages)-1]
		p.Messages = p.Messages[:len(p.Messages)-1]
	}
	if !containsPlaceholder(lastMessage.Content) {
		lastMessage.Content += config.PlaceholderToken
	}

	// Temperature override (0 → 1e-13 to avoid API quirks)
	if params.Temperature != nil {
		t := *params.Temperature
		if t == 0 {
			t = 1e-13
		}
		p.Temperature = &t
	}

	p.Messages = append(p.Messages, lastMessage)
	return p
}

func containsPlaceholder(s string) bool {
	return len(s) >= len(config.PlaceholderToken) &&
		indexOf(s, config.PlaceholderToken) >= 0
}

func removePlaceholder(s string) string {
	result := ""
	token := config.PlaceholderToken
	for {
		idx := indexOf(s, token)
		if idx < 0 {
			result += s
			break
		}
		result += s[:idx]
		s = s[idx+len(token):]
	}
	return result
}

func indexOf(s, sub string) int {
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
