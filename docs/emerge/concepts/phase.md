# Phase

## Overview

Phase represents an agent's position in its oscillation cycle, ranging from 0 to 2π radians. It's the fundamental state variable that agents adjust to achieve synchronization.

## Understanding Phase

### What is Phase?

Think of phase like the position of a clock hand:

- **0 (or 2π)** = 12 o'clock position
- **π/2** = 3 o'clock position
- **π** = 6 o'clock position
- **3π/2** = 9 o'clock position

When agents have the same phase, they're "in sync" - acting at the same time.

### Phase Difference

The phase difference between agents determines their relative timing:

- **0 difference** = Perfect synchronization
- **π difference** = Anti-phase (opposite timing)
- **π/2 difference** = Quarter cycle offset

## Phase in Different Goals

### Synchronization Goals

For goals like `MinimizeAPICalls`, agents aim for the same phase:

```
Agent A: ●────────── (phase = 0)
Agent B: ●────────── (phase = 0)
Agent C: ●────────── (phase = 0)
Result: All agents act together (batching)
```

### Distribution Goals

For goals like `DistributeLoad`, agents aim for different phases:

```
Agent A: ●────────── (phase = 0)
Agent B: ────●────── (phase = π/2)
Agent C: ────────●── (phase = π)
Result: Agents act at different times (load spreading)
```

## API Usage

### Reading Phase

```go
// Direct agent access (advanced usage)
agents := client.Agents()
for id, agent := range agents {
    phase := agent.Phase()
    fmt.Printf("Agent %s phase: %.2f radians\n", id, phase)
}
```

### Phase Updates

Phases are updated automatically by the synchronization strategies. You don't set phases directly - instead, you set goals and the system adjusts phases to achieve them.

## Phase Dynamics

### How Phases Change

The Kuramoto model governs phase evolution:

```
dθ/dt = ω + (K/N) × Σ sin(θ_j - θ_i)
```

Where:

- θ = phase
- ω = natural frequency
- K = coupling strength
- N = number of neighbors

### Synchronization Process

1. **Random initialization** - Agents start with random phases
2. **Local coupling** - Agents sense neighbor phases
3. **Adjustment** - Agents adjust toward neighbors
4. **Emergence** - Global synchronization emerges

## Frequency and Phase

### Relationship

- **Frequency** = Rate of phase change
- **Phase** = Current position in cycle

```go
// Phase advances based on frequency
newPhase = oldPhase + (frequency * deltaTime)
```

### Common Frequencies

- **Fast** (50-100ms) - Real-time coordination
- **Medium** (200-500ms) - Standard batching
- **Slow** (1-5s) - Periodic tasks
- **Very slow** (>1min) - Scheduled jobs

## Phase Patterns

### Common Patterns

1. **Uniform phase** - All agents at same phase (perfect sync)
2. **Uniform distribution** - Evenly spread phases (load balancing)
3. **Clustered phases** - Groups at different phases (partial sync)
4. **Random phases** - No pattern (chaos)

### Measuring Phase Distribution

The coherence metric captures how well phases align:

- High coherence = Phases clustered together
- Low coherence = Phases spread out

## Visualization

### Phase on Unit Circle

Phases are often visualized on a unit circle:

```
        0/2π
         ↑
    3π/2 ← → π/2
         ↓
         π
```

Multiple agents appear as dots on the circle. Synchronized agents cluster at the same angle.

## Best Practices

1. **Don't manipulate phases directly** - Let the system handle it
2. **Think in terms of goals** - Set what you want, not how to get there
3. **Monitor coherence, not individual phases** - Overall sync matters more
4. **Allow time for phase alignment** - Convergence isn't instant

## Common Issues

### Phases Not Converging

- Check energy levels - depleted agents can't adjust
- Verify coupling strength - too weak = slow convergence
- Review stubbornness - too high = agents resist change

### Phase Oscillations

- Normal during early convergence
- Problematic if persistent - may indicate parameter issues

## See Also

- [Coherence](coherence.md) - Measuring phase synchronization
- [Frequency](frequency.md) - Rate of phase change
- [Energy](energy.md) - Resource constraints on phase adjustments
