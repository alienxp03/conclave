# dbate

AI-powered debate tool that orchestrates discussions between AI agents with different personas.

## Features

- **Multi-Provider Support**: Works with Claude, OpenAI Codex, Gemini, and Qwen CLI tools
- **Agent Personas**: 6 built-in personas (Optimist, Skeptic, Pragmatist, Visionary, Analyst, Devil's Advocate)
- **Debate Styles**: 4 debate formats (Adversarial, Collaborative, Analytical, Socratic)
- **Early Consensus Detection**: Automatically ends debate when agents reach agreement
- **Export**: Save debates as Markdown, PDF, or JSON
- **CLI & Web Interface**: Use from terminal or browser
- **Session History**: SQLite-based storage for all debate sessions
- **Turn-by-Turn Mode**: Step through debates manually with voting
- **Read-Only Lock**: Protect important debates from modification

## Installation

### Prerequisites

At least one AI CLI tool installed:
- [Claude CLI](https://docs.anthropic.com/claude/docs/claude-cli) (`claude`)
- [OpenAI Codex CLI](https://github.com/openai/codex) (`codex`)
- [Gemini CLI](https://cloud.google.com/vertex-ai/docs/generative-ai/start/quickstarts/quickstart-cli) (`gemini`)
- [Qwen CLI](https://github.com/QwenLM/Qwen) (`qwen`)

### Install from Source

```bash
git clone https://github.com/alienxp03/dbate.git
cd dbate

# Option 1: Install to /usr/local/bin (system-wide, may need sudo)
make install

# Option 2: Install to ~/.local/bin (user only, no sudo)
make install-user

# Option 3: Install to $GOPATH/bin
make install-gopath
```

After installation, you can run `dbate` directly from anywhere:

```bash
dbate --help
```

### Build Only (without installing)

```bash
make build
# Binaries will be in the bin/ directory
./bin/dbate --help
```

## Quick Start

```bash
# Check available providers
dbate providers

# List personas
dbate personas

# List debate styles
dbate styles

# Start a new debate
dbate new "Is AI beneficial for humanity?"

# Start with specific agents, style, and model
dbate new "Best programming language" \
  -a claude/sonnet:optimist \
  -b gemini:skeptic \
  -s adversarial \
  -t 5

# Step-by-step mode (manual control)
dbate new "Climate solutions" --step

# List all debates
dbate list

# View a specific debate
dbate show <debate-id>

# Export a debate
dbate export <debate-id> markdown
dbate export <debate-id> pdf
dbate export <debate-id> json

# Lock a debate (read-only)
dbate lock <debate-id>

# Delete a debate
dbate delete <debate-id>
```

### Web Interface

```bash
# Start the web server (default port: 8182)
dbate serve

# Or specify a custom port
dbate serve -p 3000
```

Then open http://localhost:8182 in your browser.

## Agent Configuration

Agents are specified as `provider[/model]:persona`:

```bash
# Basic format
dbate new "Topic" -a claude:optimist -b gemini:skeptic

# With specific models
dbate new "Topic" -a claude/sonnet:analyst -b qwen/qwen-max:visionary
```

**Providers:**
- `claude` - Anthropic Claude
- `codex` - OpenAI Codex
- `gemini` - Google Gemini
- `qwen` - Alibaba Qwen

**Personas:**
- `optimist` - Focuses on opportunities and positive outcomes
- `skeptic` - Questions assumptions, identifies risks
- `pragmatist` - Practical, implementable solutions
- `visionary` - Big picture, long-term thinking
- `analyst` - Data-driven, objective evaluation
- `devils_advocate` - Argues the contrarian position

## Debate Styles

- **adversarial** - Agents argue opposite sides
- **collaborative** - Work together to find best solution
- **analytical** - Systematic pros/cons evaluation
- **socratic** - One probes with questions, other defends

## Configuration

### Database Location

Default: `~/.dbate/dbate.db`

Override with `--db` flag:
```bash
dbate --db /path/to/custom.db list
```

### Config File

Create a config file at `~/.dbate/config.yaml`:

```bash
dbate config init  # Creates example config
dbate config show  # Shows current config
```

Example config:
```yaml
providers:
  claude:
    command: claude
    default_model: sonnet
    timeout: 5m
    enabled: true
  qwen:
    command: qwen
    enabled: true

defaults:
  style: collaborative
  max_turns: 5
  provider: claude
```

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Build
make build

# Clean
make clean

# Lint (requires golangci-lint)
make lint

# Uninstall
make uninstall
```

## Architecture

```
dbate/
├── cmd/
│   ├── dbate/          # CLI entry point
│   └── server/         # Web server entry point
├── internal/
│   ├── config/         # YAML configuration
│   ├── core/           # Domain types
│   ├── engine/         # Debate orchestration
│   ├── export/         # Markdown/PDF/JSON export
│   ├── persona/        # Agent personas
│   ├── provider/       # AI provider abstractions
│   ├── storage/        # SQLite persistence
│   └── style/          # Debate styles
└── web/
    └── handlers/       # HTTP handlers & templates
```

## License

MIT
