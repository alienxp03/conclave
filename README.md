# dbate

🎭 AI-powered debate tool that orchestrates discussions between AI agents with different personas.

## Features

- **Multi-Provider Support**: Works with Claude, OpenAI Codex, and Gemini CLI tools
- **Agent Personas**: 6 built-in personas (Optimist, Skeptic, Pragmatist, Visionary, Analyst, Devil's Advocate)
- **Debate Styles**: 4 debate formats (Adversarial, Collaborative, Analytical, Socratic)
- **CLI & Web Interface**: Use from terminal or browser
- **Session History**: SQLite-based storage for all debate sessions
- **Modular Design**: Easy to extend with new providers and personas

## Installation

### Prerequisites

At least one AI CLI tool installed:
- [Claude CLI](https://docs.anthropic.com/claude/docs/claude-cli)
- [OpenAI Codex CLI](https://github.com/openai/codex)
- [Gemini CLI](https://cloud.google.com/vertex-ai/docs/generative-ai/start/quickstarts/quickstart-cli)

### Build from Source

```bash
git clone https://github.com/alienxp03/dbate.git
cd dbate
make build
```

Binaries will be in the `bin/` directory.

## Quick Start

### CLI Usage

```bash
# Check available providers
./bin/dbate providers

# List personas
./bin/dbate personas

# List debate styles
./bin/dbate styles

# Start a new debate
./bin/dbate new "Is AI beneficial for humanity?"

# Start with specific agents and style
./bin/dbate new "Best programming language" \
  -a claude:optimist \
  -b claude:skeptic \
  -s adversarial \
  -t 5

# List all debates
./bin/dbate list

# View a specific debate
./bin/dbate show <debate-id>

# Delete a debate
./bin/dbate delete <debate-id>
```

### Web Interface

```bash
# Start the web server (default port: 8182)
./bin/dbate serve

# Or specify a custom port
./bin/dbate serve -p 3000
```

Then open http://localhost:8182 in your browser.

## Agent Configuration

Agents are specified as `provider:persona`:

**Providers:**
- `claude` - Anthropic Claude
- `codex` - OpenAI Codex
- `gemini` - Google Gemini

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

Database location: `~/.dbate/dbate.db`

Override with `--db` flag:
```bash
./bin/dbate --db /path/to/custom.db list
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
```

## Architecture

```
dbate/
├── cmd/
│   ├── dbate/          # CLI entry point
│   └── server/         # Web server entry point
├── internal/
│   ├── core/           # Domain types
│   ├── provider/       # AI provider abstractions
│   ├── persona/        # Agent personas
│   ├── style/          # Debate styles
│   ├── engine/         # Debate orchestration
│   └── storage/        # SQLite persistence
└── web/
    ├── handlers/       # HTTP handlers
    └── templates/      # HTMX templates
```

## License

MIT
