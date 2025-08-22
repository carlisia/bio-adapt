# Emerge: Goal-Directed Synchronization

**Temporal coordination through adaptive strategies** - Systems that pursue synchronization targets via multiple pathways, switching strategies when defaults fail to achieve coordination goals.

## Why Emerge?

Traditional synchronization requires explicit coordination logic. Emerge lets you specify WHAT you want (target state) and automatically figures out HOW to achieve it through adaptive strategies.

### Real-World Problems It Solves

- **API Request Batching** - Reduce API calls by 80% through automatic coordination
- **Load Distribution** - Spread work across workers without central control
- **Connection Pooling** - Optimize database connections adaptively
- **Task Scheduling** - Coordinate concurrent tasks without explicit locks
- **Self-Healing Systems** - Maintain service levels despite failures

## Quick Start

```go
import "github.com/carlisia/bio-adapt/emerge/swarm"

// Simple: Use a preset for your goal
cfg := swarm.For(goal.MinimizeAPICalls)
swarm, _ := swarm.New(50, core.State{
    Phase:     0,
    Frequency: 200 * time.Millisecond,
    Coherence: 0.9,  // Target: 90% synchronization
}, swarm.WithGoalConfig(cfg))
swarm.Run(ctx)  // Pursues goal through multiple strategies
```

## Configuration Options

### 1. Presets (Goal-Based) - Recommended

```go
// Fluent builder API chains goal → trait → scale
cfg := swarm.For(goal.MinimizeAPICalls).
    TuneFor(trait.Stability).
    With(scale.Large)

swarm, _ := swarm.New(1000, targetState, swarm.WithGoalConfig(cfg))
```

### 2. Adaptive (Size-Based)

```go
// Automatically optimizes for swarm size
config := config.AutoScaleConfig(1000)
swarm, _ := swarm.New(1000, targetState, swarm.WithConfig(config))
```

### 3. Custom (Manual)

```go
customConfig := config.Swarm{
    CouplingStrength:      0.8,
    MinNeighbors:          3,
    MaxNeighbors:          10,
    ConnectionProbability: 0.15,
    UpdateInterval:        50 * time.Millisecond,
    UseBatchProcessing:    true,    // Enable for large swarms
    MaxConcurrentAgents:   100,     // Limit concurrent updates
}
swarm, _ := swarm.New(500, targetState, swarm.WithConfig(customConfig))
```

## Core Concepts

### Goal State

The target configuration you want the system to achieve:

- **Phase**: Synchronization point (0 to 2π)
- **Frequency**: How often agents act
- **Coherence**: Synchronization level (0=chaos, 1=perfect)

### Adaptive Strategies

System automatically switches between strategies to reach goals:

- **PhaseNudge**: Gentle adjustments for stability
- **FrequencyLock**: Align oscillation speeds
- **PulseCoupling**: Strong synchronization bursts
- **EnergyAware**: Resource-constrained coordination
- **Adaptive**: Context-aware strategy selection

### Energy Constraints

Realistic resource limits that prevent infinite adaptation:

```go
agent.SetEnergy(100)                // Starting energy
agent.SetEnergyRecoveryRate(5)      // Units per second
agent.SetMinEnergyThreshold(20)     // Minimum to act
```

## Use Case Examples

### API Request Batching

```go
// Goal: Minimize API calls through coordinated batching
config := swarm.For(goal.MinimizeAPICalls)
s, _ := swarm.New(50, core.State{
    Phase:     0,
    Frequency: 200 * time.Millisecond,
    Coherence: 0.9,  // High synchronization for batching
}, swarm.WithGoalConfig(config))
s.Run(ctx)
// Result: 80% reduction in API calls
```

### Load Distribution

```go
// Goal: Spread work evenly (anti-synchronization)
config := swarm.For(goal.DistributeLoad)
s, _ := swarm.New(100, core.State{
    Phase:     0,
    Frequency: 1 * time.Hour,
    Coherence: 0.1,  // Low coherence for distribution
}, swarm.WithGoalConfig(config))
s.Run(ctx)
// Result: Even load distribution without central scheduler
```

### Database Connection Pooling

```go
// Goal: Balance connection reuse and distribution
config := swarm.For(goal.OptimizeConnections)
s, _ := swarm.New(200, core.State{
    Phase:     0,
    Frequency: 50 * time.Millisecond,
    Coherence: 0.7,  // Moderate clustering
}, swarm.WithGoalConfig(config))
s.Run(ctx)
// Result: Adaptive connection management under varying load
```

## Architecture

```text
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Agent 1   │◀────▶│   Agent 2   │◀────▶│   Agent 3   │
│   φ=0.2π    │     │   φ=0.3π    │     │   φ=0.25π   │
└─────────────┘     └─────────────┘     └─────────────┘
        ▲                   ▲                   ▲
        └───────────────────┼───────────────────┘
                   Adaptive Strategies
                            │
                            ▼
                ┌─────────────────────────┐
                │      Goal State         │
                │   Target: Coherence=0.9 │
                │   Multiple Paths →      │
                └─────────────────────────┘
```

### Package Structure

```bash
emerge/
├── agent/          # Core agent implementation with optimizations
├── swarm/          # Swarm coordination and management
├── core/           # Fundamental types and interfaces
├── strategy/       # Adaptive strategy implementations
├── goal/           # Goal definitions and management
├── monitoring/     # Convergence tracking
└── decision/       # Strategy selection engines
```

## Advanced Features

### Disruption Recovery

```go
swarm.DisruptAgents(0.3)  // Disrupt 30% of agents
// System automatically finds alternative paths to target
// Recovery means achieving the original goal, not just stability
```

### Performance Optimizations

The system automatically optimizes based on scale:

- **Small swarms (≤100)**: Simple sync.Map for flexibility
- **Large swarms (>100)**: Fixed arrays for cache locality
- **Atomic grouping**: Related fields share cache lines
- **Batch processing**: Configurable for massive swarms

### Performance Characteristics

- **Convergence**: Sub-linear scaling with swarm size
- **Memory**: ~2-3KB per agent at scale
- **Fault tolerance**: Automatic multi-path recovery
- **Network overhead**: O(log N) coordination traffic

## Theory Foundation

Emerge combines mathematical models with goal-directed principles:

- **Kuramoto Model**: Mathematical foundation for phase synchronization
- **Goal-Directedness**: Systems maintain targets as invariants (Levin's research)
- **Attractor Basins**: Multiple convergence paths to same target
- **Adaptive Navigation**: Switch strategies when progress stalls

## API Stability

Production-ready with semantic versioning:

- Core interfaces (`Agent`, `Swarm`) are stable
- Strategy interfaces support custom implementations
- Internal optimizations are transparent to users

## Contributing

We welcome contributions to make emerge even better! Areas of interest:

- Performance optimizations for massive swarms (10,000+ agents)
- New adaptive strategies for goal achievement
- Integration with distributed systems frameworks
- More real-world use case examples
- More benchmarking and performance analysis

Please check our [development guide](../../docs/development.md) and open an issue to discuss your ideas.

## Learn More

- [Architecture](architecture.md) - Detailed design documentation
- [Optimization Guide](optimization.md) - Performance benchmarks and tuning
- [Examples](../../examples/emerge/) - Complete working examples
- [API Reference](https://pkg.go.dev/github.com/carlisia/bio-adapt/emerge) - Full API documentation
- [Main README](../../README.md) - Project overview
- [Primitives Overview](../primitives.md) - Compare all coordination primitives
