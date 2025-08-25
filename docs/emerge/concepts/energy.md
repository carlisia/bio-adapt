# Energy

## Overview

Energy represents an agent's capacity to perform actions. It's a resource constraint that prevents infinite adjustments and ensures stable convergence. Energy depletes with actions and recovers over time.

## Energy Mechanics

### Energy States

- **Full** (80-100%) - Agent can perform any action
- **Available** (20-80%) - Agent can act but may be selective
- **Low** (5-20%) - Agent conserves energy, minimal actions
- **Depleted** (0-5%) - Agent cannot act until recovery

### Energy Consumption

Different actions consume different amounts of energy:

- **Phase adjustment** - 1-5 units per adjustment
- **Frequency change** - 5-10 units per change
- **Strategy switch** - 10-20 units per switch
- **Pulse coupling** - 20-50 units per pulse

### Energy Recovery

Energy recovers continuously over time:

```go
newEnergy = currentEnergy + (recoveryRate * deltaTime)
```

Default recovery rate: 5 units per second

## Energy in the API

### Default Energy Settings

Each scale has optimized energy parameters:

- **Tiny/Small** - High energy, fast recovery (quick convergence)
- **Medium** - Balanced energy and recovery
- **Large/Huge** - Conservative energy use (stability)

### Energy and Strategies

Different strategies use energy differently:

#### EnergyAware Strategy

Explicitly manages energy for efficiency:

```go
// Automatically selected for SaveEnergy goal
client := emerge.SaveEnergy(scale.Medium)
// Uses EnergyAware strategy by default
```

#### High-Energy Strategies

PulseCoupling uses more energy but converges faster:

- Good for small swarms with plenty of energy
- May deplete large swarms if overused

#### Low-Energy Strategies

PhaseNudge uses minimal energy:

- Slower but sustainable
- Good for long-running systems

## Energy and Convergence

### Impact on Convergence

1. **High energy** = Faster convergence, risk of oscillation
2. **Moderate energy** = Balanced convergence and stability
3. **Low energy** = Slow but stable convergence

### Energy Depletion Effects

When agents run out of energy:

- Cannot adjust phase (stuck in current position)
- Cannot respond to neighbors (temporary isolation)
- Coherence may degrade temporarily
- System recovers as energy replenishes

## Energy Management

### Preventing Depletion

1. **Use appropriate strategies** - Match strategy to available energy
2. **Scale recovery rates** - Larger swarms may need higher recovery
3. **Implement backoff** - Reduce action frequency when energy is low
4. **Monitor energy levels** - Track average swarm energy

### Energy-Efficient Patterns

```go
// For long-running systems
client := emerge.Custom().
    WithGoal(goal.MaintainRhythm).
    WithScale(scale.Medium).
    Build()
// Uses energy-efficient defaults
```

### High-Performance Patterns

```go
// For quick convergence (energy-intensive)
client := emerge.Custom().
    WithGoal(goal.MinimizeLatency).
    WithScale(scale.Tiny).
    Build()
// Can afford high energy use with small scale
```

## Energy Monitoring

### Signs of Energy Problems

**Symptom: Convergence stalls**

- Check: Average energy levels
- Solution: Increase recovery rate or reduce action frequency

**Symptom: Oscillating coherence**

- Check: Energy depletion cycles
- Solution: Smooth energy consumption

**Symptom: Very slow convergence**

- Check: Conservative energy settings
- Solution: Increase initial energy or recovery rate

### Debugging Energy Issues

```go
// Advanced: Check agent energy levels
agents := client.Swarm().Agents()
for id, agent := range agents {
    energy := agent.Energy()
    if energy < 10 {
        fmt.Printf("Agent %s low energy: %.1f\n", id, energy)
    }
}
```

## Energy Best Practices

### DO:

- Let the system manage energy automatically
- Use scale-appropriate defaults
- Monitor for energy depletion in production
- Allow recovery time between operations

### DON'T:

- Manually set extreme energy values
- Ignore energy in long-running systems
- Force rapid changes without energy consideration
- Assume infinite energy availability

## Energy and Goals

Different goals have different energy profiles:

| Goal               | Energy Usage   | Recovery Needs | Typical Pattern     |
| ------------------ | -------------- | -------------- | ------------------- |
| MinimizeAPICalls   | High burst     | Moderate       | Burst then maintain |
| DistributeLoad     | Low continuous | Low            | Steady state        |
| ReachConsensus     | Medium bursts  | High           | Cyclic              |
| MinimizeLatency    | Very high      | Very high      | Continuous high     |
| SaveEnergy         | Minimal        | Minimal        | Conservative        |
| MaintainRhythm     | Low steady     | Low            | Predictable         |
| RecoverFromFailure | Variable       | Adaptive       | Responsive          |
| AdaptToTraffic     | Dynamic        | Dynamic        | Load-dependent      |

## Advanced Energy Concepts

### Energy Budgets

Systems can implement energy budgets:

- Total energy pool shared by all agents
- Prevents system-wide resource exhaustion
- Useful for cloud cost management

### Energy-Based Scheduling

Coordinate actions based on energy availability:

- High-energy tasks when energy is plentiful
- Conservation mode when energy is scarce
- Automatic load shedding

### Energy Cascades

Energy can trigger cascading effects:

- Low energy → reduced activity → lower coherence
- Energy recovery → increased activity → improved coherence
- System naturally self-regulates

## See Also

- [Strategies](strategies.md) - Energy usage by strategy
- [Goals](goals.md) - Energy requirements by goal
- [Scale Definitions](../emerge/scales.md) - Energy parameters by scale
