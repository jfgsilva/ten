# ten

**ten** — like `cat`, but ten times smarter. A CLI tool that pipes text through an LLM.

[![CI](https://github.com/jfgsilva/ten/actions/workflows/ci.yml/badge.svg)](https://github.com/jfgsilva/ten/actions/workflows/ci.yml)
[![Publish](https://github.com/jfgsilva/ten/actions/workflows/publish.yml/badge.svg)](https://github.com/jfgsilva/ten/actions/workflows/publish.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/dl/)

`ten` is a Go port of [smartcat](https://github.com/efugier/smartcat). It reads text from stdin
or arguments, sends it to an LLM API, and writes the response to stdout. Designed for Unix
pipelines, Neovim/editor integration, and terminal power users.

Supports: OpenAI, Anthropic, Gemini, Ollama, Mistral, Groq, Cerebras, Azure OpenAI.

---

## Installation

### go install (requires Go 1.22+)

```sh
go install github.com/jfgsilva/ten/cmd/ten@latest
```

### Docker / Podman alias (no local Go required)

Add this alias to your `~/.bashrc` or `~/.zshrc`:

```sh
# Docker
alias ten='docker run --rm -i \
  -v $HOME/.config/ten:/root/.config/ten:ro \
  -v $(pwd):/work \
  -w /work \
  ghcr.io/jfgsilva/ten:latest'
```

```sh
# Podman
alias ten='podman run --rm -i \
  -v $HOME/.config/ten:/root/.config/ten:ro \
  -v $(pwd):/work \
  -w /work \
  ghcr.io/jfgsilva/ten:latest'
```

Then reload your shell:

```sh
source ~/.zshrc   # or source ~/.bashrc
```

The config directory is mounted read-only (`:ro`) — the container can read your API keys and
prompts but cannot modify them. The current working directory is mounted at `/work` so ten can
read context files via `-c` globs.

### Build from source

```sh
git clone https://github.com/jfgsilva/ten.git
cd ten
CGO_ENABLED=0 go build -ldflags="-s -w" -o ten ./cmd/ten
```

---

## Configuration

ten looks for config files in `$HOME/.config/ten/` by default.
Override with the `TEN_CONFIG_PATH` or `SMARTCAT_CONFIG_PATH` environment variable.

On first run, ten creates default config files if they do not exist.

If you are coming from smartcat, you can copy your existing configs directly:

```sh
cp ~/.config/smartcat/.api_configs.toml ~/.config/ten/.api_configs.toml
cp ~/.config/smartcat/prompts.toml ~/.config/ten/prompts.toml
```

### `~/.config/ten/.api_configs.toml`

```toml
[openai]
api_key_command = "op read op://vault/openai/credential"
url = "https://api.openai.com/v1/chat/completions"
default_model = "gpt-4o"
timeout_seconds = 30

[anthropic]
api_key_command = "op read op://vault/anthropic/credential"
url = "https://api.anthropic.com/v1/messages"
default_model = "claude-3-5-sonnet-20241022"
version = "2023-06-01"

[gemini]
api_key_command = "bw get password gemini-api-key"
default_model = "gemini-1.5-pro"

[ollama]
url = "http://localhost:11434/api/chat"
default_model = "phi3"
timeout_seconds = 180
```

`api_key_command` runs any shell command to retrieve the API key. Supports:
- **1Password**: `op read op://vault/item/field`
- **Bitwarden**: `bw get password item-name`
- **pass**: `pass path/to/secret`
- Any arbitrary shell command

Plain `api_key = "sk-..."` is also accepted but not recommended.

### `~/.config/ten/prompts.toml`

```toml
[default]
api = "ollama"
model = "phi3"

[[default.messages]]
role = "system"
content = "You are a helpful assistant."

[explain]
api = "anthropic"

[[explain.messages]]
role = "user"
content = "Explain the following:\n\n#[<input>]"
```

`#[<input>]` is the placeholder for piped text or argument input. If absent, the input is
appended to the last user message.

---

## Usage

```sh
# Ask a one-off question
ten "what is the capital of Portugal"

# Pipe text through the default prompt
echo "some code" | ten "explain this"

# Use a named prompt template
git diff | ten explain

# Override api and model inline
git diff | ten --api openai --model gpt-4o "summarize these changes"

# Include context files
ten "refactor this" -c internal/**/*.go

# Extend the previous conversation
ten -e "now make it more concise"

# Repeat input before output (useful for editor workflows)
echo "fix the bug below" | ten -r
```

### Neovim / Editor integration

**Neovim** — pipe a visual selection through ten:
```vim
:'<,'>!ten "rewrite this more clearly"
```

**Helix** — pipe selection:
```
pipe-to ten "fix typos"
```

**Kakoune:**
```
<a-|>ten "explain this"<ret>
```

---

## CLI reference

```
ten [OPTIONS] [INPUT_OR_TEMPLATE] [INPUT_IF_TEMPLATE]

Flags:
  -e, --extend-conversation   extend previous conversation
  -r, --repeat-input          repeat input before output
      --api string            override api
  -m, --model string          override model
  -t, --temperature float64   override temperature
  -l, --char-limit int        max chars (0 = no limit)
  -c, --context stringArray   glob patterns for context files
```

---

## Supported APIs

| Name         | Notes                               |
|--------------|-------------------------------------|
| openai       | OpenAI-compatible                   |
| azureopenai  | OpenAI-compatible, `api-key` header |
| mistral      | OpenAI-compatible                   |
| groq         | OpenAI-compatible                   |
| cerebras     | OpenAI-compatible                   |
| ollama       | Local, no key required              |
| anthropic    | Native Anthropic format             |
| gemini       | Google Gemini, query-param auth     |

---

## Releases

Releases are triggered by pushing a semver git tag:

```sh
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This triggers the Publish workflow which builds a multi-platform image (`linux/amd64`,
`linux/arm64`) and pushes it to GHCR. The following image tags are produced:

- `ghcr.io/jfgsilva/ten:1.0.0` — exact version
- `ghcr.io/jfgsilva/ten:1.0` — floating minor pointer
- `ghcr.io/jfgsilva/ten:latest` — always the latest release

---

## Development

```sh
go test ./...
go build ./cmd/ten
echo "say hi" | TEN_TEST=1 ./ten
```

---

## License

Apache 2.0. See [LICENSE](LICENSE).

This project is a Go port of [smartcat](https://github.com/efugier/smartcat) by efugier,
also licensed under Apache 2.0. See [NOTICE](NOTICE).
