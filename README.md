# ten

**ten** — like `cat`, but ten times smarter. A CLI tool that pipes text through an LLM.

[![CI](https://github.com/jfgsilva/ten/actions/workflows/ci.yml/badge.svg)](https://github.com/jfgsilva/ten/actions/workflows/ci.yml)
[![Publish](https://github.com/jfgsilva/ten/actions/workflows/publish.yml/badge.svg)](https://github.com/jfgsilva/ten/actions/workflows/publish.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/dl/)

`ten` is a Go port of [smartcat](https://github.com/efugier/smartcat). It reads text from stdin
or arguments, sends it to an LLM API, and writes the response to stdout. Designed for Unix
pipelines, editor integration, and terminal power users.

What makes it special:

- Made for power users; tailor the config to reduce overhead on your most frequent tasks
- Minimalist, built according to the Unix philosophy with terminal and editor integration in mind
- Good I/O handling to insert user input in prompts and use the result in CLI-based workflows
- Built-in default prompt to make the model behave as a CLI tool (no markdown, no explanations)
- Full configurability on which API, LLM version, and temperature you use
- Write and save your own prompt templates for faster recurring tasks (simplify, optimize, tests, etc.)
- Conversation support
- Glob expressions to include context files

Supports: OpenAI, Anthropic, Gemini, Ollama, Mistral, Groq, Cerebras, Azure OpenAI.

---

## Table of Contents

- [Installation](#installation)
- [Recommended Models](#recommended-models)
- [A few examples to get started](#a-few-examples-to-get-started)
  - [Integrating with editors](#integrating-with-editors)
    - [Example workflows](#example-workflows)
- [Configuration](#configuration)
  - [Ollama setup](#ollama-setup)
- [CLI reference](#cli-reference)
- [Supported APIs](#supported-apis)

---

## Installation

On the first run (`ten`), it will generate default configuration files and provide guidance on
finalizing the setup (see the [Configuration](#configuration) section).

### go install

Requires Go 1.26+. The binary runs natively on your machine, so password managers
(`op`, `bw`, `pass`) and local tools like Ollama work without any extra setup.

```sh
go install github.com/jfgsilva/ten/cmd/ten@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`:

```sh
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

To update, run the same command again.

### Download a release binary (tar -xvzf)

Pre-built binaries for Linux, macOS, and Windows are available on the
[Releases page](https://github.com/jfgsilva/ten/releases).

```sh
# macOS arm64 (Apple Silicon)
curl -L https://github.com/jfgsilva/ten/releases/latest/download/ten_darwin_arm64.tar.gz | tar -xvzf -
sudo mv ten /usr/local/bin/

# Linux x86_64
curl -L https://github.com/jfgsilva/ten/releases/latest/download/ten_linux_x86_64.tar.gz | tar -xvzf -
sudo mv ten /usr/local/bin/
```

To update, download the latest release and replace the binary.

| Platform       | File                        |
|----------------|-----------------------------|
| Linux x86_64   | `ten_linux_x86_64.tar.gz`   |
| Linux arm64    | `ten_linux_arm64.tar.gz`    |
| macOS x86_64   | `ten_darwin_x86_64.tar.gz`  |
| macOS arm64    | `ten_darwin_arm64.tar.gz`   |
| Windows x86_64 | `ten_windows_x86_64.zip`    |

### Docker / Podman alias

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

The config directory is mounted read-only (`:ro`) — the container reads your API keys and prompts
but cannot modify them. The current working directory is mounted at `/work` so `ten` can read
context files via `-c` globs.

> **Note:** Password manager CLIs (`op`, `bw`, `pass`) do **not** work inside the container.
> Use a plain `api_key = "sk-..."` in your config when running via Docker.

To update the Docker alias, pull the latest image:

```sh
docker pull ghcr.io/jfgsilva/ten:latest
```

---

## Recommended Models

Currently the best results are achieved with APIs from Anthropic, Mistral, or OpenAI. It costs
about $2–3 a month for typical use with the best models.

---

## A few examples to get started

```sh
ten "say hi"                                    # just ask (uses default prompt template)

ten test                                         # use a named prompt template
ten test "and parametrize them"                  # extend it on the fly

ten "explain how to use this program" -c **/*.md main.go   # use files as context

git diff | ten "summarize the changes"           # pipe data in

cat en.md | ten "translate to french" >> fr.md  # write data out
ten -e "use a more informal tone" -t 2 >> fr.md # extend the conversation and raise temperature
```

**The key to making this work seamlessly is a good default prompt** that tells the model to behave
like a CLI tool — no markdown formatting, no explanations, just the result.

### Integrating with editors

The key for good editor integration is a good default prompt combined with the `-r` flag to decide
whether to replace or extend the current selection.

#### Vim

Start by selecting some text, then press `:`. You can then pipe the selection to `ten`.

```vim
:'<,'>!ten "replace the versions with wildcards"
```

```vim
:'<,'>!ten "fix this function"
```

This will **overwrite** the current selection with the text transformed by the model.

```vim
:'<,'>!ten -r test
```

This will **repeat** the input, effectively appending the model's output at the end of the
current selection.

Add the following remap to your `vimrc` for easy access:

```vimrc
nnoremap <leader>ten :'<,'>!ten
```

#### Helix and Kakoune

Same concept, different shortcut — press the pipe key to redirect the selection to `ten`.

```
pipe:ten test -r
```

With some remapping you can attach your most frequent action to a few keystrokes, e.g. `<leader>wt`.

#### Example Workflows

**For quick questions:**

```sh
ten "my quick question"
```

This is likely **your fastest path to an answer**: open your terminal, type `ten`, done. No tab
switching, no logins, no redirects.

**To help with coding:**

Select a struct in your editor:

```vim
:'<,'>!ten "implement the traits FromStr and ToString for this struct"
```

Select the generated impl block:

```vim
:'<,'>!ten -e "can you make it more concise?"
```

Put the cursor at the bottom of the file and provide example usage as input:

```vim
:'<,'>!ten -e "now write tests for it knowing it's used like this" -c internal/main.go
```

**To have a full conversation from a markdown file:**

```sh
vim problem_solving.md

# Write your question as a comment in the markdown file, then select it
# and send it to ten using the trick above. Use -r to repeat the input.

# To continue the conversation, write your next question and repeat the step with -e -r.

# This lets you keep track of your questions and build a reusable document.
```

---

## Configuration

- By default, lives at `$HOME/.config/ten/`
- Override with the `TEN_CONFIG_PATH` or `SMARTCAT_CONFIG_PATH` environment variable
- Use `#[<input>]` as the placeholder for input when writing prompts; if none is provided, it is
  automatically appended to the last user message
- The prompt named `default` is used when no template name is given
- You can adjust the temperature and set a default for each prompt depending on its use case

Three files are used:

- `.api_configs.toml` — API credentials and endpoints; you need at least one provider
- `prompts.toml` — prompt templates; you need at least the `default` prompt
- `conversation.toml` — stores the latest conversation for `-e`; auto-managed

On first run, ten creates these files with sensible defaults.

If you are coming from smartcat, copy your existing configs directly:

```sh
cp ~/.config/smartcat/.api_configs.toml ~/.config/ten/.api_configs.toml
cp ~/.config/smartcat/prompts.toml ~/.config/ten/prompts.toml
```

### `~/.config/ten/.api_configs.toml`

```toml
[ollama]  # local, no key required
url = "http://localhost:11434/api/chat"
default_model = "phi3"
timeout_seconds = 180

[openai]
api_key = "<your_api_key>"
url = "https://api.openai.com/v1/chat/completions"
default_model = "gpt-4o"
timeout_seconds = 30

[anthropic]
api_key = "<your_api_key>"
url = "https://api.anthropic.com/v1/messages"
default_model = "claude-3-5-sonnet-20241022"
version = "2023-06-01"

[gemini]
api_key = "<your_api_key>"
default_model = "gemini-1.5-pro"

[mistral]
# you can use a command to grab the key — requires a working `sh`
api_key_command = "pass mistral/api_key"
default_model = "mistral-medium"
url = "https://api.mistral.ai/v1/chat/completions"

[groq]
api_key_command = "echo $MY_GROQ_API_KEY"
default_model = "llama3-70b-8192"
url = "https://api.groq.com/openai/v1/chat/completions"

[cerebras]
api_key = "<your_api_key>"
default_model = "llama3.1-70b"
url = "https://api.cerebras.ai/v1/chat/completions"
```

`api_key_command` runs any shell command to retrieve the API key. Supports:
- **1Password**: `op read op://vault/item/field`
- **Bitwarden**: `bw get password item-name`
- **pass**: `pass path/to/secret`
- Any arbitrary shell command

Plain `api_key = "sk-..."` is also accepted but not recommended for keys stored on disk.

### `~/.config/ten/prompts.toml`

```toml
[default]  # a prompt is a section
api = "ollama"
model = "phi3"

[[default.messages]]
role = "system"
content = """\
You are an expert programmer and a shell master. You value code efficiency and clarity above all things. \
What you write will be piped in and out of cli programs so you do not explain anything unless explicitly asked to. \
Never write ``` around your answer, provide only the result of the task you are given. Preserve input formatting.\
"""

[empty]  # always nice to have an empty prompt available
api = "openai"
messages = []

[test]
api = "anthropic"
temperature = 0.0

[[test.messages]]
role = "system"
content = """\
You are an expert programmer and a shell master. You value code efficiency and clarity above all things. \
What you write will be piped in and out of cli programs so you do not explain anything unless explicitly asked to. \
Never write ``` around your answer, provide only the result of the task you are given. Preserve input formatting.\
"""

[[test.messages]]
role = "user"
# #[<input>] is replaced with the piped text or argument input
content = '''Write tests using pytest for the following code. Parametrize it if appropriate.

#[<input>]
'''
```

### Ollama setup

1. [Install Ollama](https://github.com/ollama/ollama#ollama)
2. Pull the model you plan to use: `ollama pull phi3`
3. Test it: `ollama run phi3 "say hi"`
4. Verify the server is running: `curl http://localhost:11434` should say "Ollama is running"
   (if not, run `ollama serve`)
5. `ten` will now reach your local Ollama — enjoy!

> Answers may be slow depending on your setup. Timeout is configurable and defaults to 30s.

---

## CLI reference

```
ten [OPTIONS] [INPUT_OR_TEMPLATE] [INPUT_IF_TEMPLATE]

Arguments:
  INPUT_OR_TEMPLATE   a named prompt template from config, or plain input text
                      (uses the `default` template if it matches no template name)
  INPUT_IF_TEMPLATE   if the first arg matched a template, this is used as input

Options:
  -e, --extend-conversation   extend the previous conversation
  -r, --repeat-input          repeat input before output (useful for editor workflows)
      --api string            override which API to use
  -m, --model string          override which model to use
  -t, --temperature float64   higher temperature = answers further from the average
  -l, --char-limit int        max chars to include (0 = no limit)
  -c, --context stringArray   glob patterns or files to use as context (must be last)
  -h, --help                  print help
      --version               print version and exit
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

## Development

```sh
go build ./cmd/ten
echo "say hi" | TEN_TEST=1 ./ten

# run tests with coverage
go test -coverprofile=coverage.out -covermode=atomic ./internal/...
go tool cover -func=coverage.out

# run all tests (including integration)
go test ./...

```

---

## License

Apache 2.0. See [LICENSE](LICENSE).

This project is a Go port of [smartcat](https://github.com/efugier/smartcat) by efugier,
also licensed under Apache 2.0. See [NOTICE](NOTICE).
