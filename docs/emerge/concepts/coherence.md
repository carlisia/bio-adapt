# Coherence

## Overview

Coherence measures the degree of synchronization in a swarm, ranging from 0 (no synchronization) to 1 (perfect synchronization). It's the primary metric for determining whether a swarm has achieved its goal.

## Understanding Coherence Values

| Coherence | Synchronization Level | Description                             |
| --------- | --------------------- | --------------------------------------- |
| 0.0 - 0.2 | None/Chaos            | Agents are completely unsynchronized    |
| 0.2 - 0.4 | Poor                  | Minimal coordination, mostly random     |
| 0.4 - 0.6 | Partial               | Some groups forming, inconsistent       |
| 0.6 - 0.8 | Good                  | Most agents coordinated, some outliers  |
| 0.8 - 0.9 | Very Good             | Strong synchronization, few outliers    |
| 0.9 - 1.0 | Excellent             | Near-perfect to perfect synchronization |

## Mathematical Foundation

Coherence is calculated using the Kuramoto order parameter:

```
R = |Σ(e^(iθ))| / N
```

Where:

- R = coherence (0 to 1)
- θ = phase of each agent (0 to 2π)
- N = number of agents
- i = imaginary unit

## Goal-Specific Coherence Targets

Different goals require different coherence levels:

### High Coherence Goals (0.8 - 0.95)

- **MinimizeAPICalls** - Tight synchronization for batching
- **MaintainRhythm** - Consistent timing
- **MinimizeLatency** - Quick response coordination

### Medium Coherence Goals (0.5 - 0.7)

- **ReachConsensus** - Partial agreement sufficient
- **SaveEnergy** - Moderate coordination

### Low Coherence Goals (0.1 - 0.3)

- **DistributeLoad** - Anti-synchronization for spreading
- **RecoverFromFailure** - Distributed resilience

## API Usage

### Checking Coherence

```go
// With emerge client
client := emerge.MinimizeAPICalls(scale.Medium)
client.Start(ctx)

// Check current coherence
coherence := client.Coherence()
if coherence > 0.8 {
    // Good synchronization achieved
}

// Check if target coherence reached
if client.IsConverged() {
    // Target coherence achieved
}
```

### Setting Target Coherence

```go
// Use scale defaults
client := emerge.MinimizeAPICalls(scale.Large) // Uses 0.80 default

// Override target coherence
client := emerge.Custom().
    WithGoal(goal.MinimizeAPICalls).
    WithScale(scale.Large).
    WithTargetCoherence(0.95). // Override to 0.95
    Build()
```

## Convergence Behavior

### Typical Convergence Pattern

1. **Initial chaos** (0.0 - 0.2) - Random starting phases
2. **Rapid improvement** (0.2 - 0.6) - Agents find neighbors
3. **Steady progress** (0.6 - 0.8) - Groups merge
4. **Fine tuning** (0.8+) - Final synchronization

### Factors Affecting Convergence

- **Swarm size** - Larger swarms take longer
- **Coupling strength** - Stronger coupling = faster convergence
- **Network topology** - Full mesh fastest, ring slowest
- **Agent stubbornness** - Higher stubbornness = slower convergence

## Monitoring and Debugging

### Signs of Problems

- **Stuck at low coherence** (<0.3 after 30s) - Check energy levels
- **Oscillating coherence** - May indicate parameter conflicts
- **Very slow convergence** - Consider reducing stubbornness
- **Never reaching target** - Target may be too high for scale

### Best Practices

1. Start with scale-appropriate defaults
2. Monitor coherence over time, not just final value
3. Allow sufficient time for convergence (larger scales need more)
4. Use `IsConverged()` rather than checking exact values

## See Also

- [Phase](phase.md) - Understanding agent phases
- [Scale Definitions](../emerge/scales.md) - Default coherence per scale
- [Goals](goals.md) - Goal-specific coherence requirements
