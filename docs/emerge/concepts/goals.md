# Goals

## Overview

Goals define what a swarm is trying to achieve. Each goal represents a different optimization objective and determines how agents should coordinate.

## Available Goals

### MinimizeAPICalls

**Objective:** Reduce API calls through synchronized batching  
**Target Coherence:** 0.85-0.95 (high synchronization)  
**Use Case:** Batch database writes, API requests, file I/O

```go
client := emerge.MinimizeAPICalls(scale.Medium)
// Agents synchronize to act together, batching operations
```

### DistributeLoad

**Objective:** Spread work evenly across time/resources  
**Target Coherence:** 0.1-0.3 (anti-synchronization)  
**Use Case:** Load balancing, prevent thundering herd

```go
client := emerge.DistributeLoad(scale.Large)
// Agents deliberately desynchronize to spread load
```

### ReachConsensus

**Objective:** Achieve distributed agreement  
**Target Coherence:** 0.5-0.7 (partial synchronization)  
**Use Case:** Voting systems, quorum decisions

```go
client := emerge.ReachConsensus(scale.Small)
// Agents form consensus groups
```

### MinimizeLatency

**Objective:** Reduce response time through coordination  
**Target Coherence:** 0.8-0.9 (high synchronization)  
**Use Case:** Real-time systems, gaming servers

```go
client := emerge.MinimizeLatency(scale.Tiny)
// Quick, tight coordination for fast responses
```

### SaveEnergy

**Objective:** Minimize resource consumption  
**Target Coherence:** 0.4-0.6 (moderate synchronization)  
**Use Case:** IoT devices, battery-powered systems

```go
client := emerge.SaveEnergy(scale.Medium)
// Sparse synchronization to conserve resources
```

### MaintainRhythm

**Objective:** Keep consistent timing  
**Target Coherence:** 0.85-0.95 (high synchronization)  
**Use Case:** Scheduled tasks, periodic operations

```go
client := emerge.MaintainRhythm(scale.Small)
// Maintain steady, predictable timing
```

### RecoverFromFailure

**Objective:** Maintain service despite disruptions  
**Target Coherence:** 0.2-0.4 (low synchronization)  
**Use Case:** Fault-tolerant systems, self-healing

```go
client := emerge.RecoverFromFailure(scale.Large)
// Distributed resilience through independence
```

### AdaptToTraffic

**Objective:** Respond to load changes  
**Target Coherence:** 0.6-0.8 (adaptive)  
**Use Case:** Auto-scaling, dynamic systems

```go
client := emerge.AdaptToTraffic(scale.Medium)
// Adjust coordination based on load
```

## Goal Configuration

### Using Presets

Each goal comes with optimized presets:

```go
// Simple one-liner with all settings configured
client := emerge.MinimizeAPICalls(scale.Large)
```

### Custom Goal Configuration

```go
import "github.com/carlisia/bio-adapt/emerge/goal"

// Use builder for custom configuration
client := emerge.Custom().
    WithGoal(goal.MinimizeAPICalls).
    WithScale(scale.Large).
    WithTargetCoherence(0.95). // Override default
    Build()
```

### Direct Swarm Configuration (Advanced)

```go
import "github.com/carlisia/bio-adapt/emerge/swarm"

// Configure swarm directly
cfg := swarm.For(goal.MinimizeAPICalls).
    With(scale.Large)

swarm, _ := swarm.New(1000, targetState, swarm.WithGoalConfig(cfg))
```

## Goal Parameters

Each goal optimizes different parameters:

| Goal               | Coupling Strength | Update Rate | Strategy       | Network     |
| ------------------ | ----------------- | ----------- | -------------- | ----------- |
| MinimizeAPICalls   | High              | Fast        | PulseCoupling  | Full mesh   |
| DistributeLoad     | Negative          | Medium      | PhaseRepulsion | Ring        |
| ReachConsensus     | Medium            | Medium      | VotingClusters | Small world |
| MinimizeLatency    | Very High         | Very Fast   | FrequencyLock  | Full mesh   |
| SaveEnergy         | Low               | Slow        | EnergyAware    | Sparse      |
| MaintainRhythm     | High              | Steady      | PhaseNudge     | Ring        |
| RecoverFromFailure | Low               | Adaptive    | Resilient      | Redundant   |
| AdaptToTraffic     | Variable          | Dynamic     | Adaptive       | Dynamic     |

## Choosing the Right Goal

### Decision Matrix

**Need to batch operations?** → MinimizeAPICalls  
**Need to spread load?** → DistributeLoad  
**Need agreement?** → ReachConsensus  
**Need speed?** → MinimizeLatency  
**Need efficiency?** → SaveEnergy  
**Need consistency?** → MaintainRhythm  
**Need resilience?** → RecoverFromFailure  
**Need adaptability?** → AdaptToTraffic

### Goal Combinations

Some goals work well together:

- MinimizeAPICalls + SaveEnergy = Efficient batching
- DistributeLoad + RecoverFromFailure = Resilient distribution
- ReachConsensus + MaintainRhythm = Consistent decisions

## Performance by Goal

### Fast Convergence Goals

- MinimizeLatency (< 1 second)
- MinimizeAPICalls (2-5 seconds)
- MaintainRhythm (2-5 seconds)

### Medium Convergence Goals

- ReachConsensus (5-10 seconds)
- SaveEnergy (5-15 seconds)
- AdaptToTraffic (5-20 seconds)

### Slow Convergence Goals

- DistributeLoad (10-30 seconds)
- RecoverFromFailure (continuous)

## Monitoring Goal Achievement

```go
// Check if goal is achieved
if client.IsConverged() {
    // Target state reached
}

// Monitor progress
coherence := client.Coherence()
fmt.Printf("Progress: %.1f%%\n", coherence * 100)

// Goal-specific checks
switch currentGoal {
case goal.MinimizeAPICalls:
    if coherence > 0.85 {
        // Ready to batch
    }
case goal.DistributeLoad:
    if coherence < 0.3 {
        // Good distribution
    }
}
```

## Best Practices

1. **Start with presets** - They're optimized for each goal
2. **Match scale to goal** - Some goals work better at certain scales
3. **Allow convergence time** - Don't rush goal achievement
4. **Monitor appropriate metrics** - Coherence interpretation varies by goal
5. **Test goal switches** - Ensure smooth transitions between goals

## See Also

- [Coherence](coherence.md) - Understanding synchronization levels
- [Scale Definitions](../emerge/scales.md) - Optimal scales for each goal
- [Strategies](strategies.md) - How goals are achieved
