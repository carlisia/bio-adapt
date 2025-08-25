# Agents

## Overview

An agent is the fundamental unit of coordination in bio-adapt. Think of an agent as an independent entity that can synchronize its behavior with other agents, like fireflies synchronizing their flashing or heart cells synchronizing their beating.

## What is an Agent?

An agent is a software component that:

- Maintains its own internal state (phase, frequency, energy)
- Observes the state of neighboring agents
- Adjusts its behavior to achieve coordination goals
- Operates independently without central control

### Real-World Analogy

Imagine a group of musicians trying to play in sync without a conductor:

- Each musician (agent) has their own tempo (frequency)
- They listen to nearby musicians (observe neighbors)
- They adjust their timing (phase) to match others
- Eventually, they all play in harmony (synchronization)

## Agent Components

### Phase (0 to 2π)

The agent's position in its cycle, like the position of a clock hand:

- 0 = start of cycle (12 o'clock)
- π = middle of cycle (6 o'clock)
- 2π = end of cycle (back to 12 o'clock)

When agents have the same phase, they act at the same time.

### Frequency

How fast the agent's phase advances:

- High frequency = rapid cycling (e.g., every 50ms)
- Low frequency = slow cycling (e.g., every 5 seconds)

Agents adjust their frequency to match neighbors.

### Energy

The agent's capacity to make adjustments:

- High energy = can make large adjustments quickly
- Low energy = must conserve, make small adjustments
- Depleted = cannot adjust until energy recovers

Energy prevents infinite adjustments and ensures stability.

### Neighbors

The other agents this agent can observe and coordinate with:

- In small systems, every agent might see all others
- In large systems, agents only see nearby neighbors
- The connection pattern affects how synchronization spreads

## How Agents Work

### Step 1: Initialization

```go
// Each agent starts with:
agent := &Agent{
    Phase:     random(),        // Random starting position
    Frequency: 1.0 + noise(),   // Slightly different speed
    Energy:    100,             // Full energy
}
```

### Step 2: Observation

Each agent observes its neighbors' states:

```go
// Agent checks neighbors
for _, neighbor := range agent.Neighbors() {
    neighborPhase := neighbor.Phase()
    // Agent notes the difference
    phaseDiff := neighborPhase - agent.Phase
}
```

### Step 3: Adjustment

Based on observations, the agent adjusts its state:

```go
// Agent adjusts toward neighbors
if phaseDiff > 0 {
    agent.Phase += adjustment  // Speed up
} else {
    agent.Phase -= adjustment  // Slow down
}
agent.Energy -= adjustmentCost  // Uses energy
```

### Step 4: Emergence

Through repeated local adjustments, global patterns emerge:

- Random phases → Synchronized phases (all agents in sync)
- Random phases → Distributed phases (evenly spread out)
- Random phases → Clustered phases (groups in sync)

## Types of Agents in Bio-adapt

### Emerge Agents

Specialized for temporal synchronization:

- Focus on coordinating **when** actions happen
- Use phase and frequency to achieve timing goals
- Examples: API batching, scheduled tasks, load distribution

### Navigate Agents (Coming Soon)

Specialized for resource navigation:

- Focus on finding **what** resources to use
- Navigate configuration spaces
- Examples: Resource allocation, path finding

### Glue Agents (Planned)

Specialized for collective intelligence:

- Focus on understanding **how** things work
- Build shared knowledge through interaction
- Examples: Schema discovery, consensus building

## Agent Communication

Agents don't send messages directly. Instead, they:

1. **Observe state** - See neighbors' phases and frequencies
2. **Adjust locally** - Change their own state based on observations
3. **Influence indirectly** - Their changes affect how neighbors behave

This is like birds in a flock - no bird commands the others, but they all coordinate by observing and adjusting.

## Agent Lifecycle

### Creation

```go
// Agents are created as part of a swarm
swarm := emerge.NewSwarm(
    100,  // Create 100 agents
    goal.MinimizeAPICalls,  // Their coordination goal
)
```

### Running

```go
// Agents continuously update their state
for {
    agent.ObserveNeighbors()
    agent.ComputeAdjustment()
    agent.ApplyAdjustment()
    agent.RecoverEnergy()
    time.Sleep(updateInterval)
}
```

### Convergence

```go
// Agents reach their goal state
if swarm.IsConverged() {
    // All agents are now coordinated
}
```

## Agent Properties

### Autonomy

Each agent makes its own decisions:

- No central controller
- No global knowledge required
- Continues working even if some agents fail

### Adaptability

Agents adjust strategies when needed:

- If gentle adjustments don't work, try stronger ones
- If energy is low, conserve resources
- If conditions change, adapt approach

### Robustness

The system works even with failures:

- If an agent stops, others continue
- New agents can join anytime
- System self-heals from disruptions

## Using Agents in Your Application

### Don't Create Agents Directly

Instead, use the high-level client API:

```go
// The client manages agents for you
client := emerge.MinimizeAPICalls(scale.Medium)
client.Start(ctx)
```

### Agents vs Workloads

- **Agents** = Synchronization mechanism (provided by bio-adapt)
- **Workloads** = Your application logic (what you implement)

Your workload wraps around agents:

```go
type MyWorkload struct {
    emergeAgent *agent.Agent  // The bio-adapt agent
    myData      []Task        // Your application data
}
```

### Monitoring Agents

```go
// Check overall synchronization
coherence := client.Coherence()  // 0 = chaos, 1 = perfect sync

// Check if goal is achieved
if client.IsConverged() {
    // Agents have reached target state
}
```

## Common Questions

### How many agents do I need?

Depends on your use case:

- **Tiny** (20 agents) - Testing, demos
- **Small** (50 agents) - Team coordination
- **Medium** (200 agents) - Department systems
- **Large** (1000 agents) - Organization scale
- **Huge** (2000+ agents) - Enterprise systems

### How fast do agents synchronize?

Depends on scale and goal:

- Small scale + simple goal = 1-2 seconds
- Large scale + complex goal = 10-30 seconds

### Can agents have different behaviors?

Yes, through:

- Different natural frequencies (some faster, some slower)
- Different stubbornness levels (resistance to change)
- Different energy levels (capacity for adjustment)

### What if an agent fails?

The system continues working:

- Other agents detect the missing neighbor
- They adjust their coordination pattern
- System maintains goal despite failure

## Best Practices

### DO:

- Let agents handle synchronization automatically
- Focus on your application logic, not agent mechanics
- Monitor coherence to track progress
- Use appropriate scale for your needs

### DON'T:

- Try to control individual agents directly
- Assume instant synchronization
- Ignore energy constraints
- Create more agents than needed

## See Also

- [Phase](phase.md) - Understanding agent positioning
- [Frequency](frequency.md) - Understanding agent timing
- [Energy](energy.md) - Understanding agent resources
- [Workload Integration](workload-integration.md) - How to use agents in your app
- [Goals](goals.md) - What agents can achieve
- [Strategies](strategies.md) - How agents coordinate

