package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	tenapi "github.com/jfgsilva/ten/internal/api"
	"github.com/jfgsilva/ten/internal/config"
	tenprompt "github.com/jfgsilva/ten/internal/prompt"
)

const defaultPromptName = "default"

var version = "dev"

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
}

func main() {
	// Test mode: mirror stdin to stdout for integration testing
	if os.Getenv("TEN_TEST") == "1" {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Hello, World!\n```\n%s\n```\n", input)
		os.Exit(0)
	}

	var (
		extendConversation bool
		repeatInput        bool
		apiOverride        string
		modelOverride      string
		temperatureFlag    float64
		temperatureSet     bool
		charLimit          int
		contextGlobs       []string
	)

	root := &cobra.Command{
		Use:   "ten [INPUT_OR_TEMPLATE] [INPUT_IF_TEMPLATE]",
		Short: "Putting a brain behind cat. CLI interface to LLMs in the Unix ecosystem.",
		Long: `ten — like cat, but ten times smarter.

Examples:
  ten "say hi"
  echo "some code" | ten "explain this"
  git diff | ten "summarize the changes"
  ten test "parametrize the template"
  ten -e "follow up question"
  ten "explain this" -c **/*.go main.go`,
		Version:      version,
		Args:         cobra.MaximumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			isInteractive := term.IsTerminal(int(os.Stdin.Fd())) // #nosec G115 -- standard idiom, safe on all supported platforms

			if err := config.EnsureConfigFiles(isInteractive); err != nil {
				return fmt.Errorf("config setup: %w", err)
			}

			var customText string
			var p config.Prompt

			if extendConversation {
				var err error
				p, err = config.LoadConversation()
				if err != nil {
					return fmt.Errorf("loading conversation: %w\nRun ten first to create a conversation.", err)
				}
				if len(args) > 1 {
					return fmt.Errorf("cannot provide a template ref when extending a conversation.\nUse: ten -e \"<your prompt>\"")
				}
				if len(args) == 1 {
					customText = args[0]
				}
			} else {
				prompts, err := config.LoadPrompts()
				if err != nil {
					return fmt.Errorf("loading prompts: %w", err)
				}

				inputOrRef := defaultPromptName
				if len(args) > 0 {
					inputOrRef = args[0]
				}

				if tmpl, ok := prompts[inputOrRef]; ok {
					p = tmpl
					if len(args) > 1 {
						customText = args[1]
					}
				} else {
					// First arg is not a prompt name — use default prompt, treat arg as input
					if len(args) > 1 {
						return fmt.Errorf("invalid parameters: provide a valid template ref then input, or just input.\n" +
							"Use: ten <template> \"<input>\" or ten \"<input>\"")
					}
					defaultPrompt, ok := prompts[defaultPromptName]
					if !ok {
						return fmt.Errorf("default prompt not found in prompts.toml")
					}
					p = defaultPrompt
					if inputOrRef != defaultPromptName {
						customText = inputOrRef
					}
				}
			}

			// Read stdin if piped
			var stdinInput string
			if !isInteractive {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				stdinInput = string(data)
			}

			// Input priority: piped stdin > customText arg
			input := stdinInput
			if input == "" {
				input = customText
				customText = ""
			}

			// Build prompt customization params
			params := tenprompt.Params{
				API:      apiOverride,
				Model:    modelOverride,
				Context:  contextGlobs,
			}
			if temperatureSet {
				params.Temperature = &temperatureFlag
			}
			if charLimit != 0 {
				params.CharLimit = &charLimit
			}

			p = tenprompt.Customize(p, params, customText)

			// Replace placeholder with actual input
			for i := range p.Messages {
				p.Messages[i].Content = strings.ReplaceAll(
					p.Messages[i].Content,
					config.PlaceholderToken,
					input,
				)
			}

			// Validate char limit
			totalChars := 0
			for _, m := range p.Messages {
				totalChars += len(m.Content)
			}
			if p.CharLimit != nil && *p.CharLimit > 0 && totalChars > *p.CharLimit {
				if isInteractive {
					fmt.Fprintf(os.Stderr, "Warning: prompt is %d chars (limit %d). Continue? [y/N] ",
						totalChars, *p.CharLimit)
					var answer string
					if _, err := fmt.Scanln(&answer); err != nil {
						return fmt.Errorf("aborted")
					}
					if strings.ToLower(answer) != "y" {
						return fmt.Errorf("aborted")
					}
				} else {
					return fmt.Errorf("prompt exceeds char limit: %d > %d", totalChars, *p.CharLimit)
				}
			}

			// Look up API config
			apiCfg := config.GetAPIConfig(p.API)

			// Make API call
			response, err := tenapi.Call(apiCfg, p)
			if err != nil {
				return fmt.Errorf("api call: %w", err)
			}

			// Output
			if repeatInput {
				fmt.Print(input)
				if len(input) > 0 && input[len(input)-1] != '\n' {
					fmt.Println()
				}
			}
			fmt.Print(response.Content)
			if len(response.Content) > 0 && response.Content[len(response.Content)-1] != '\n' {
				fmt.Println()
			}

			// Save conversation
			p.Messages = append(p.Messages, response)
			if err := config.SaveConversation(p); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not save conversation: %v\n", err)
			}

			return nil
		},
	}

	root.Flags().BoolVarP(&extendConversation, "extend-conversation", "e", false, "extend previous conversation")
	root.Flags().BoolVarP(&repeatInput, "repeat-input", "r", false, "repeat input before output")
	root.Flags().StringVar(&apiOverride, "api", "", "override api")
	root.Flags().StringVarP(&modelOverride, "model", "m", "", "override model")
	root.Flags().StringArrayVarP(&contextGlobs, "context", "c", nil, "glob patterns for context files (must be last)")
	root.Flags().IntVarP(&charLimit, "char-limit", "l", 0, "max chars (0 = no limit)")

	// Custom temperature handling to detect whether it was set
	root.Flags().Float64VarP(&temperatureFlag, "temperature", "t", 0, "override temperature")
	root.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("temperature") {
			temperatureSet = true
		}
		return nil
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
