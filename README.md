# conclave

AI-powered multi-agent deliberation platform that orchestrates debates and council discussions between AI agents with diverse personas.

![conclave UI](docs/screenshot.png)

> **Note:** This project is 99.9% vibe coded. Expect rough edges, experimental features, and the occasional surprise.

**Based on:** [karpathy/llm-council](https://github.com/karpathy/llm-council) — reimplemented to work with existing CLI coding agents (Claude, Gemini, Qwen, etc.) instead of direct API calls.

## Features

- **Multi-Agent Councils** — Run discussions with N agents using a 3-stage pipeline (Responses → Ranking → Synthesis)
- **Real-Time Streaming UI** — React 19 SPA with character-by-character SSE streaming
- **Multi-Provider Support** — Works with Claude, Gemini, Qwen, OpenCode, and other AI CLI tools
- **Custom Personas** — Create AI agents with unique personalities (Optimist, Skeptic, Pragmatist, etc.)
- **Debate Styles** — Choose from Adversarial, Collaborative, Socratic, or define your own
- **Session History** — SQLite persistence for all debates and councils
- **Export Options** — Save deliberations as Markdown, PDF, or JSON

## Installation

**Prerequisites:** Go 1.21+, Node.js/npm, and at least one AI CLI tool ([Claude](https://docs.anthropic.com/claude/docs/claude-cli), [Gemini](https://cloud.google.com/vertex-ai/docs/generative-ai/start/quickstarts/quickstart-cli), [Qwen](https://github.com/QwenLM/Qwen), etc.)

```bash
git clone https://github.com/alienxp03/conclave.git
cd conclave
make build
make install
```

## Quick Start

Start the web server:

```bash
conclave serve --port 8080
```

Open `http://localhost:8080` in your browser to:
- **Create Councils** — Set up multi-agent deliberations with custom personas
- **Watch Live** — Real-time streaming debates with character-by-character rendering
- **Browse History** — View, search, and manage past sessions
- **Export** — Download sessions as Markdown, PDF, or JSON

## Configuration

Optionally configure default settings:

```bash
conclave config init    # Create ~/.conclave/config.yaml
conclave config show    # View current configuration
```

## Architecture

```
conclave/
├── cmd/              # CLI and server entry points
├── internal/
│   ├── council/      # N-agent deliberation logic
│   ├── engine/       # 2-agent debate orchestration
│   ├── provider/     # AI provider abstractions (CLI wrappers)
│   ├── storage/      # SQLite persistence
│   └── workspace/    # Project workspace management
├── web/
│   ├── app/          # React 19 frontend (Vite + TS + Tailwind)
│   └── handlers/     # HTTP handlers and SSE streaming
```

## Development

```bash
make dev-frontend     # Frontend dev server with hot reload
make dev-serve        # Backend with auto-reload (requires air)
make test             # Run test suite
```

## License

MIT