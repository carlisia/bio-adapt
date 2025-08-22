# Architecture

## Overview

Bio-adapt implements goal-directed distributed coordination inspired by Michael Levin's research on how biological systems achieve reliable outcomes through multiple pathways. The architecture enables systems to maintain target states as invariants, finding alternative routes when defaults fail.

The library provides three complementary patterns for goal-directed behavior:
- **Emerge**: When should agents act? (temporal coordination)
- **Navigate**: What resources to use? (resource allocation)
- **Glue**: How does the API work? (collective understanding)

See [patterns overview](patterns.md) for details.

## Core design principles

### 1. Goal-directedness as the core principle

Inspired by Levin's research, systems maintain goals as invariants rather than following fixed procedures. When one path to a goal is blocked, the system discovers alternatives—just as regenerating tissue finds new routes to target morphologies.

### 2. Local interactions, global emergence

Agents only interact with nearby neighbors (local coupling), yet global synchronization emerges naturally - similar to how fireflies synchronize their flashing or how cardiac cells coordinate heartbeats.

### 3. Multiple pathways to goals

The system uses attractor basins to provide multiple convergence paths. When the default path fails, agents explore alternative strategies to reach the same target state—a key insight from biological goal-directedness.

### 4. Energy-constrained adaptation

Agents have limited energy for actions, creating realistic constraints that prevent oscillation and ensure stable convergence.

## System architecture

```
┌─────────────────────────────────────────────┐
│                Application Layer             │
│         (Your workloads/services)            │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│              Bio-Adapt Library               │
│                                              │
│  ┌─────────────────────────────────────┐    │
│  │       Goal-Directed Orchestration    │    │
│  │   - Target state maintenance         │    │
│  │   - Strategy adaptation              │    │
│  │   - Multi-pathway exploration        │    │
│  └─────────────┬───────────────────────┘    │
│                │                             │
│  ┌─────────────▼───────────────────────┐    │
│  │         Agent Coordination           │    │
│  │   - Neighbor discovery               │    │
│  │   - State synchronization            │    │
│  │   - Decision making                  │    │
│  └─────────────┬───────────────────────┘    │
│                │                             │
│  ┌─────────────▼───────────────────────┐    │
│  │          Core Primitives             │    │
│  │   - Phase dynamics                   │    │
│  │   - Energy management                │    │
│  │   - Attractor calculations           │    │
│  └─────────────────────────────────────┘    │
└──────────────────────────────────────────────┘
```

## Package structure

### Pattern packages

**emerge/** - Goal-directed synchronization (production-ready)

- Pursues target coordination states through adaptive strategies
- Kuramoto model provides the dynamics, goal-directedness provides the adaptation
- See [emerge documentation](emerge/pattern.md)

**navigate/** - Goal-directed resource allocation (coming soon)

- Navigates resource configuration spaces to reach allocation goals
- Discovers alternative paths when constraints block direct routes
- See [navigate documentation](navigate/pattern.md)

**glue/** - Goal-directed collective intelligence (planned)

- Collectively discovers solutions through distributed hypothesis testing
- Achieves understanding that no individual agent could reach alone
- See [glue documentation](glue/pattern.md)

### Shared packages

**internal/** - Internal utilities used across patterns

- `topology/` - Network topology builders (ring, star, full mesh)
- `resource/` - Resource management primitives
- `analysis/` - Quality and diagnostic tools
- `config/` - Configuration and validation

## Common agent architecture

All patterns share goal-directed agent concepts:

### Universal components

- **Current State** - Agent's present configuration
- **Target Goals** - Desired states maintained as invariants
- **Strategy Set** - Multiple pathways to achieve goals
- **Local Network** - Neighbors for collective problem-solving

### Pattern-specific implementations

**Emerge agents**

- Phase (0 to 2π) as state, synchronization as goal
- Multiple strategies: phase nudging, frequency locking, pulse coupling
- Adaptive switching when convergence stalls
- See [emerge architecture](emerge/architecture.md)

**Navigate agents** (future)

- Resource allocation as state, optimal distribution as goal
- Pathfinding through configuration space
- Memory of successful allocation paths

**Glue agents** (future)

- Partial knowledge as state, complete understanding as goal
- Hypothesis generation and testing
- Consensus building through local interactions

## Network topologies

### Supported topologies

**Full mesh** - Every agent connects to all others

- Fastest convergence
- Highest communication overhead
- Best for small swarms (<50 agents)

**Ring** - Each agent connects to k nearest neighbors

- Good balance of convergence and efficiency
- Scales well to large swarms
- Natural for geographic distribution

**Star** - Hub agents connect clusters

- Hierarchical coordination
- Good for multi-region deployments
- Natural bottleneck at hubs

**Small world** - Mostly local with random long connections

- Fast global convergence
- Resilient to disruptions
- Best for large swarms (>100 agents)

## Fault tolerance

### Resilience mechanisms

**Agent failures**

- Neighbors detect missing agents
- Automatic topology reconfiguration
- Graceful degradation of coherence

**Network partitions**

- Local coherence within partitions
- Automatic re-merge when healed
- No split-brain issues

**Byzantine agents**

- Energy limits prevent unlimited disruption
- Stubbornness parameters limit influence
- Statistical convergence despite bad actors

## Performance characteristics

### Scaling behavior

| Aspect                | Complexity | Notes                     |
| --------------------- | ---------- | ------------------------- |
| Agent creation        | O(1)       | Constant per agent        |
| Neighbor discovery    | O(k)       | k = neighbors per agent   |
| State update          | O(k)       | Parallel updates possible |
| Coherence measurement | O(n)       | Can be sampled            |
| Memory per agent      | O(k)       | ~2-3KB at scale           |

### Optimization triggers

The system automatically optimizes when:

- Swarm size > 100 agents → Array storage
- Update frequency > 1000/sec → Atomic grouping
- Neighbors > 20 → Fixed neighbor arrays

## Integration patterns

Each pattern excels at different integration scenarios:

### Emerge - Goal: Achieve synchronization

```go
// Goal: Minimize API calls through coordinated batching
swarm := emerge.NewSwarm(100)
target := emerge.Pattern{
    Frequency: 200*time.Millisecond,
    Coherence: 0.9,  // Goal: 90% synchronization
}
swarm.AchieveSynchronization(ctx, target)
```

### Navigate - Goal: Optimal resource allocation (future)

```go
// Goal: Navigate to target resource distribution
navigator := navigate.NewNavigator(100)
target := navigate.ResourceState{
    CPU: 0.7,
    Memory: 0.5,
}
navigator.NavigateToTarget(ctx, target)
```

### Glue - Goal: Discover API schema (future)

```go
// Goal: Collectively understand API contract
network := glue.NewNetwork(30)
observations := collectAPIResponses()
schema := network.SolveSchema(ctx, observations)
```

## Future directions

### Near-term (Navigate package)

- Implement resource space navigation
- Add alternative path discovery
- Create constraint-aware allocation examples

### Medium-term (Glue package)

- Implement distributed hypothesis testing
- Add collective knowledge building
- Create schema discovery examples

### Long-term enhancements

- Cross-pattern integration APIs
- Hardware acceleration (SIMD, GPU)
- Distributed multi-node deployments
- Real-time visualization tools

