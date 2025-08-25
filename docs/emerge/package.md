# Emerge Package

## Overview

The emerge package implements goal-directed synchronization for distributed systems. It enables agents to achieve target coordination states through adaptive strategies, inspired by biological systems that maintain goals despite disruptions.

## Package Structure

```
emerge/
├── agent/          # Core agent implementation with optimizations
├── swarm/          # Swarm coordination and convergence
├── core/           # Fundamental types and interfaces
├── strategy/       # Multiple synchronization strategies
├── goal/           # Goal configuration and management
├── scale/          # Scale definitions (Tiny to Huge)
├── monitoring/     # Metrics and monitoring
└── decision/       # Decision-making engines
```

## Quick Start

For most users, we recommend using the client API:

```go
import (
    "github.com/carlisia/bio-adapt/client/emerge"
    "github.com/carlisia/bio-adapt/emerge/scale"
)

// Simple usage with client
client := emerge.MinimizeAPICalls(scale.Medium)
err := client.Start(ctx)
```

For advanced users who need direct swarm control:

```go
import (
    "github.com/carlisia/bio-adapt/emerge/swarm"
    "github.com/carlisia/bio-adapt/emerge/goal"
)

// Direct swarm usage
cfg := swarm.For(goal.MinimizeAPICalls).With(scale.Large)
s, err := swarm.New(1000, targetState, swarm.WithGoalConfig(cfg))
s.Run(ctx)
```

## Core Concepts

### Agents

Individual units that maintain phase, frequency, and energy. Agents interact with neighbors to achieve synchronization. See [Agents](../concepts/agents.md) for detailed concepts.

### Swarms

Collections of agents that work together to achieve target states. Swarms automatically optimize based on size. See [Swarm](../concepts/swarm.md) for more details.

### Scales

Predefined configurations optimized for different swarm sizes. See [Scale Definitions](scales.md) for detailed information about each scale including performance characteristics and resource requirements.

### Goals

High-level objectives that determine swarm behavior. See [Goals](../concepts/goals.md) for detailed descriptions and [Goal-Directed](goal-directed.md) for how emerge pursues these goals:

- **MinimizeAPICalls** - Batch operations to reduce costs
- **DistributeLoad** - Spread work evenly across agents
- **ReachConsensus** - Achieve distributed agreement
- **MinimizeLatency** - Optimize for speed
- **SaveEnergy** - Minimize resource consumption
- **MaintainRhythm** - Keep consistent timing
- **RecoverFromFailure** - Self-healing behavior
- **AdaptToTraffic** - Respond to load changes

### Strategies

Multiple pathways to achieve synchronization goals. See [Strategies](../concepts/strategies.md) for detailed descriptions and [Protocol](protocol.md) for how strategies are applied:

- **PhaseNudge** - Gentle phase adjustments
- **FrequencyLock** - Align oscillation speeds
- **PulseCoupling** - Strong synchronization bursts
- **EnergyAware** - Resource-conscious coordination
- **Adaptive** - Automatic strategy selection

## Performance Optimizations

The emerge package includes several performance optimizations that activate automatically. See [Optimization](optimization.md) for detailed benchmarks:

### For swarms >100 agents:

- Fixed-size array storage for better cache locality
- Grouped atomic fields to reduce cache line bouncing
- Pre-allocated memory pools

### Benchmark Results:

- Field access: **62% faster** with grouped atomics
- Neighbor iteration: **45% faster** with array storage
- Convergence: Sub-linear scaling with swarm size

## Testing

```bash
# Run unit tests
go test ./emerge/...

# Run benchmarks
go test -bench=. ./emerge/agent
go test -bench=. ./emerge/swarm

# Run with race detector
go test -race ./emerge/...
```

## Documentation

- [Architecture](architecture.md) - Detailed architecture documentation
- [Optimization Guide](optimization.md) - Performance optimization details
- [Primitive Guide](primitive.md) - High-level usage patterns
- [API Reference](https://pkg.go.dev/github.com/carlisia/bio-adapt/emerge) - Complete API documentation

## Contributing

When contributing to the emerge package:

1. Maintain backward compatibility
2. Add tests for new features
3. Run benchmarks to verify performance
4. Update documentation as needed
5. Follow existing code patterns

## License

See the [main LICENSE file](../../LICENSE) for license information.

