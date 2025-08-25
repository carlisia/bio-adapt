# Frequency

## Overview

Frequency determines how fast an agent's phase advances through its oscillation cycle. It's measured in cycles per unit time and directly affects how often agents perform actions.

## Understanding Frequency

### What is Frequency?

- **Frequency** = Rate of phase change
- **Period** = 1/frequency = Time for one complete cycle
- **Higher frequency** = Faster oscillation, more frequent actions
- **Lower frequency** = Slower oscillation, less frequent actions

### Common Frequencies

| Frequency       | Period         | Use Case               |
| --------------- | -------------- | ---------------------- |
| 20 Hz (50ms)    | Very fast      | Real-time coordination |
| 5 Hz (200ms)    | Fast           | API batching           |
| 2 Hz (500ms)    | Medium         | Standard operations    |
| 1 Hz (1s)       | Slow           | Periodic tasks         |
| 0.1 Hz (10s)    | Very slow      | Scheduled jobs         |
| 0.017 Hz (1min) | Extremely slow | Cron-like tasks        |

## Frequency in Synchronization

### Natural Frequency

Each agent has a natural frequency (ω) - its preferred oscillation rate:

```go
// Agents start with slightly different natural frequencies
agent1.naturalFrequency = 1.0 Hz
agent2.naturalFrequency = 1.1 Hz
agent3.naturalFrequency = 0.9 Hz
```

### Frequency Synchronization

Agents adjust their frequencies to match neighbors:

1. **Initial state** - Different frequencies, phases drift apart
2. **Frequency locking** - Agents match frequencies
3. **Phase alignment** - With same frequency, phases can align
4. **Full synchronization** - Same frequency AND phase

## Frequency and Goals

### Fast Frequency Goals

Goals requiring quick coordination:

```go
// MinimizeLatency - Very fast (50-100ms)
client := emerge.MinimizeLatency(scale.Tiny)

// MinimizeAPICalls - Fast (200-500ms)
client := emerge.MinimizeAPICalls(scale.Medium)
```

### Slow Frequency Goals

Goals with relaxed timing:

```go
// MaintainRhythm - Steady (1-5s)
client := emerge.MaintainRhythm(scale.Small)

// SaveEnergy - Slow (5-10s)
client := emerge.SaveEnergy(scale.Large)
```

### Variable Frequency Goals

Goals that adapt frequency:

```go
// AdaptToTraffic - Dynamic based on load
client := emerge.AdaptToTraffic(scale.Medium)
```

## Frequency Dynamics

### The Kuramoto Model

Frequency evolution follows:

```
dθ/dt = ω + (K/N) × Σ sin(θ_j - θ_i)
```

Where:

- dθ/dt = frequency (rate of phase change)
- ω = natural frequency
- K = coupling strength
- Second term = frequency adjustment from neighbors

### Frequency Locking

When coupling is strong enough, frequencies lock:

- All agents converge to common frequency
- Usually close to average natural frequency
- Enables phase synchronization

### Frequency Drift

Without sufficient coupling:

- Agents maintain different frequencies
- Phases continuously drift
- No synchronization possible

## API Usage

### Default Frequencies

Each goal has optimized default frequencies:

| Goal               | Default Frequency | Rationale            |
| ------------------ | ----------------- | -------------------- |
| MinimizeAPICalls   | 200ms             | Good batching window |
| DistributeLoad     | 500ms             | Spread over time     |
| ReachConsensus     | 1s                | Time for agreement   |
| MinimizeLatency    | 50ms              | Fast response        |
| SaveEnergy         | 5s                | Minimal activity     |
| MaintainRhythm     | 1s                | Steady beat          |
| RecoverFromFailure | Variable          | Adaptive             |
| AdaptToTraffic     | Dynamic           | Load-based           |

### Frequency and Scale

Larger scales may need different frequencies:

- **Small scales** - Can use higher frequencies
- **Large scales** - Lower frequencies for stability
- **Huge scales** - Very low frequencies to prevent overload

## Performance Implications

### High Frequency Effects

**Pros:**

- Fast response times
- Quick convergence
- Real-time coordination

**Cons:**

- Higher CPU usage
- More network traffic
- Increased energy consumption

### Low Frequency Effects

**Pros:**

- Low resource usage
- Energy efficient
- Stable operation

**Cons:**

- Slow response times
- Delayed convergence
- Less responsive

## Frequency Best Practices

### DO:

- Use goal-appropriate frequencies
- Consider scale when setting frequency
- Allow frequency locking time
- Monitor for frequency drift

### DON'T:

- Set extremely high frequencies (< 10ms)
- Mix vastly different frequencies
- Change frequency during convergence
- Ignore frequency in large systems

## Troubleshooting Frequency Issues

### Problem: Agents not synchronizing

**Check:** Are frequencies locking?
**Solution:** Increase coupling strength

### Problem: System overload

**Check:** Is frequency too high?
**Solution:** Reduce frequency or scale

### Problem: Slow response

**Check:** Is frequency too low?
**Solution:** Increase frequency carefully

### Problem: Frequency oscillations

**Check:** Competing frequency adjustments
**Solution:** Stabilize with FrequencyLock strategy

## Advanced Concepts

### Frequency Modulation

Dynamically adjust frequency based on:

- System load
- Time of day
- External events
- Resource availability

### Frequency Hierarchies

Different frequency tiers:

- Fast local coordination
- Medium regional coordination
- Slow global coordination

### Frequency Resonance

Optimal frequencies for specific scales:

- Natural resonance points
- Avoid destructive interference
- Harmonic relationships

## See Also

- [Phase](phase.md) - Relationship between frequency and phase
- [Goals](goals.md) - Frequency requirements by goal
- [Strategies](strategies.md) - How strategies handle frequency
