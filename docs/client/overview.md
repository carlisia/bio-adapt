# Client Packages

## Overview

This directory contains idiomatic Go client libraries for bio-adapt primitives. These clients provide clean, simple APIs for common use cases while allowing advanced configuration when needed.

## Available Clients

### emerge/

Production-ready client for goal-directed synchronization.

```go
import "github.com/carlisia/bio-adapt/client/emerge"

// Simple one-liner for common goals
client := emerge.MinimizeAPICalls(scale.Medium)
client.Start(ctx)
```

[Full documentation →](emerge.md)

### navigate/ (Coming Soon)

Client for goal-directed resource allocation.

### glue/ (Planned)

Client for goal-directed collective intelligence.

## Design Philosophy

Our client libraries follow Go best practices:

- **Simple things are simple** - One-line creation for common cases
- **Complex things are possible** - Full customization available
- **Clear over clever** - Explicit, readable APIs
- **No magic** - Predictable behavior, no hidden state

## Usage Patterns

### Quick Start (Most Common)

```go
// Use predefined configurations
client := emerge.MinimizeAPICalls(scale.Large)
client := emerge.DistributeLoad(scale.Medium)
```

### Custom Configuration

```go
// Builder pattern for fine control
client := emerge.Custom().
    WithGoal(goal.MinimizeAPICalls).
    WithScale(scale.Large).
    WithTargetCoherence(0.95).
    Build()
```

### Functional Options

```go
// Alternative configuration style
client := emerge.NewWithOptions(
    emerge.WithGoalOption(goal.MinimizeAPICalls),
    emerge.WithScaleOption(scale.Large),
)
```

## Best Practices

1. **Start with presets** - Use `MinimizeAPICalls()` etc. for common cases
2. **Monitor convergence** - Check `IsConverged()` before critical operations
3. **Use contexts** - Always pass contexts with timeouts
4. **Clean shutdown** - Let contexts expire or cancel them explicitly
5. **One client per goal** - Don't reuse clients for different purposes

## Testing with Clients

```go
func TestMyFeature(t *testing.T) {
    // Use small scales for tests
    client := emerge.MinimizeAPICalls(scale.Tiny)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := client.Start(ctx)
    require.NoError(t, err)

    assert.True(t, client.IsConverged())
}
```

## Contributing

When adding new client packages:

1. Follow the emerge client as a template
2. Provide both simple and advanced APIs
3. Include comprehensive examples
4. Write clear documentation
5. Add integration tests

## License

See the [main LICENSE file](../../LICENSE) for license information.
