# Council CLI Design (Simplified)

## Philosophy

**Simple, composable, no flags unless necessary.**

## Member Specification Format

```
provider[/model][:persona]
```

### Examples

| Input | Provider | Model | Persona |
|-------|----------|-------|---------|
| `claude` | claude | sonnet (default) | auto-assigned |
| `claude:optimist` | claude | sonnet (default) | optimist |
| `claude/opus` | claude | opus | auto-assigned |
| `claude/opus:optimist` | claude | opus | optimist |
| `gemini/pro:skeptic` | gemini | pro | skeptic |

## Basic Commands

### Create Council

```bash
# Minimal - 4 members, auto-assign personas
conclave new "Should we adopt GraphQL?" --models claude,gemini,qwen,codex

# Output:
Creating council with 4 members:
  • claude (Optimist)
  • gemini (Skeptic)
  • qwen (Pragmatist)
  • codex (Analyst)
  • Chairman: claude/opus

Running council...
[Stage 1/3] Collecting responses... ━━━━━━━━━━━━━━━━━━━━ 100%
[Stage 2/3] Collecting rankings... ━━━━━━━━━━━━━━━━━━━━ 100%
[Stage 3/3] Synthesizing... ━━━━━━━━━━━━━━━━━━━━━━━━━ 100%

Council ID: a3f9d8c1
```

### Specify Personas

```bash
conclave new "topic" --models claude:optimist,gemini:skeptic,qwen:pragmatist,codex:analyst
```

### Specify Models

```bash
conclave new "topic" --models claude/opus,gemini/pro,qwen/max,codex/gpt4
```

### Full Control

```bash
conclave new "topic" \
  --models claude/opus:optimist,gemini/pro:skeptic,qwen/max:pragmatist,codex/gpt4:analyst \
  --chairman claude/opus
```

### Mixed Specification

```bash
# Some with models, some with personas, some with both
conclave new "topic" --models claude/opus:optimist,gemini:skeptic,qwen/max,codex
```

## Flags

### Required: --models

```bash
--models <comma-separated-list>

# Format: provider[/model][:persona],provider[/model][:persona],...

# Examples:
--models claude,gemini,qwen,codex
--models claude:optimist,gemini:skeptic,qwen:pragmatist
--models claude/opus,gemini/pro,qwen/max
--models claude/opus:optimist,gemini/pro:skeptic,qwen/max:pragmatist
```

### Optional: --chairman

```bash
--chairman <provider[/model]>

# Examples:
--chairman claude           # Uses claude with default model
--chairman claude/opus      # Uses claude with opus model
--chairman gemini/pro       # Uses gemini with pro model

# Default: First member's provider with upgraded model
# If members include 'claude', chairman = 'claude/opus'
```

### Optional: --no-run

```bash
--no-run    # Create council but don't run (run manually later)

# Example:
conclave new "topic" --models claude,gemini,qwen --no-run
# Council ID: abc123
conclave run abc123   # Run it later
```

### Future flags (extensibility)

```bash
--max-turns <n>          # Limit response length
--temperature <0.0-1.0>  # Control randomness
--parallel <bool>        # Force sequential if needed
--format <json|yaml>     # Output format
--save-to <file>         # Auto-export
```

## Persona Auto-Assignment

If personas not specified, assign in order:

```go
var defaultPersonaOrder = []string{
    "optimist",
    "skeptic",
    "pragmatist",
    "analyst",
    "visionary",
    "devils_advocate",
}

// 2 members: optimist, skeptic
// 3 members: optimist, skeptic, pragmatist
// 4 members: optimist, skeptic, pragmatist, analyst
// etc.
```

## Model Defaults

Each provider has a default model:

```go
var providerDefaults = map[string]string{
    "claude": "sonnet",
    "gemini": "pro",
    "qwen":   "max",
    "codex":  "gpt4",
}
```

## Chairman Defaults

If `--chairman` not specified:

```go
// 1. Find first member's provider
// 2. Upgrade to best model
// 3. Use that as chairman

func defaultChairman(members []Agent) Agent {
    if len(members) == 0 {
        return Agent{Provider: "claude", Model: "opus"}
    }

    provider := members[0].Provider
    bestModel := getBestModel(provider)

    return Agent{
        Provider: provider,
        Model:    bestModel,
    }
}

var bestModels = map[string]string{
    "claude": "opus",
    "gemini": "pro",
    "qwen":   "max",
    "codex":  "gpt4",
}
```

## Other Commands

### List

```bash
conclave list

# Output:
ID        Topic                          Members  Status      Created
a3f9d8c1  Should we adopt GraphQL?       4        completed   2 hours ago
b7e2c4f3  Microservices vs Monolith      3        in_progress 1 day ago
```

### Show

```bash
conclave show <id>

# Output:
Council: a3f9d8c1
Topic: Should we adopt GraphQL?
Status: completed
Created: 2024-01-15 14:23:00

Members:
  • claude/sonnet (Optimist)
  • gemini/pro (Skeptic)
  • qwen/max (Pragmatist)
  • codex/gpt4 (Analyst)

Chairman: claude/opus

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

STAGE 1: RESPONSES

[Optimist - claude/sonnet]
GraphQL offers significant benefits...

[Skeptic - gemini/pro]
However, there are concerns about...

[Pragmatist - qwen/max]
From an implementation perspective...

[Analyst - codex/gpt4]
Looking at the data...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

STAGE 2: RANKINGS

Aggregate Rankings (by quality):
1. Analyst (codex/gpt4)    - Avg: 1.5
2. Pragmatist (qwen/max)   - Avg: 2.3
3. Skeptic (gemini/pro)    - Avg: 2.8
4. Optimist (claude/sonnet) - Avg: 3.0

Individual Rankings:
  claude: [4, 2, 3, 1]  (ranked Analyst best, self worst)
  gemini: [3, 1, 2, 4]  (ranked Skeptic best)
  qwen:   [2, 3, 1, 4]  (ranked Pragmatist best)
  codex:  [1, 4, 2, 3]  (ranked Analyst best)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

STAGE 3: SYNTHESIS

[Chairman - claude/opus]

After reviewing all perspectives and rankings, here's the synthesis:

The Analyst's data-driven approach was ranked highest, showing that
GraphQL adoption depends on specific use case metrics. The Pragmatist
correctly identified implementation challenges that the Optimist
underweighted. The Skeptic raised valid concerns about complexity.

Recommendation: Adopt GraphQL for new services where flexibility
is needed, but maintain REST for stable legacy APIs. The transition
cost identified by the Pragmatist is real and should be budgeted.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Export

```bash
conclave export <id> [format]

# Formats: markdown (default), pdf, json
conclave export a3f9d8c1 pdf
```

### Delete

```bash
conclave delete <id>
```

### Run (if created with --no-run)

```bash
conclave run <id>
```

## Parsing Logic

```go
type MemberSpec struct {
    Provider string
    Model    string  // optional
    Persona  string  // optional
}

func parseMemberSpec(spec string) (MemberSpec, error) {
    // Format: provider[/model][:persona]

    var m MemberSpec

    // Split by ':'
    parts := strings.SplitN(spec, ":", 2)

    // parts[0] = provider[/model]
    // parts[1] = persona (if exists)

    // Split provider/model
    providerParts := strings.SplitN(parts[0], "/", 2)
    m.Provider = providerParts[0]

    if len(providerParts) == 2 {
        m.Model = providerParts[1]
    }
    // else: model defaults to provider's default

    if len(parts) == 2 {
        m.Persona = parts[1]
    }
    // else: persona auto-assigned

    return m, nil
}

// Examples:
// "claude" -> {Provider: "claude", Model: "", Persona: ""}
// "claude:optimist" -> {Provider: "claude", Model: "", Persona: "optimist"}
// "claude/opus" -> {Provider: "claude", Model: "opus", Persona: ""}
// "claude/opus:optimist" -> {Provider: "claude", Model: "opus", Persona: "optimist"}
```

## Migration from Old CLI

**No backward compatibility** - clean break.

Old (2-agent debates):
```bash
conclave new "topic" -a claude:optimist -b gemini:skeptic -s adversarial
```

New (N-agent councils):
```bash
conclave new "topic" --models claude:optimist,gemini:skeptic
```

More extensible with flags, scales to N agents naturally.

## Full Example Session

```bash
# Create council
$ conclave new "Should we adopt Rust for backend services?" \
    --models claude/opus:optimist,gemini/pro:skeptic,qwen:pragmatist,codex:analyst

Creating council with 4 members...
Running council stages...
✓ Stage 1: Responses collected
✓ Stage 2: Rankings collected
✓ Stage 3: Synthesis generated

Council ID: f3a8d9c2

# View results
$ conclave show f3a8d9c2

[... full output ...]

# Export as PDF
$ conclave export f3a8d9c2 pdf

Exported to: council-f3a8d9c2.pdf

# List all councils
$ conclave list

ID        Topic                          Members  Status      Created
f3a8d9c2  Should we adopt Rust...        4        completed   5 min ago
a3f9d8c1  Should we adopt GraphQL?       4        completed   2 hours ago
```

## Advanced: Custom Personas

```bash
# Create custom persona
$ conclave persona create security_expert \
  --description "Focuses on security implications" \
  --prompt "You are a security expert. Analyze everything from a security perspective..."

# Use in council
$ conclave new "API design patterns" \
    --models claude:security_expert,gemini:optimist,qwen:pragmatist
```

## Provider Management

```bash
# List available providers
$ conclave providers

Available providers:
✓ claude  (models: sonnet, opus, haiku)
✓ gemini  (models: pro, flash)
✗ qwen    (CLI not installed)
✓ codex   (models: gpt4, gpt35)

# Check specific provider
$ conclave providers claude

Provider: claude
Display name: Anthropic Claude
Status: ✓ Available
Models: sonnet (default), opus, haiku
Command: claude
Timeout: 5m
```

## Summary

**Simple patterns:**
```bash
# Minimal
conclave new "topic" --models claude,gemini,qwen

# With personas
conclave new "topic" --models claude:optimist,gemini:skeptic,qwen:pragmatist

# With models
conclave new "topic" --models claude/opus,gemini/pro,qwen/max

# Full control
conclave new "topic" \
  --models claude/opus:optimist,gemini/pro:skeptic,qwen/max:pragmatist \
  --chairman claude/opus

# View and export
conclave show <id>
conclave export <id> pdf
```

**Extensible with flags. No complexity unless you need it.**
