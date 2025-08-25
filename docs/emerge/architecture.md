# Emerge architecture

## Overview

The emerge package implements [goal-directed](goal-directed.md) synchronization - distributed systems that pursue target coordination states through adaptive [strategies](../concepts/strategies.md). Inspired by how biological systems achieve morphological goals through multiple pathways, emerge maintains synchronization targets as invariants and switches strategies when convergence stalls.

## Core concepts

### Goal-directed phase synchronization

[Agents](../concepts/agents.md) maintain a [phase](../concepts/phase.md) (0 to 2π) representing their position in an oscillation cycle. The system pursues target phase alignment through multiple strategies, adapting when the default approach fails to achieve the [synchronization](../concepts/synchronization.md) goal.

### Dynamics and adaptation

The Kuramoto model provides the synchronization dynamics (see [Algorithm](emerge_algorithm.md) for details):

```text
dθᵢ/dt = ωᵢ + (K/N) × Σⱼ sin(θⱼ - θᵢ)
```

Where:

- θᵢ = phase of agent i
- ωᵢ = natural frequency of agent i
- K = coupling strength (adaptively adjusted)
- N = number of neighbors

Goal-directedness adds adaptive strategy switching when this default dynamics fails to achieve targets (see [Goal-Directed](goal-directed.md)).

### Energy constraints

Agents have limited [energy](../concepts/energy.md) for adjustments, preventing oscillation and ensuring stable convergence. Energy depletes with actions and recovers over time.

## Package structure

```bash
emerge/
├── agent/          # Core agent implementation with optimizations
├── swarm/          # Swarm coordination and goal-directed convergence
│   └── goal_directed.go  # Adaptive strategy switching (key file)
├── core/           # Fundamental types and interfaces
├── strategy/       # Multiple pathways to goals
├── completion/     # Pattern completion for gap filling
├── convergence/    # Convergence monitoring
├── goal/           # Goal management and blending
├── monitoring/     # System monitoring and metrics
├── decision/       # Decision engines
├── scale/          # Scaling utilities
└── trait/          # Agent traits
```

## Agent implementation

### State components

Each agent maintains:

- **[Phase](../concepts/phase.md)** (0 to 2π) - Position in oscillation cycle
- **[Frequency](../concepts/frequency.md)** - Oscillation speed
- **[Energy](../concepts/energy.md)** - Available action resources
- **LocalGoal** - Individual preferences (see [Goals](../concepts/goals.md))

### Behavioral parameters

- **Influence** - How much agent affects neighbors (0.0 to 1.0)
- **Stubbornness** - Resistance to external influence (0.0 to 1.0)
- **CouplingStrength** - Connection strength to neighbors

### Optimization layers

Agents automatically optimize based on swarm size:

**Small swarms (≤100 agents)**

- sync.Map for neighbor storage
- Standard atomic fields
- Simple iteration patterns

**Large swarms (>100 agents)**

- Fixed-size arrays for neighbors
- Grouped atomic fields to reduce cache bouncing
- Pre-allocated storage pools

## Swarm coordination

### Goal-directed convergence process

1. **Goal setting** - Define target synchronization state
2. **Strategy selection** - Choose initial approach
3. **Local sensing** - Agents observe neighbor states
4. **Adaptive adjustment** - Apply current strategy
5. **Progress monitoring** - Check convergence toward goal
6. **Strategy switching** - Change approach if stuck
7. **Goal achievement** - Continue until target reached

### Coherence measurement

[Coherence](../concepts/coherence.md) measures synchronization using the Kuramoto order parameter:

```go
R = |Σ(e^(iθ))| / N
```

- 0.0 = No synchronization (chaos)
- 0.5 = Partial synchronization
- 1.0 = Perfect synchronization

### Goal management

[Swarms](../concepts/swarm.md) maintain target states as invariants, finding alternative paths when blocked (see [Disruption](disruption.md)):

- **Phase** - Target alignment point (maintained despite disruptions)
- **Frequency** - Goal oscillation rate (achieved through multiple strategies)
- **Coherence** - Target synchronization level (pursued adaptively)

## Decision strategies

### Multiple pathways to goals

See [Strategies](../concepts/strategies.md) for detailed descriptions.

**PhaseNudge** (Gentle approach)

- Small incremental phase adjustments
- Minimal energy consumption
- First strategy tried for efficiency

**FrequencyLock** (Frequency-first approach)

- Aligns frequencies before phases
- Alternative path when phase adjustment alone fails
- Effective for disparate natural frequencies

**PulseCoupling** (Strong synchronization)

- Powerful synchronization pulses
- Used when gentle approaches stall
- Higher energy cost but faster convergence

**EnergyAware** (Resource-conscious)

- Balances goal achievement with resource limits
- Adapts strategy based on available energy
- Ensures sustainable convergence

### Custom strategies

Implement the `DecisionMaker` interface:

```go
type DecisionMaker interface {
    Decide(context DecisionContext) Decision
}

type DecisionContext struct {
    Current   State
    Target    State
    Neighbors []NeighborState
    Energy    float64
}
```

## Performance characteristics

### Scalability

| Swarm size | Convergence time | Memory/agent | CPU usage |
| ---------- | ---------------- | ------------ | --------- |
| 10-100     | ~800ms           | ~5KB         | Minimal   |
| 100-1000   | ~300ms/agent     | ~3KB         | Moderate  |
| 1000-5000  | Sub-linear       | ~2KB         | Optimized |

### Optimization triggers

- Swarm size > 100 → Array storage
- Update rate > 1000/sec → Atomic grouping
- Neighbors > 20 → Fixed neighbor arrays

## Use cases

### API request batching

Coordinate microservices to batch API calls:

```go
import (
    "github.com/carlisia/bio-adapt/client/emerge"
    "github.com/carlisia/bio-adapt/emerge/scale"
)

// Simple: Use client API
client := emerge.MinimizeAPICalls(scale.Small)  // 50 agents
client.Start(ctx)  // Pursues goal adaptively

// Advanced: Direct swarm access if needed
import "github.com/carlisia/bio-adapt/emerge/swarm"
cfg := swarm.For(goal.MinimizeAPICalls)
swarm, _ := swarm.New(50, targetState, swarm.WithGoalConfig(cfg))
swarm.Run(ctx)
```

### Distributed cron

Prevent thundering herd in scheduled tasks (see [Use Cases](use_cases.md)):

```go
// Use load distribution for anti-synchronization
client := emerge.DistributeLoad(scale.Small)  // Automatically targets low coherence
client.Start(ctx)
```

### Load balancing

Natural load distribution:

```go
// Custom configuration for moderate clustering
client := emerge.Custom().
    WithGoal(goal.DistributeLoad).
    WithScale(scale.Medium).  // 200 agents
    WithTargetCoherence(0.5).  // Moderate clustering
    Build()

client.Start(ctx)
```

## Fault tolerance

### Resilience mechanisms

See [Disruption](disruption.md) for detailed coverage.

**Agent failures**

- Neighbors detect missing agents
- Automatic topology reconfiguration
- Graceful coherence degradation

**Network partitions**

- Local coherence within partitions
- Automatic re-merge when healed
- No split-brain issues

**Byzantine agents**

- Energy limits prevent unlimited disruption
- Stubbornness limits influence spread
- Statistical convergence despite bad actors

See [Security](security.md) for security considerations.

## See Also

### Core Documentation
- [Algorithm](emerge_algorithm.md) - Mathematical foundation
- [Protocol](protocol.md) - Synchronization protocol
- [Goal-Directed](goal-directed.md) - Goal pursuit mechanisms
- [Disruption](disruption.md) - Failure handling
- [Decentralization](decentralization.md) - No central control
- [Concurrency](concurrency.md) - Go implementation patterns
- [Optimization](optimization.md) - Performance enhancements
- [Security](security.md) - Security considerations

### Concepts
- [Agents](../concepts/agents.md) - Fundamental units
- [Swarm](../concepts/swarm.md) - Agent collections
- [Synchronization](../concepts/synchronization.md) - Coordination
- [Coherence](../concepts/coherence.md) - Measurement
- [Phase](../concepts/phase.md) - Oscillation position
- [Frequency](../concepts/frequency.md) - Rate of change
- [Energy](../concepts/energy.md) - Resource constraints
- [Goals](../concepts/goals.md) - Objectives
- [Strategies](../concepts/strategies.md) - Approaches

### Practical Guides
- [Use Cases](use_cases.md) - Real-world applications
- [Scales](scales.md) - Configuration sizes
- [FAQ](faq.md) - Common questions
- [Glossary](glossary.md) - Term definitions
