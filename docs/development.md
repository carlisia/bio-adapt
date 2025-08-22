# Development guide

## Overview

This guide covers setting up your development environment, building, testing, and contributing to bio-adapt.

## Prerequisites

### Required tools

- Go 1.21+ (1.22+ recommended)
- Task (task runner) - [Install guide](https://taskfile.dev/installation/)
- Git

### Optional tools

- golangci-lint - For linting
- entr - For watch mode development
- Go tools:
  - `go install golang.org/x/tools/cmd/goimports@latest`
  - `go install github.com/daixiang0/gci@latest`

## Getting started

### Clone the repository

```bash
git clone https://github.com/carlisia/bio-adapt
cd bio-adapt
```

### Install dependencies

```bash
go mod download
go mod tidy
```

### Verify setup

```bash
# Run tests
task test

# Run linter
task lint

# Build examples
task build:examples
```

## Project structure

```
bio-adapt/
├── emerge/           # Main synchronization package
│   ├── agent/       # Agent implementation
│   ├── swarm/       # Swarm coordination
│   ├── core/        # Core types and interfaces
│   └── ...
├── internal/        # Internal utilities
├── examples/        # Usage examples
├── e2e/            # End-to-end tests
├── docs/           # Documentation
├── scripts/        # Build and utility scripts
└── Taskfile.yml    # Task definitions
```

## Development workflow

### Using Task (recommended)

Task provides convenient commands for development:

```bash
# View all available tasks
task --list

# Common development tasks
task test          # Run all tests
task test:short    # Run quick tests only
task test:coverage # Run tests with coverage
task lint         # Run linter
task fmt          # Format code
task build:all    # Build everything
task clean        # Clean build artifacts
```

### Watch mode development

If you have `entr` installed:

```bash
# Auto-rebuild on file changes
task dev

# Watch specific package
find emerge -name "*.go" | entr -r go test ./emerge/...
```

### Manual commands

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./emerge/...

# Build
go build ./...

# Format code
go fmt ./...
goimports -w .
```

## Testing

### Test structure

Tests are organized by package:

- Unit tests: `*_test.go` files in each package
- Benchmarks: `*_bench_test.go` files
- E2E tests: `e2e/` directory

### Running tests

```bash
# All tests
task test

# Specific package
go test ./emerge/agent

# With coverage
task test:coverage

# Benchmarks only
task bench

# E2E tests only
go test ./e2e -v

# With race detector
go test -race ./...
```

### Writing tests

Example test structure:

```go
package agent_test

import (
    "testing"
    "github.com/carlisia/bio-adapt/emerge/agent"
)

func TestAgent_UpdatePhase(t *testing.T) {
    // Arrange
    a := agent.New("test-1")
    initialPhase := a.Phase()

    // Act
    a.SetPhase(1.5)

    // Assert
    if a.Phase() == initialPhase {
        t.Error("phase should have changed")
    }
}

func BenchmarkAgent_UpdateContext(b *testing.B) {
    a := agent.New("bench-1")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        a.UpdateContext()
    }
}
```

### Test utilities

Use the testutil package for common test helpers:

```go
import "github.com/carlisia/bio-adapt/internal/testutil/emerge"

func TestSwarmConvergence(t *testing.T) {
    swarm := emerge.NewTestSwarm(t, 10)
    emerge.WaitForCoherence(t, swarm, 0.9, 5*time.Second)
}
```

## Code style

### Go conventions

Follow standard Go conventions:

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable names
- Keep functions small and focused

### Project conventions

```go
// Package comments describe the package purpose
// Package agent implements goal-directed autonomous agents.
package agent

// Exported types have descriptive comments
// Agent represents an autonomous agent in the swarm.
type Agent struct {
    // Exported fields are documented
    ID string // Unique identifier for the agent

    // unexported fields use camelCase
    phase float64
}

// Method comments describe what the method does
// UpdatePhase adjusts the agent's phase based on neighbor states.
func (a *Agent) UpdatePhase(neighbors []*Agent) {
    // Implementation
}
```

### Error handling

```go
// Return errors with context
func NewSwarm(size int, goal State) (*Swarm, error) {
    if size <= 0 {
        return nil, fmt.Errorf("invalid swarm size: %d", size)
    }

    if err := validateGoal(goal); err != nil {
        return nil, fmt.Errorf("invalid goal: %w", err)
    }

    // ...
}

// Check errors immediately
swarm, err := NewSwarm(100, goal)
if err != nil {
    return fmt.Errorf("failed to create swarm: %w", err)
}
```

## Benchmarking

### Running benchmarks

```bash
# All benchmarks
task bench

# Specific package
go test -bench=. ./emerge/agent

# With memory profiling
go test -bench=. -benchmem ./emerge/agent

# Run specific benchmark
go test -bench=BenchmarkSwarmConvergence ./emerge/swarm

# With CPU profile
go test -bench=. -cpuprofile=cpu.prof ./emerge/agent
go tool pprof cpu.prof
```

### Writing benchmarks

```go
func BenchmarkSwarmConvergence(b *testing.B) {
    sizes := []int{10, 50, 100, 500, 1000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
            b.ResetTimer()

            for i := 0; i < b.N; i++ {
                swarm, _ := NewSwarm(size, testGoal)
                ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
                swarm.Run(ctx)
                cancel()
            }
        })
    }
}
```

## Debugging

### Debug logging

```go
// Enable debug logging in tests
func TestDebugScenario(t *testing.T) {
    if testing.Verbose() {
        slog.SetLogLoggerLevel(slog.LevelDebug)
    }

    // Test code
}
```

### Using delve debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a test
dlv test ./emerge/agent -- -test.run TestAgent_UpdatePhase

# Debug an example
dlv debug ./examples/emerge/basic_sync
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./emerge/swarm
go tool pprof -http=:8080 cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./emerge/swarm
go tool pprof -http=:8080 mem.prof

# Trace analysis
go test -trace=trace.out ./emerge/swarm
go tool trace trace.out
```

## Building

### Build commands

```bash
# Build all packages
task build:all

# Build examples only
task build:examples

# Build specific example
go build -o bin/basic_sync ./examples/emerge/basic_sync

# Build with optimizations
go build -ldflags="-s -w" ./...

# Cross-compilation
GOOS=linux GOARCH=amd64 go build ./...
GOOS=darwin GOARCH=arm64 go build ./...
```

### Release builds

```bash
# Create optimized release build
go build -ldflags="-s -w" -trimpath -o bio-adapt ./cmd/bio-adapt

# With version information
VERSION=$(git describe --tags --always)
go build -ldflags="-s -w -X main.version=$VERSION" ./...
```

## Documentation

### GoDoc comments

Write clear GoDoc comments:

```go
// MeasureCoherence calculates the Kuramoto order parameter for the swarm.
// It returns a value between 0 (no synchronization) and 1 (perfect synchronization).
//
// The calculation uses the formula:
//   R = |Σ(e^(iθ))| / N
//
// where θ is the phase of each agent and N is the number of agents.
func (s *Swarm) MeasureCoherence() float64 {
    // Implementation
}
```

### Generating documentation

```bash
# View documentation locally
go doc -http=:6060

# Generate markdown docs (future)
task docs:generate
```

## Common development tasks

### Adding a new strategy

1. Create new file in `emerge/strategy/`
2. Implement the `DecisionMaker` interface
3. Add tests in `emerge/strategy/strategy_test.go`
4. Update examples to demonstrate usage

### Adding a new example

1. Create directory in `examples/emerge/`
2. Add `main.go` with example code
3. Update `examples/README.md`
4. Test with `task run:example -- your_example`

### Performance optimization

1. Identify bottleneck with profiling
2. Write benchmark before optimizing
3. Implement optimization
4. Verify improvement with benchmark
5. Add regression test

## Troubleshooting

### Common issues

**Tests failing**

```bash
# Clean and retry
task clean
go mod tidy
task test
```

**Import errors**

```bash
# Update dependencies
go mod download
go mod tidy
```

**Linter errors**

```bash
# Auto-fix some issues
task fmt
goimports -w .

# Check specific issues
golangci-lint run --fix
```

## IDE setup

### VS Code

`.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "go.testTimeout": "30s"
}
```

### GoLand/IntelliJ

1. Enable goimports on save
2. Configure golangci-lint as external tool
3. Set test timeout to 30s
4. Enable race detector for tests

## Performance tips

1. **Use benchmarks** - Measure before optimizing
2. **Profile first** - Don't guess bottlenecks
3. **Minimize allocations** - Reuse objects when possible
4. **Batch operations** - Group related updates
5. **Concurrent safe** - But avoid over-synchronization
6. **Cache computations** - Especially in hot paths

