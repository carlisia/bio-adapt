# Emerge Client

A clean, idiomatic Go client for the emerge synchronization framework.

## Installation

```bash
go get github.com/carlisia/bio-adapt/client/emerge
```

## Quick Start

```go
import "github.com/carlisia/bio-adapt/client/emerge"

// Create and start a client
client, err := emerge.MinimizeAPICalls(scale.Medium)
if err != nil {
    log.Fatal(err)
}

// Start synchronization
ctx := context.Background()
if err := client.Start(ctx); err != nil {
    log.Fatal(err)
}
```

## Design Philosophy

This client follows Go idioms and best practices:

- **Simple things are simple** - One-line creation for common cases
- **Complex things are possible** - Full customization when needed
- **Clear is better than clever** - Explicit, readable API
- **No magic** - Predictable behavior, no hidden state

## Common Usage Patterns

### Scenario-Based Constructors

```go
// Optimize API calls through synchronization
client := emerge.MinimizeAPICalls(scale.Large)

// Distribute load through anti-synchronization
client := emerge.DistributeLoad(scale.Medium)

// Just use sensible defaults
client := emerge.Default()
```

### Custom Configuration

```go
// Builder pattern for fine control
client := emerge.New().
    WithGoal(goal.MinimizeAPICalls).
    WithScale(scale.Large).
    WithTargetCoherence(0.95).
    Build()

// Functional options pattern
client := emerge.NewWithOptions(
    emerge.WithGoalOption(goal.MinimizeAPICalls),
    emerge.WithScaleOption(scale.Large),
    emerge.WithCoherenceOption(0.90),
)
```

## Monitoring Synchronization

```go
// Start the client
go client.Start(ctx)

// Monitor progress
ticker := time.NewTicker(100 * time.Millisecond)
defer ticker.Stop()

for range ticker.C {
    coherence := client.Coherence()
    fmt.Printf("Coherence: %.2f%%\n", coherence*100)

    if client.IsConverged() {
        fmt.Println("Synchronization achieved!")
        break
    }
}
```

## API Reference

### Client Methods

| Method          | Description                         |
| --------------- | ----------------------------------- |
| `Start(ctx)`    | Begin synchronization process       |
| `Stop()`        | Gracefully stop synchronization     |
| `Agents()`      | Get all agents in the swarm         |
| `Coherence()`   | Current synchronization level (0-1) |
| `IsConverged()` | Check if target coherence achieved  |
| `Size()`        | Number of agents in swarm           |
| `Config()`      | Get swarm configuration             |

### Builder Methods

| Method                   | Description                      |
| ------------------------ | -------------------------------- |
| `New()`                  | Create new builder with defaults |
| `WithGoal(g)`            | Set synchronization goal         |
| `WithScale(s)`           | Set swarm scale/size             |
| `WithTargetCoherence(c)` | Set target coherence             |
| `Build()`                | Create the client                |

## Scales

Bio-adapt uses predefined scales with optimized parameters. See [Scale Definitions](../emerge/scales.md) for complete details including:

- Agent counts and memory requirements
- Convergence times and performance characteristics
- Resource requirements and optimization thresholds
- Guidelines for choosing the right scale

## Goals

Synchronization objectives that determine agent behavior:

| Goal               | Behavior             | Use Case                       |
| ------------------ | -------------------- | ------------------------------ |
| `MinimizeAPICalls` | High synchronization | Batch operations, reduce costs |
| `DistributeLoad`   | Anti-synchronization | Spread load, avoid contention  |

## Examples

### Complete Example: API Batching

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/carlisia/bio-adapt/client/emerge"
    "github.com/carlisia/bio-adapt/emerge/scale"
)

func main() {
    // Create client for API optimization
    client, err := emerge.MinimizeAPICalls(scale.Medium)
    if err != nil {
        log.Fatal(err)
    }

    // Start synchronization
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    go func() {
        if err := client.Start(ctx); err != nil {
            log.Printf("Error: %v", err)
        }
    }()

    // Monitor and use synchronization
    for {
        coherence := client.Coherence()

        if coherence > 0.8 {
            // High synchronization achieved
            // Agents can now batch their operations
            fmt.Println("Ready for batched operations!")
            performBatchedWork(client)
            break
        }

        time.Sleep(100 * time.Millisecond)
    }

    // Cleanup
    client.Stop()
}

func performBatchedWork(client *emerge.Client) {
    // Your application logic here
    // Agents are synchronized and can coordinate
}
```

## Architecture

The emerge client provides a clean abstraction over the emerge/swarm framework:

```text
Your Application
       ↓
  emerge.Client     (Simple, idiomatic API)
       ↓
  emerge/swarm      (Synchronization engine)
       ↓
  emerge/agent      (Individual oscillators)
```

## Thread Safety

All client methods are thread-safe. The underlying swarm handles synchronization internally using appropriate primitives based on swarm size.

## Performance

The framework automatically optimizes based on swarm size:

- Small swarms (<100): Flexible sync.Map storage
- Large swarms (≥100): Cache-optimized slice storage
- Enterprise (≥1000): Additional optimizations enabled

## Contributing

See the main [bio-adapt repository](https://github.com/carlisia/bio-adapt) for contribution guidelines.
