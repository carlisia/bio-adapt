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
# Run all checks
task check

# Run tests
task test

# Build simulations
task build:simulations
```

## Project structure

```bash
bio-adapt/
├── docs/            # Documentation
├── e2e/             # End-to-end integration tests
├── emerge/          # Goal-directed synchronization primitive
├── simulations/     # Interactive demonstrations
├── glue/            # Goal-directed collective intelligence (planned)
├── internal/        # Internal utilities and helpers
├── navigate/        # Goal-directed resource allocation (coming soon)
└── Taskfile.yml     # Task runner configuration
```

## Development workflow

### Using Task (recommended)

Task provides convenient commands for development:

```bash
# View all available tasks
task --list

# Common development tasks
task check          # Run all checks (fmt, vet, lint, vuln)
task test           # Run all tests
task test:e2e       # Run all e2e tests (no caching)
task test:short     # Run quick tests only
task test:coverage  # Run tests with coverage
task lint           # Run linter
task lint:fix       # Run linter with auto-fix
task fmt            # Format code
task build:all      # Build everything
task clean          # Clean build artifacts
```

### Watch mode development

If you have `entr` installed:

```bash
# Auto-rebuild on file changes
task dev

# Watch and run tests
task watch:test
```

### Manual commands (when Task is not available)

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...

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
task test:e2e

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

Use the internal testutil packages for common test helpers:

```go
import (
    "github.com/carlisia/bio-adapt/internal/testutil"
    "github.com/carlisia/bio-adapt/emerge/swarm"
)

func TestSwarmConvergence(t *testing.T) {
    // Example test utility usage
    cfg := config.AutoScaleConfig(10)
    s, _ := swarm.New(10, targetState, swarm.WithConfig(cfg))
    // Test implementation
}
```

## Benchmarking

### Running benchmarks

```bash
# All benchmarks (preferred)
task bench

# Emerge package benchmarks
task bench:emerge

# E2E benchmarks
task test:e2e:bench

# Manual: Specific package with memory profiling
go test -bench=. -benchmem ./emerge/agent

# Manual: Run specific benchmark
go test -bench=BenchmarkSwarmConvergence ./emerge/swarm

# Manual: With CPU profile
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

# View specific package docs
go doc github.com/carlisia/bio-adapt/emerge/swarm
```

## Common development tasks

### Adding a new strategy

1. Create new file in `emerge/strategy/`
2. Implement the strategy interface
3. Add tests in the same package
4. Update simulations to demonstrate usage

### Running the simulation

```bash
# Run the interactive simulation
task run:sim -- emerge

# With different scales
task run:sim -- emerge -scale=medium

# List available scales
task run:sim -- emerge -list
```

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
# Clean test cache and retry
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
# Auto-fix formatting issues
task fmt

# Auto-fix linter issues
task lint:fix

# Check specific issues manually
golangci-lint run --fix
```

## Performance tips

1. **Use benchmarks** - Measure before optimizing
2. **Profile first** - Don't guess bottlenecks
3. **Minimize allocations** - Reuse objects when possible
4. **Batch operations** - Group related updates
5. **Concurrent safe** - But avoid over-synchronization
6. **Cache computations** - Especially in hot paths
