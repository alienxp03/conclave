# Provider - Reusable AI CLI Wrapper

A Go package for wrapping AI command-line tools (Claude, Gemini, OpenAI, etc.) with a unified interface.

## Features

- **Unified Interface**: Single API for multiple AI providers
- **Request/Response Pattern**: Clean, explicit request handling
- **JSON Parsing**: Automatic parsing of provider-specific JSON formats
- **Metadata Tracking**: Token usage, duration, and other statistics
- **Timeout Support**: Configurable timeouts for CLI commands
- **Error Handling**: Structured error types with context
- **Working Directory Support**: Run commands in specific directories
- **Extensible**: Easy to add custom providers

## Installation

```bash
go get github.com/alienxp03/conclave/provider
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/alienxp03/conclave/provider"
    "github.com/alienxp03/conclave/provider/claude"
)

func main() {
    // Create a Claude provider
    p := claude.New(provider.Config{
        Name:    "claude",
        Command: "claude",
        Args:    []string{"--print"},
    })

    // Check if the CLI is available
    if !p.Available() {
        log.Fatal("Claude CLI not found")
    }

    // Execute a request
    resp, err := p.Execute(context.Background(), &provider.Request{
        Prompt: "What is a monorepo?",
        Model:  "sonnet",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Response:", resp.Content)
    fmt.Printf("Tokens: %d input, %d output\n",
        resp.Metadata.InputTokens,
        resp.Metadata.OutputTokens)
}
```

### Using the Registry

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/alienxp03/conclave/provider"
    "github.com/alienxp03/conclave/provider/claude"
    "github.com/alienxp03/conclave/provider/gemini"
    "github.com/alienxp03/conclave/provider/openai"
)

func main() {
    // Create a registry
    registry := provider.NewRegistry()

    // Register providers
    registry.Register(claude.New(provider.Config{
        Name:         "claude",
        Command:      "claude",
        Args:         []string{"--print"},
        DefaultModel: "sonnet",
    }))

    registry.Register(gemini.New(provider.Config{
        Name:         "gemini",
        Command:      "gemini",
        DefaultModel: "flash",
    }))

    registry.Register(openai.New(provider.Config{
        Name:         "openai",
        Command:      "codex",
        DefaultModel: "gpt-4",
    }))

    // Get a provider by name
    p, err := registry.Get("claude")
    if err != nil {
        log.Fatal(err)
    }

    // Use the provider
    resp, err := p.Execute(context.Background(), &provider.Request{
        Prompt: "Hello!",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
```

### Using the Factory

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/alienxp03/conclave/provider"
)

func main() {
    // Create a provider using the factory
    p, err := provider.Factory("claude", provider.Config{
        Command:      "claude",
        Args:         []string{"--print"},
        DefaultModel: "sonnet",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Execute request
    resp, err := p.Execute(context.Background(), &provider.Request{
        Prompt:     "Explain dependency injection",
        WorkingDir: "/path/to/project",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Provider: %s\n", resp.Provider)
    fmt.Printf("Model: %s\n", resp.Model)
    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Duration: %v\n", resp.Metadata.Duration)
}
```

## Supported Providers

| Provider | Package | CLI Command | Notes |
|----------|---------|-------------|-------|
| Claude | `provider/claude` | `claude` | Anthropic's Claude CLI |
| Gemini | `provider/gemini` | `gemini` | Google's Gemini CLI |
| OpenAI | `provider/openai` | `codex` | OpenAI CLI (codex) |
| Qwen | `provider/qwen` | `qwen` | Qwen CLI |
| Opencode | `provider/opencode` | `opencode` | Opencode CLI |
| Generic | `provider/generic` | custom | For any CLI tool |

## API Reference

### Core Types

#### Provider Interface

```go
type Provider interface {
    Name() string
    Available() bool
    Execute(ctx context.Context, req *Request) (*Response, error)
}
```

#### Request

```go
type Request struct {
    Prompt     string   // Required: input text
    Model      string   // Optional: model name
    WorkingDir string   // Optional: working directory
    Args       []string // Optional: additional CLI args
}
```

#### Response

```go
type Response struct {
    Content  string    // AI-generated response
    Model    string    // Model used
    Provider string    // Provider name
    Metadata *Metadata // Usage statistics
    Raw      string    // Raw CLI output
}
```

#### Metadata

```go
type Metadata struct {
    InputTokens  int
    OutputTokens int
    TotalTokens  int
    Duration     time.Duration
    StopReason   string
    SessionID    string
}
```

#### Config

```go
type Config struct {
    Name         string        // Provider identifier
    DisplayName  string        // Human-friendly name
    Command      string        // CLI executable
    Args         []string      // Default arguments
    DefaultModel string        // Default model
    Models       []string      // Available models
    Timeout      time.Duration // Command timeout
}
```

### Registry Methods

```go
func NewRegistry() *Registry
func (r *Registry) Register(p Provider)
func (r *Registry) Get(name string) (Provider, error)
func (r *Registry) Has(name string) bool
func (r *Registry) List() []Provider
func (r *Registry) Available() []Provider
func (r *Registry) Names() []string
```

### Factory Methods

```go
func Factory(name string, cfg Config) (Provider, error)
func MustFactory(name string, cfg Config) Provider
```

## Creating Custom Providers

### Using the Generic Provider

```go
import "github.com/alienxp03/conclave/provider/generic"

p := generic.New(provider.Config{
    Name:    "my-ai",
    Command: "my-ai-cli",
    Args:    []string{"--json"},
})
```

### Implementing a Custom Provider

```go
package myprovider

import (
    "context"
    "github.com/alienxp03/conclave/provider"
)

type MyProvider struct {
    provider.BaseProvider
}

func New(cfg provider.Config) *MyProvider {
    return &MyProvider{
        BaseProvider: provider.NewBaseProvider(cfg),
    }
}

func (p *MyProvider) Execute(ctx context.Context, req *provider.Request) (*provider.Response, error) {
    // Build CLI arguments
    args := []string{"--output-json"}
    if req.Model != "" {
        args = append(args, "--model", req.Model)
    }
    args = append(args, req.Prompt)

    // Execute command
    execReq := &provider.Request{
        Prompt:     req.Prompt,
        Model:      req.Model,
        WorkingDir: req.WorkingDir,
        Args:       args,
    }

    rawOutput, err := p.ExecuteCommand(ctx, execReq)
    if err != nil {
        return nil, err
    }

    // Parse response (custom logic here)
    return &provider.Response{
        Content:  rawOutput,
        Provider: p.Name(),
        Model:    req.Model,
    }, nil
}
```

## Error Handling

```go
resp, err := provider.Execute(ctx, req)
if err != nil {
    if cliErr, ok := err.(*provider.CLIError); ok {
        fmt.Printf("Provider: %s\n", cliErr.Provider)
        fmt.Printf("Message: %s\n", cliErr.Message)
        fmt.Printf("Underlying: %v\n", cliErr.Unwrap())
    }
    return err
}
```

## License

MIT
