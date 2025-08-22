# Emerge architecture

## Overview

The emerge package implements goal-directed synchronization - distributed systems that pursue target coordination states through adaptive strategies. Inspired by how biological systems achieve morphological goals through multiple pathways, emerge maintains synchronization targets as invariants and switches strategies when convergence stalls.

## Core concepts

### Goal-directed phase synchronization

Agents maintain a phase (0 to 2π) representing their position in an oscillation cycle. The system pursues target phase alignment through multiple strategies, adapting when the default approach fails to achieve the synchronization goal.

### Dynamics and adaptation

The Kuramoto model provides the synchronization dynamics:

```text
dθᵢ/dt = ωᵢ + (K/N) × Σⱼ sin(θⱼ - θᵢ)
```

Where:

- θᵢ = phase of agent i
- ωᵢ = natural frequency of agent i  
- K = coupling strength (adaptively adjusted)
- N = number of neighbors

Goal-directedness adds adaptive strategy switching when this default dynamics fails to achieve targets.

### Energy constraints

Agents have limited energy for adjustments, preventing oscillation and ensuring stable convergence. Energy depletes with actions and recovers over time.

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

- **Phase** (0 to 2π) - Position in oscillation cycle
- **Frequency** - Oscillation speed
- **Energy** - Available action resources
- **LocalGoal** - Individual preferences

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

Coherence measures synchronization using the Kuramoto order parameter:

```go
R = |Σ(e^(iθ))| / N
```

- 0.0 = No synchronization (chaos)
- 0.5 = Partial synchronization
- 1.0 = Perfect synchronization

### Goal management

Swarms maintain target states as invariants, finding alternative paths when blocked:

- **Phase** - Target alignment point (maintained despite disruptions)
- **Frequency** - Goal oscillation rate (achieved through multiple strategies)
- **Coherence** - Target synchronization level (pursued adaptively)

## Decision strategies

### Multiple pathways to goals

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
swarm, _ := emerge.New(50, emerge.State{
    Phase:     0,
    Frequency: 200*time.Millisecond, // Goal: 5 batches/sec
    Coherence: 0.9,                  // Goal: 90% synchronization
})
swarm.Run(ctx) // Pursues goal adaptively
```

### Distributed cron

Prevent thundering herd in scheduled tasks:

```go
swarm, _ := emerge.New(100, emerge.State{
    Phase:     0,
    Frequency: 1*time.Hour,
    Coherence: 0.1, // Spread out (anti-sync)
})
```

### Load balancing

Natural load distribution:

```go
swarm, _ := emerge.New(200, emerge.State{
    Phase:     0,
    Frequency: 250*time.Millisecond,
    Coherence: 0.5, // Moderate clustering
})
```

## Fault tolerance

### Resilience mechanisms

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

