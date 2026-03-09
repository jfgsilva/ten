package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const PlaceholderToken = "#[<input>]" // #nosec G101 -- not a credential, this is a template placeholder

// Message is a single chat message.
type Message struct {
	Role    string `toml:"role" json:"role"`
	Content string `toml:"content" json:"content"`
}

// Prompt is a named prompt template loaded from prompts.toml.
type Prompt struct {
	API         string   `toml:"api"`
	Model       string   `toml:"model,omitempty"`
	Messages    []Message `toml:"messages"`
	Temperature *float64 `toml:"temperature,omitempty"`
	CharLimit   *int     `toml:"char_limit,omitempty"`
}

// DefaultPrompt returns the built-in default prompt (ollama/phi3).
func DefaultPrompt() Prompt {
	limit := 50000
	return Prompt{
		API: "ollama",
		Messages: []Message{
			{
				Role: "system",
				Content: "You are an extremely skilled programmer with a keen eye for detail and an emphasis on readable code. " +
					"You have been tasked with acting as a smart version of the cat unix program. You take text and a prompt in and write text out. " +
					"For that reason, it is of crucial importance to just write the desired output. Do not under any circumstance write any comment or thought " +
					"as your output will be piped into other programs. Do not write the markdown delimiters for code as well. " +
					"Sometimes you will be asked to implement or extend some input code. Same thing goes here, write only what was asked because what you write will " +
					"be directly added to the user's editor. " +
					"Never ever write ``` around the code. " +
					"Make sure to keep the indentation and formatting. ",
			},
		},
		CharLimit: &limit,
	}
}

// EmptyPrompt returns a prompt with no messages and the default API settings.
func EmptyPrompt() Prompt {
	limit := 50000
	return Prompt{
		API:       "ollama",
		CharLimit: &limit,
	}
}

// LoadPrompts reads prompts.toml and returns the map.
func LoadPrompts() (map[string]Prompt, error) {
	var prompts map[string]Prompt
	if _, err := toml.DecodeFile(PromptsPath(), &prompts); err != nil {
		return nil, fmt.Errorf("reading %s: %w", PromptsPath(), err)
	}
	return prompts, nil
}

// LoadConversation reads conversation.toml as a Prompt.
func LoadConversation() (Prompt, error) {
	var p Prompt
	if _, err := toml.DecodeFile(ConversationPath(), &p); err != nil {
		return Prompt{}, fmt.Errorf("reading conversation: %w", err)
	}
	return p, nil
}

// SaveConversation writes p to conversation.toml.
func SaveConversation(p Prompt) error {
	if err := os.MkdirAll(ConfigDir(), 0o750); err != nil {
		return err
	}
	f, err := os.Create(ConversationPath())
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	return toml.NewEncoder(f).Encode(p)
}

// GenerateDefaultPrompts writes a starter prompts.toml.
func GenerateDefaultPrompts() error {
	if err := os.MkdirAll(ConfigDir(), 0o750); err != nil {
		return err
	}
	f, err := os.Create(PromptsPath())
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	header := "# Prompt config file\n# more details at https://github.com/jfgsilva/ten\n\n"
	if _, err := f.WriteString(header); err != nil {
		return err
	}

	prompts := map[string]Prompt{
		"default": DefaultPrompt(),
		"empty":   EmptyPrompt(),
	}
	return toml.NewEncoder(f).Encode(prompts)
}

// EnsureConfigFiles creates config files if they don't exist yet.
// In non-interactive mode it skips printing setup guidance.
func EnsureConfigFiles(interactive bool) error {
	if _, err := os.Stat(PromptsPath()); os.IsNotExist(err) {
		if interactive {
			fmt.Fprintf(os.Stderr, "Prompt config not found at %s, generating one.\n", PromptsPath())
		}
		if err := GenerateDefaultPrompts(); err != nil {
			return fmt.Errorf("generating prompts file: %w", err)
		}
	}
	if _, err := os.Stat(APIConfigPath()); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "API config not found at %s, generating one.\n", APIConfigPath())
		if err := GenerateDefaultAPIConfigs(); err != nil {
			return fmt.Errorf("generating api config: %w", err)
		}
	}
	return nil
}
