# Strategies

## Overview

Strategies are the different approaches agents use to achieve synchronization goals. The system automatically switches between strategies when one approach isn't making progress, ensuring goals are reached through multiple pathways.

## Available Strategies

### PhaseNudge

**Approach:** Gentle, incremental phase adjustments  
**Energy Use:** Low (1-2 units per adjustment)  
**Best For:** Stable, long-running synchronization  
**When Used:** Default starting strategy for most goals

```go
// Automatically selected for steady synchronization
client := emerge.MaintainRhythm(scale.Medium)
```

### FrequencyLock

**Approach:** Align frequencies first, then phases  
**Energy Use:** Medium (5-10 units per adjustment)  
**Best For:** Systems with varying natural frequencies  
**When Used:** When agents have different speeds

```go
// Often used with MinimizeLatency goal
client := emerge.MinimizeLatency(scale.Small)
```

### PulseCoupling

**Approach:** Strong synchronization pulses  
**Energy Use:** High (20-50 units per pulse)  
**Best For:** Rapid synchronization, breaking out of local minima  
**When Used:** When gentle approaches fail to converge

```go
// Selected when rapid sync needed
client := emerge.MinimizeAPICalls(scale.Tiny)
```

### EnergyAware

**Approach:** Balances synchronization with energy conservation  
**Energy Use:** Adaptive (varies based on available energy)  
**Best For:** Long-running systems, resource-constrained environments  
**When Used:** SaveEnergy goal or low-energy situations

```go
// Default for energy-conscious goals
client := emerge.SaveEnergy(scale.Large)
```

### Adaptive

**Approach:** Dynamically selects strategy based on context  
**Energy Use:** Variable  
**Best For:** Complex scenarios, changing conditions  
**When Used:** When explicit strategy selection is difficult

```go
// Used with AdaptToTraffic goal
client := emerge.AdaptToTraffic(scale.Medium)
```

## Strategy Selection

### Automatic Selection

The system automatically selects strategies based on:

1. **Goal requirements** - Each goal has preferred strategies
2. **Convergence progress** - Switches if stuck
3. **Energy availability** - Uses appropriate strategy for energy level
4. **Scale** - Some strategies work better at certain scales

### Strategy Switching

When the system switches strategies:

```
PhaseNudge (10 seconds, no progress)
    ↓
FrequencyLock (align frequencies)
    ↓
PulseCoupling (force synchronization)
    ↓
EnergyAware (if energy depleted)
```

## Strategy Performance

### Convergence Speed

| Strategy      | Small Scale | Large Scale | Energy Efficiency |
| ------------- | ----------- | ----------- | ----------------- |
| PhaseNudge    | Slow        | Very Slow   | Excellent         |
| FrequencyLock | Medium      | Medium      | Good              |
| PulseCoupling | Fast        | Slow        | Poor              |
| EnergyAware   | Adaptive    | Adaptive    | Excellent         |
| Adaptive      | Variable    | Variable    | Good              |

### Strategy by Goal

| Goal               | Primary Strategy | Fallback Strategy | Energy Profile |
| ------------------ | ---------------- | ----------------- | -------------- |
| MinimizeAPICalls   | PulseCoupling    | FrequencyLock     | High burst     |
| DistributeLoad     | PhaseNudge       | EnergyAware       | Low steady     |
| ReachConsensus     | FrequencyLock    | PulseCoupling     | Medium         |
| MinimizeLatency    | FrequencyLock    | PulseCoupling     | Very high      |
| SaveEnergy         | EnergyAware      | PhaseNudge        | Minimal        |
| MaintainRhythm     | PhaseNudge       | FrequencyLock     | Low steady     |
| RecoverFromFailure | Adaptive         | EnergyAware       | Variable       |
| AdaptToTraffic     | Adaptive         | All               | Dynamic        |

## How Strategies Work

### PhaseNudge Algorithm

```
for each agent:
    calculate average neighbor phase
    nudge toward average by small amount
    respect energy constraints
```

### FrequencyLock Algorithm

```
for each agent:
    match frequency to neighbors first
    once frequencies aligned, adjust phase
    maintain frequency lock
```

### PulseCoupling Algorithm

```
for each agent:
    if phase crosses threshold:
        send strong pulse to neighbors
        force phase alignment
    consume significant energy
```

### EnergyAware Algorithm

```
for each agent:
    check available energy
    if energy high: use stronger adjustments
    if energy low: minimal adjustments
    if energy critical: pause adjustments
```

## Strategy Optimization

### For Fast Convergence

- Start with PulseCoupling (if energy available)
- Use small scales (less coordination needed)
- Ensure high coupling strength

### For Energy Efficiency

- Use PhaseNudge or EnergyAware
- Allow longer convergence times
- Reduce update frequency

### For Stability

- Prefer PhaseNudge
- Avoid frequent strategy switches
- Use moderate coupling strength

## Monitoring Strategies

### Signs of Strategy Issues

**Problem: Frequent strategy switches**

- Indicates convergence difficulties
- May need different goal or scale

**Problem: Stuck in one strategy**

- Strategy may not be appropriate
- Check energy levels and parameters

**Problem: No convergence with any strategy**

- Goal may be unreachable with current configuration
- Consider adjusting target coherence

## Advanced Strategy Concepts

### Strategy Combinations

Some scenarios benefit from multiple active strategies:

- **Hybrid approach** - Different agents use different strategies
- **Layered strategies** - Coarse then fine adjustments
- **Regional strategies** - Different strategies in different areas

### Custom Strategies

Advanced users can implement custom strategies by implementing the `DecisionMaker` interface (see emerge package documentation).

### Strategy Memory

The system remembers which strategies worked:

- Successful strategies are preferred in similar conditions
- Failed strategies are avoided
- Learning improves over time

## Best Practices

1. **Trust automatic selection** - The system usually picks well
2. **Monitor strategy switches** - Frequent switches indicate issues
3. **Consider energy** - High-energy strategies aren't always better
4. **Match scale** - Some strategies don't scale well
5. **Allow adaptation time** - Strategy switches need adjustment period

## See Also

- [Energy](energy.md) - Energy consumption by strategy
- [Goals](goals.md) - Strategy selection by goal
- [Coherence](coherence.md) - How strategies affect synchronization
