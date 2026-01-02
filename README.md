# conclave

AI-powered debate and deliberation tool that orchestrates discussions between AI agents with diverse personas.

## Features

- **Multi-Agent Councils**: Run discussions with N agents (not just 2) using a 3-stage pipeline (Responses → Ranking → Synthesis).
- **Modern Web Interface**: A lightning-fast React 19 SPA with real-time character-by-character streaming (SSE).
- **Multi-Provider Support**: Integrates with Claude, Gemini, Qwen, and OpenAI CLI tools.
- **Dynamic Personas**: Create and manage custom AI personas (Optimist, Skeptic, Pragmatist, etc.) with unique system prompts.
- **Flexible Debate Styles**: Built-in and custom styles (Adversarial, Collaborative, Socratic) that define how agents interact.
- **Session History**: Persistent SQLite-based storage for all debates and council sessions.
- **Export**: Save your deliberations as Markdown, PDF, or JSON.

## Installation

### Prerequisites

1. **Go 1.21+** installed.
2. **Node.js & npm** (required to build the modern Web UI).
3. At least one supported AI CLI tool:
   - [Claude CLI](https://docs.anthropic.com/claude/docs/claude-cli) (`claude`)
   - [Gemini CLI](https://cloud.google.com/vertex-ai/docs/generative-ai/start/quickstarts/quickstart-cli) (`gemini`)
   - [Qwen CLI](https://github.com/QwenLM/Qwen) (`qwen`)

### Build and Install

```bash
git clone https://github.com/alienxp03/conclave.git
cd conclave

# Build both CLI and the embedded React frontend
make build

# Install to ~/.local/bin
make install
```

## Quick Start (CLI)

### 2-Agent Debates
```bash
# Basic debate (uses default agents)
conclave new "Should we use Go for our next microservice?"

# Specify agents, style, and turn limit
conclave new "The merits of functional programming" \
  -a claude/sonnet:optimist \
  -b gemini:skeptic \
  -s adversarial \
  -t 5
```

### N-Agent Councils
Councils use a multi-stage process where agents first respond, then rank each other's ideas, followed by a final synthesis.
```bash
# Start a 3-agent council
conclave new "Project Roadmap 2026" \
  --models claude:optimist,gemini:skeptic,qwen:pragmatist
```

### Manage Knowledge
```bash
conclave list                # See session history
conclave show <id>           # View turns and conclusion
conclave export <id> pdf     # Export to PDF
conclave lock <id>           # Prevent accidental deletion/modification
```

## Web Interface

Start the web server to access the modern dashboard:

```bash
conclave serve --port 8080
```

Visit `http://localhost:8080` to:
- **Watch Live**: See AI agents debate in real-time as text streams in.
- **Manage History**: Browse, search, and delete past sessions.
- **Configure**: Visually set up councils and debate parameters.

## Advanced Configuration

### Custom Personas
Create your own agents to specialize the discussion:
```bash
conclave persona create --id researcher --name "Deep Researcher" --prompt "You are a meticulous researcher..."
conclave persona list
```

### Debate Styles
Define the tone and structure of the interaction:
```bash
conclave style list
conclave style show socratic
```

### Config File
Initialize a config file at `~/.conclave/config.yaml` to set default providers, models, and timeouts:
```bash
conclave config init
conclave config show
```

## Architecture

```
conclave/
├── cmd/
│   ├── conclave/          # Main CLI entry point
│   └── server/         # Standalone web server
├── internal/
│   ├── council/        # N-agent deliberation logic
│   ├── engine/         # 2-agent debate orchestration
│   ├── provider/       # AI provider abstractions (CLI wrappers)
│   ├── storage/        # SQLite persistence layer
│   └── ...             # Core, Config, Export, Persona, Style
├── web/
│   ├── app/            # Modern React 19 Frontend (Vite + TS + Tailwind)
│   └── handlers/       # Go HTTP handlers and SSE streaming logic
└── bin/                # Compiled binaries
```

## Development

```bash
# Run frontend dev server with hot reload
make dev-frontend

# Run backend with auto-reload (requires air)
make dev-serve

# Run full test suite
make test
```

## License 

MIT

```