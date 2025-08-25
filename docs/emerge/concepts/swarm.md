# Swarm

## Overview

A swarm is a collection of agents working together toward a common goal. Like a flock of birds or a school of fish, a swarm achieves complex collective behavior through simple local interactions between agents, without any central control.

## What is a Swarm?

Think of a swarm as a team where:

- Every member (agent) follows simple rules
- Members only interact with nearby teammates
- No single leader controls everyone
- The team achieves goals that no individual could accomplish alone

### Real-World Examples

**Fireflies synchronizing:**

- Each firefly flashes on its own schedule
- They adjust their timing based on nearby flashes
- Eventually, thousands flash in perfect unison
- No "master firefly" coordinates them

**Traffic patterns:**

- Each car follows simple rules (stay in lane, maintain distance)
- Drivers react to nearby vehicles
- Traffic patterns emerge (waves, jams, flows)
- No central traffic controller needed

## How Swarms Work

### 1. Many Simple Agents

```
Agent 1: phase=0.2π, frequency=1.0Hz
Agent 2: phase=0.5π, frequency=1.1Hz
Agent 3: phase=1.8π, frequency=0.9Hz
... (tens to thousands more)
```

### 2. Local Interactions

Each agent only sees a few neighbors:

```
Agent 5 sees: [Agent 4, Agent 6, Agent 11]
Agent 5 adjusts based only on these three
```

### 3. Emergent Behavior

From local interactions, global patterns emerge:

```
Time 0s:  Chaos - all agents random
Time 5s:  Clusters forming
Time 10s: Full synchronization
```

## Swarm Properties

### Decentralized

- No central coordinator
- No single point of failure
- Each agent acts independently
- Decisions emerge from collective behavior

### Scalable

Swarms work at different scales:

- **Tiny** (20 agents) - Quick demos
- **Small** (50 agents) - Team coordination
- **Medium** (200 agents) - Department scale
- **Large** (1000 agents) - Organization scale
- **Huge** (2000+ agents) - Enterprise scale

### Robust

Swarms handle disruptions well:

- Agents can fail without breaking the swarm
- New agents can join anytime
- Swarm adapts to changing conditions
- Self-healing behavior

### Adaptive

Swarms find alternative paths to goals:

- If one approach fails, try another
- Multiple strategies available
- Responds to environmental changes
- Learns from collective experience

## Creating a Swarm

### Using the Client API (Recommended)

```go
// The client creates and manages a swarm for you
client := emerge.MinimizeAPICalls(scale.Medium)
// This creates a 200-agent swarm optimized for API batching

err := client.Start(ctx)
// The swarm begins coordinating
```

### Direct Swarm Creation (Advanced)

```go
import "github.com/carlisia/bio-adapt/emerge/swarm"

// Create a swarm with 100 agents
s, err := swarm.New(
    100,                      // Number of agents
    goal.MinimizeAPICalls,    // What to achieve
)

// Start the swarm
s.Run(ctx)
```

## Swarm Dynamics

### Initial State: Chaos

When a swarm starts, agents are uncoordinated:

```
Agent phases: [0.3π, 1.7π, 0.8π, 2.1π, 1.2π...]
Coherence: 0.05 (nearly random)
```

### Convergence Process

Agents begin adjusting to neighbors:

```
Step 1: Local clusters form
Step 2: Clusters merge
Step 3: Global pattern emerges
Step 4: Goal state achieved
```

### Final State: Coordination

The swarm reaches its goal:

```
Agent phases: [0.1π, 0.1π, 0.12π, 0.11π, 0.1π...]
Coherence: 0.95 (highly synchronized)
```

## Swarm Topologies

### How Agents Connect

Different connection patterns affect swarm behavior:

#### Full Mesh

Every agent sees all others:

```
A ←→ B
↑ ╳ ↓
C ←→ D
```

- Fast convergence
- High communication overhead
- Best for small swarms

#### Ring

Agents only see immediate neighbors:

```
A → B
↑   ↓
D ← C
```

- Slower convergence
- Low overhead
- Scales well

#### Small World

Most connections local, some long-range:

```
A ←→ B ←→ C
↑         ↓
D ←→ E ←→ F
    ↓
    G
```

- Balance of speed and efficiency
- Robust to failures
- Good for medium/large swarms

## Monitoring a Swarm

### Coherence

Measures how synchronized the swarm is:

```go
coherence := client.Coherence()
// 0.0 = completely random
// 0.5 = partially synchronized
// 1.0 = perfectly synchronized
```

### Convergence

Checks if the swarm reached its goal:

```go
if client.IsConverged() {
    // Swarm achieved target state
    fmt.Println("Ready to batch operations!")
}
```

### Progress Tracking

```go
ticker := time.NewTicker(1 * time.Second)
for range ticker.C {
    progress := client.Coherence() * 100
    fmt.Printf("Synchronization: %.1f%%\n", progress)
}
```

## Swarm Goals

Different goals create different swarm behaviors:

### Synchronization Goals

Agents try to act together:

```go
// For batching operations
client := emerge.MinimizeAPICalls(scale.Medium)
// Target: High coherence (0.85-0.95)
```

### Distribution Goals

Agents try to spread out:

```go
// For load balancing
client := emerge.DistributeLoad(scale.Large)
// Target: Low coherence (0.1-0.3)
```

### Consensus Goals

Agents form agreement groups:

```go
// For distributed decisions
client := emerge.ReachConsensus(scale.Small)
// Target: Medium coherence (0.5-0.7)
```

## Swarm Performance

### Convergence Times

| Swarm Size   | Simple Goal | Complex Goal |
| ------------ | ----------- | ------------ |
| Tiny (20)    | 1-2 sec     | 2-5 sec      |
| Small (50)   | 2-5 sec     | 5-10 sec     |
| Medium (200) | 5-10 sec    | 10-20 sec    |
| Large (1000) | 10-20 sec   | 20-40 sec    |
| Huge (2000+) | 20-40 sec   | 40-80 sec    |

### Resource Usage

| Swarm Size | Memory | CPU (updates/sec) |
| ---------- | ------ | ----------------- |
| Tiny       | ~100KB | ~200              |
| Small      | ~250KB | ~500              |
| Medium     | ~1MB   | ~2,000            |
| Large      | ~5MB   | ~10,000           |
| Huge       | ~10MB  | ~20,000           |

## Common Patterns

### Pattern 1: Batch Processing

```go
// Create swarm for batching
client := emerge.MinimizeAPICalls(scale.Medium)
client.Start(ctx)

// Wait for synchronization
for !client.IsConverged() {
    time.Sleep(100 * time.Millisecond)
}

// Now safe to batch
processBatchedOperations()
```

### Pattern 2: Load Distribution

```go
// Create swarm for distribution
client := emerge.DistributeLoad(scale.Large)
client.Start(ctx)

// Check for good distribution
if client.Coherence() < 0.3 {
    // Well distributed - process work
    processNextTask()
}
```

### Pattern 3: Continuous Coordination

```go
// Swarm runs continuously
go client.Start(ctx)

// Application checks swarm state
for {
    if client.IsConverged() {
        doSynchronizedWork()
    }
    time.Sleep(checkInterval)
}
```

## Troubleshooting Swarms

### Problem: Slow Convergence

**Causes:**

- Swarm too large for goal
- Weak coupling between agents
- High stubbornness values

**Solutions:**

- Use smaller scale
- Increase coupling strength
- Reduce agent stubbornness

### Problem: Oscillating Coherence

**Causes:**

- Energy depletion cycles
- Competing influences
- Unstable parameters

**Solutions:**

- Increase energy recovery
- Adjust topology
- Use different strategy

### Problem: Stuck at Partial Coherence

**Causes:**

- Local minima
- Insufficient coupling
- Wrong goal for scenario

**Solutions:**

- Strategy will auto-switch
- Increase coupling
- Choose different goal

## Best Practices

### DO:

- Start with smaller swarms for testing
- Let swarms self-organize
- Monitor coherence trends
- Allow adequate convergence time
- Use appropriate scale for your use case

### DON'T:

- Try to control individual agents
- Expect instant convergence
- Create swarms larger than needed
- Ignore resource constraints
- Manually coordinate agents

## Advanced Concepts

### Swarm Intelligence

The swarm knows more than any individual:

- Collective problem solving
- Distributed decision making
- Emergent optimization
- Adaptive learning

### Multi-Goal Swarms

Swarms can pursue multiple objectives:

- Primary goal: synchronization
- Secondary goal: energy efficiency
- Balancing competing objectives

### Hierarchical Swarms

Swarms of swarms:

- Local swarms coordinate internally
- Regional swarms coordinate local swarms
- Global patterns from hierarchical structure

## See Also

- [Agents](agents.md) - The individual units in a swarm
- [Synchronization](synchronization.md) - How swarms coordinate
- [Goals](goals.md) - What swarms can achieve
- [Coherence](coherence.md) - Measuring swarm coordination
- [Scale Definitions](../emerge/scales.md) - Swarm size guidelines

