# CLAUDE.md — smartcat Go Port

## Project Goal

Rewrite [smartcat](https://github.com/efugier/smartcat) (originally in Rust) in Go.
Smartcat is a CLI tool that acts as a "smart cat" — it takes text input and a prompt, sends them to an LLM API, and returns the result. It is designed for Unix pipelines, editor integration (vim, helix, kakoune), and terminal power users.

The original is not being actively maintained and the owner uses it daily but cannot maintain Rust. Go is the target language.

## Key Design Goals

1. **Container-first**: The binary must work well when run inside Docker. Users should be able to alias it like skopeo:
   ```sh
   alias sc='docker run --rm -i \
     -v $HOME/.config/smartcat:/root/.config/smartcat \
     -v $(pwd):/work \
     -w /work \
     ghcr.io/yourname/smartcat:latest'
   ```
   This means:
   - Config must be mountable via `-v`
   - stdin/stdout must work correctly (no TTY assumptions in piped mode)
   - No interactive prompts when stdin is not a terminal

2. **Secret management via password managers**: No secrets stored in plain config files. Support:
   - `api_key_command` field in config (already in Rust version) — executes a shell command to get the key
   - 1Password CLI: `op read op://vault/item/field`
   - Bitwarden CLI: `bw get password item-name`
   - Only use these if the corresponding binary (`op`, `bw`) is present in PATH

3. **Gemini support**: Add Google Gemini API support (not in the original Rust version).

4. **Full feature parity with original smartcat**.

## Original Smartcat Architecture (Rust)

### CLI Args (clap)
```
sc [OPTIONS] [INPUT_OR_TEMPLATE_REF] [INPUT_IF_TEMPLATE_REF]

  -e, --extend-conversation   extend previous conversation
  -r, --repeat-input          repeat input before output
      --api <API>             override api
  -m, --model <MODEL>         override model
  -t, --temperature <TEMP>    override temperature
  -l, --char-limit <LIMIT>    max chars (0 = no limit)
  -c, --context <GLOBS>...    glob patterns for context files (must be last)
```

### Config Files
Location: `$HOME/.config/smartcat/` or `$SMARTCAT_CONFIG_PATH`

- `.api_configs.toml` — API credentials and endpoints
- `prompts.toml` — named prompt templates
- `conversation.toml` — last conversation state (auto-managed)

### `.api_configs.toml` structure
```toml
[openai]
api_key = "..."                        # plain key (avoid — use command instead)
api_key_command = "op read op://..."   # shell command to fetch key
url = "https://api.openai.com/v1/chat/completions"
default_model = "gpt-4"
timeout_seconds = 30

[anthropic]
api_key_command = "op read op://vault/anthropic/credential"
url = "https://api.anthropic.com/v1/messages"
default_model = "claude-3-opus-20240229"
version = "2023-06-01"

[ollama]
url = "http://localhost:11434/api/chat"
default_model = "phi3"
timeout_seconds = 180
```

### `prompts.toml` structure
```toml
[default]
api = "ollama"
model = "phi3"

[[default.messages]]
role = "system"
content = "You are a smart cat unix tool..."

[test]
api = "anthropic"
temperature = 0.0

[[test.messages]]
role = "user"
content = "Write tests for:\n\n#[<input>]"
```

Placeholder: `#[<input>]` is replaced with the actual input text. If absent, it is appended to the last user message.

### Supported APIs
| Name         | Auth header              | Request format   | Response format |
|--------------|--------------------------|------------------|-----------------|
| openai       | `Authorization: Bearer`  | OpenAI chat      | OpenAI          |
| azureopenai  | `api-key`                | OpenAI chat      | OpenAI          |
| mistral      | `Authorization: Bearer`  | OpenAI chat      | OpenAI          |
| groq         | `Authorization: Bearer`  | OpenAI chat      | OpenAI          |
| cerebras     | `Authorization: Bearer`  | OpenAI chat      | OpenAI          |
| ollama       | none                     | Ollama chat      | Ollama          |
| anthropic    | `x-api-key` + version    | Anthropic        | Anthropic       |
| **gemini**   | (to add)                 | Gemini           | Gemini          |

OpenAI-compatible format is reused for: openai, azureopenai, mistral, groq, cerebras.

Anthropic requires:
- Merging consecutive same-role messages
- Converting `system` role to `user` role
- Adding `max_tokens` field
- `anthropic-version` header

Ollama response schema differs from OpenAI.

### Core Flow
1. Parse args
2. Load/generate config files
3. Determine prompt: use named template or `default`, optionally extend last conversation
4. Customize prompt: inject context files, override api/model/temp, insert input placeholder
5. Replace `#[<input>]` placeholder with actual stdin or arg input
6. Validate char limit (prompt user if interactive, fail if not)
7. POST to API, parse response
8. Write response to stdout (optionally prepend input with `-r`)
9. Save full conversation to `conversation.toml`

### Input Logic
- If stdin is a pipe: read stdin as input
- If no stdin and no args: use empty string
- First positional arg: either a named prompt template OR the input text
- Second positional arg (only if first matched a template): the actual input

### Conversation Extension (`-e`)
Loads `conversation.toml` as the prompt and appends the new user message to it.

## New Features to Add (vs Rust original)

### Gemini API
- Endpoint: `https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={api_key}`
- Request/response format differs from OpenAI — needs its own schema
- Auth: query param `key=` not a header

### Secret Management
Config field `api_key_command` already exists in original. Extend to document and test:
- `op read op://vault/item/field` (1Password, requires `op` CLI)
- `bw get password item-name` (Bitwarden, requires `bw` CLI)
- `pass path/to/secret` (pass, requires `pass` CLI)
- Any arbitrary shell command works since it's just `sh -c <command>`

### Container compatibility
- No interactive TTY prompts when stdin is not a terminal (original does this — preserve it)
- Config path must be overridable via `SMARTCAT_CONFIG_PATH` env var (already exists — keep it)
- Dockerfile should be minimal (scratch or alpine), single binary
- The alias mounts `$HOME/.config/smartcat` into `/root/.config/smartcat` in the container

## Go Implementation Notes

- Use `cobra` or plain `flag`/`os.Args` for CLI parsing (cobra preferred for subcommand extensibility)
- Use `encoding/json` for API payloads
- Use `github.com/BurntSushi/toml` for config parsing
- Use `path/filepath.Glob` for context file globs
- Config path resolution: `SMARTCAT_CONFIG_PATH` env var → `$HOME/.config/smartcat/`
- HTTP client with configurable timeout per API config
- Detect terminal with `golang.org/x/term` or `os.Stdin.Stat()`

## File Structure (proposed)

```
/
  cmd/sc/main.go          # entry point, CLI parsing
  internal/
    config/
      api.go              # ApiConfig, load .api_configs.toml
      prompt.go           # Prompt, Message, load prompts.toml
      paths.go            # config path resolution
    api/
      client.go           # HTTP POST, get API key (including command exec)
      openai.go           # OpenAI-compatible request/response
      anthropic.go        # Anthropic request/response
      ollama.go           # Ollama request/response
      gemini.go           # Gemini request/response (new)
    prompt/
      customize.go        # placeholder injection, context loading, overrides
  Dockerfile
  README.md
```

## Source Reference

Original Rust source is in `./smartcat/` (no .git, kept for reference only).
Key files:
- `smartcat/src/main.rs` — CLI and entry point
- `smartcat/src/config/api.rs` — API config structs and key resolution
- `smartcat/src/config/prompt.rs` — Prompt and Message structs
- `smartcat/src/config/mod.rs` — config file generation and path resolution
- `smartcat/src/text/api_call.rs` — HTTP request dispatch per API
- `smartcat/src/text/request_schemas.rs` — OpenAI and Anthropic request bodies
- `smartcat/src/text/response_schemas.rs` — response parsing
- `smartcat/src/prompt_customization.rs` — placeholder injection, context, overrides
