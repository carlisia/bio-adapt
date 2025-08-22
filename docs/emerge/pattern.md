# Emerge package

**Goal-directed synchronization through adaptive strategies** - Distributed systems that pursue target coordination states through multiple pathways, finding alternatives when defaults fail.

This package implements goal-directed temporal coordination, inspired by how biological systems reliably achieve morphological goals despite perturbations. Agents maintain synchronization targets as invariants and adaptively switch strategies to reach them.

## Core Concepts

### 📊 Phase (0 to 2π)

Where in the cycle each agent is - like the position of a clock hand. When phases align, agents act simultaneously.

### 🎵 Frequency

How fast agents cycle - like a heartbeat. Synchronized frequencies create rhythm.

### 🎯 Coherence (0 to 1)

The goal metric - how synchronized the swarm is. 0 = chaos, 1 = perfect sync. The system pursues target coherence levels through adaptive strategies.

### ⚡ Energy

Resource that agents spend to adjust behavior. Creates realistic constraints on adaptation.

## Architecture

```markdown
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ Agent 1 │◀────▶│ Agent 2 │◀────▶│ Agent 3 │
│ φ=0.2π │ │ φ=0.3π │ │ φ=0.25π │
└─────────────┘ └─────────────┘ └─────────────┘
▲ ▲ ▲
└────────────────────┼────────────────────┘
Adaptive Strategies
│
▼
┌─────────────────────────┐
│ Goal State │
│ Target: φ=0 │
│ Coherence: 0.9 │
│ Multiple Paths → │
└─────────────────────────┘
```

## Real-World Use Cases

### 📦 API Request Batching

```go
// Goal: Minimize API calls through coordinated batching
swarm, _ := emerge.NewSwarm(50)
target := emerge.Pattern{
    Frequency: 200 * time.Millisecond, // Target: 5 batches/second
    Coherence: 0.9,                    // Goal: 90% synchronization
}
swarm.AchieveSynchronization(ctx, target)
// Result: System finds optimal strategy to reduce API calls by 80%
```

### 🔄 Distributed Cron Jobs

```go
// Goal: Prevent thundering herd through controlled desynchronization
swarm, _ := emerge.NewSwarm(100)
target := emerge.Pattern{
    Frequency: 1 * time.Hour,  // Target interval
    Coherence: 0.1,            // Goal: Spread out (anti-sync)
}
swarm.AchieveSynchronization(ctx, target)
// System maintains low coherence to distribute load
```

### 💾 Database Connection Pooling

```go
// Goal: Optimal connection distribution without overload
swarm, _ := emerge.NewSwarm(200)
target := emerge.Pattern{
    Frequency: 50 * time.Millisecond,  // Target connection rate
    Coherence: 0.7,                    // Goal: Moderate clustering
}
swarm.AchieveSynchronization(ctx, target)
// Adaptively maintains connection patterns despite load changes
```

## Advanced Features

### Goal-Directed Strategy Switching

```go
// System automatically switches between strategies to achieve goals
strategies := []Strategy{
    &PhaseNudge{},      // Gentle phase adjustments
    &FrequencyLock{},   // Frequency alignment
    &PulseCoupling{},   // Strong synchronization pulses
}

swarm.SetStrategies(strategies)
// System will adaptively switch strategies when convergence stalls
```

### Energy Management

```go
agent.SetEnergy(100)                // Starting energy
agent.SetEnergyRecoveryRate(5)      // Units per second
agent.SetMinEnergyThreshold(20)     // Won't act below this
```

### Disruption Handling

```go
swarm.DisruptAgents(0.3)            // Disrupt 30% of agents
// Goal-directed system finds alternative paths to target state
// Recovery isn't just stability - it's achieving the original goal
```

## Performance characteristics

- **Convergence time**: Sub-linear scaling with swarm size
- **Memory usage**: ~2-3KB per agent at scale
- **Fault tolerance**: Automatic recovery from disruptions
- **Network traffic**: O(log N) - minimal coordination overhead

See [optimization guide](../docs/emerge/optimization.md) for detailed benchmarks.

## Package structure

```
emerge/
├── agent/          # Core agent implementation with optimizations
├── swarm/          # Swarm coordination and management
├── core/           # Fundamental types and interfaces
├── strategy/       # Decision-making strategies
├── goal/           # Goal management and weighting
├── monitoring/     # Convergence monitoring
└── decision/       # Decision engines
```

## Implementation details

### Agent optimization

Agents automatically optimize based on swarm size:

- **Small swarms (≤100)**: Use sync.Map for simplicity
- **Large swarms (>100)**: Switch to fixed arrays for cache locality

### State management

Atomic operations are grouped to reduce cache line bouncing:

- `AtomicState`: Phase, energy, frequency (frequently accessed together)
- `AtomicBehavior`: Influence, stubbornness (changed less often)

## Theory & Research

This implementation combines goal-directedness from biology with mathematical models:

- **Goal-Directedness** - Inspired by Michael Levin's research on how biological systems maintain target states as invariants
- **Kuramoto Model** - Provides the synchronization dynamics
- **Adaptive Strategies** - Multiple pathways to achieve goals, switching when stuck
- **Attractor Basins** - Enable convergence through different routes to the same target

## API stability

The emerge package API is stable and production-ready. We follow semantic versioning:

- Core interfaces (`Agent`, `Swarm`) are stable
- Strategy interfaces allow custom implementations
- Internal optimizations are transparent to users

## Documentation

### Package-specific
- [Architecture](../docs/emerge/architecture.md) - Emerge design details
- [Optimization](../docs/emerge/optimization.md) - Performance benchmarks

### Project-wide
- [Main README](../) - Project overview
- [Patterns overview](../docs/patterns.md) - All available patterns
- [Examples](../examples/emerge/) - Working code samples
- [API Reference](https://pkg.go.dev/github.com/carlisia/bio-adapt/emerge) - Complete API documentation
