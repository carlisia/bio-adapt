# Request Patterns

## Overview

Request patterns describe how workloads generate tasks and make requests over time. These patterns simulate real-world scenarios where different applications have different activity profiles. Request patterns are independent of synchronization - they represent the natural rhythm of work generation, while emerge handles when that work gets coordinated.

## What are Request Patterns?

Request patterns are workload behaviors that determine:

- **When** tasks are generated
- **How many** tasks are created at once
- **How frequently** new work appears
- **Whether activity is consistent or variable**

Think of request patterns like different types of traffic:

- **Rush hour traffic** = Burst pattern (intense periods, then quiet)
- **Highway traffic** = Steady pattern (consistent flow)
- **Country road** = Sparse pattern (occasional vehicles)
- **City traffic** = Mixed pattern (unpredictable variations)

## Available Patterns

### High-Frequency Pattern

**Behavior:** Continuous stream of requests  
**Rate:** >10 requests per second per workload  
**Task Generation:** 2-4 tasks at once  
**Real-World Example:** Real-time gaming server, stock trading system

```
Time: ●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●●
      Constant high activity
```

### Burst Pattern

**Behavior:** Intense activity followed by quiet periods  
**Rate:** Variable - high during bursts, zero during quiet  
**Task Generation:** 3-7 tasks during bursts, none during quiet  
**Real-World Example:** E-commerce during flash sales, news sites during breaking news

```
Time: ●●●●●●●      ●●●●●●      ●●●●●●●●
      Burst  Quiet  Burst Quiet  Burst
```

### Steady Pattern

**Behavior:** Consistent, predictable request rate  
**Rate:** ~2 requests per second  
**Task Generation:** 1 task at regular intervals  
**Real-World Example:** Scheduled batch jobs, regular API polling

```
Time: ● ● ● ● ● ● ● ● ● ● ● ● ● ● ● ●
      Regular, evenly-spaced activity
```

### Mixed Pattern

**Behavior:** Combination of different patterns  
**Rate:** Varies between 1-5 requests per second  
**Task Generation:** Sometimes 1, sometimes 2-3 tasks  
**Real-World Example:** Typical web application with varied user activity

```
Time: ● ●● ● ●●● ● ● ●● ● ●●●● ● ●● ●
      Unpredictable variations
```

### Sparse Pattern

**Behavior:** Infrequent, irregular requests  
**Rate:** <1 request per second  
**Task Generation:** Single tasks with long gaps  
**Real-World Example:** Backup systems, monitoring alerts

```
Time: ●       ●     ●         ●    ●
      Long gaps between activity
```

## How Patterns Work with Synchronization

Request patterns and synchronization are orthogonal concepts:

1. **Patterns define the workload** - Natural rhythm of task generation
2. **Synchronization coordinates the timing** - When tasks actually execute

### Example: Burst Pattern + MinimizeAPICalls

```
Without Synchronization:
Workload A: ●●●●●     ●●●●     (bursts at random times)
Workload B:   ●●●●●     ●●●●   (bursts at different times)
Result: Many separate API calls

With Synchronization:
Workload A: ●●●●●     ●●●●     (bursts at random times)
Workload B: ●●●●●     ●●●●     (synchronized to same phase)
Result: Bursts aligned, API calls batched
```

## Pattern Selection by Goal

Different goals work better with different patterns:

### Optimal Combinations

| Goal               | Best Pattern             | Why                       |
| ------------------ | ------------------------ | ------------------------- |
| MinimizeAPICalls   | High-Frequency or Burst  | Many requests to batch    |
| DistributeLoad     | Steady or High-Frequency | Consistent load to spread |
| ReachConsensus     | Steady                   | Regular participation     |
| MinimizeLatency    | High-Frequency           | Quick response needed     |
| SaveEnergy         | Sparse                   | Minimal activity          |
| MaintainRhythm     | Steady                   | Perfect rhythm            |
| RecoverFromFailure | Mixed                    | Handles variability       |
| AdaptToTraffic     | Burst                    | Simulates traffic surges  |

### Pattern Impact on Goals

#### MinimizeAPICalls

- **High-Frequency**: Excellent (1.3x) - Lots to batch
- **Burst**: Excellent (1.3x) - Concentrated batching opportunities
- **Steady**: OK (1.0x) - Moderate batching
- **Sparse**: Poor (0.7x) - Few calls to batch

#### DistributeLoad

- **Steady**: Excellent (1.3x) - Even distribution possible
- **High-Frequency**: Excellent (1.3x) - Plenty to distribute
- **Burst**: Fair (0.9x) - Spikes hard to flatten
- **Sparse**: Poor (0.6x) - Already distributed

## Implementation Details

### Pattern Configuration

Patterns affect workload behavior:

```go
// Pattern determines task generation interval
switch pattern {
case HighFrequency:
    interval = 50ms    // Very fast
case Burst:
    interval = 100ms   // During burst
case Steady:
    interval = 500ms   // Regular
case Sparse:
    interval = 2000ms  // Slow
}
```

### Burst Pattern Logic

```go
// Burst pattern alternates between active and quiet
if pattern == Burst {
    if !inBurst {
        // 20% chance to start burst
        if random() < 0.2 {
            startBurst()
            generateTasks(3-7)
        }
    } else if burstExpired() {
        endBurst()
        // Enter quiet period
    }
}
```

### Task Batch Sizes by Pattern

| Pattern        | Tasks Generated |
| -------------- | --------------- |
| High-Frequency | 2-4 at once     |
| Burst (active) | 3-7 at once     |
| Burst (quiet)  | 0               |
| Steady         | 1 at a time     |
| Mixed          | 1-3 variable    |
| Sparse         | 1 occasionally  |

## Pattern Effects on System Behavior

### High-Frequency Pattern

**Pros:**

- Maximum batching opportunities
- Quick synchronization benefits
- High throughput

**Cons:**

- High resource usage
- Can overwhelm at large scale
- Requires fast synchronization

### Burst Pattern

**Pros:**

- Realistic traffic simulation
- Good for testing adaptation
- Natural batching during bursts

**Cons:**

- Unpredictable load
- Quiet periods may lose sync
- Harder to optimize

### Steady Pattern

**Pros:**

- Predictable behavior
- Easy to optimize
- Good for testing

**Cons:**

- Doesn't test edge cases
- May not reflect reality
- Less challenging

### Sparse Pattern

**Pros:**

- Minimal resource usage
- Good for energy testing
- Low overhead

**Cons:**

- Few coordination benefits
- Slow to show results
- May not trigger synchronization

## Choosing the Right Pattern

### For Testing Synchronization

Use **High-Frequency** or **Burst**:

- Lots of activity to coordinate
- Quick results
- Clear benefits visible

### For Realistic Simulation

Use **Mixed** or **Burst**:

- Models real-world variability
- Tests adaptation capabilities
- Challenges the system

### For Baseline Testing

Use **Steady**:

- Predictable results
- Easy to measure
- Good for comparisons

### For Energy Testing

Use **Sparse**:

- Minimal activity
- Tests efficiency
- Long-running scenarios

## Pattern Modifiers

Patterns affect synchronization effectiveness:

```
Coherence Modifier = Base × Pattern Factor × Scale Factor

Example:
MinimizeAPICalls + Burst Pattern + Medium Scale
= 1.0 × 1.3 (burst bonus) × 1.0 (scale neutral)
= 1.3x effectiveness
```

## Best Practices

### DO:

- Match patterns to real workload behavior
- Use appropriate patterns for each goal
- Consider scale when selecting patterns
- Test with multiple patterns

### DON'T:

- Use sparse patterns for batching goals
- Use high-frequency at huge scales without testing
- Ignore pattern effects on convergence
- Assume one pattern fits all scenarios

## Pattern Visualization in Simulation

The simulation shows pattern effects:

```
Pattern: Burst (optimal for this goal)
Activity: ████████░░░░████████░░░░
          Active  Quiet Active Quiet

Tasks Generated: 1,234
Batches Sent: 45 (96% reduction!)
```

## See Also

- [Workload Integration](workload-integration.md) - How workloads use patterns
- [Goals](goals.md) - Which patterns work with which goals
- [Synchronization](synchronization.md) - How patterns interact with sync
- [Energy](energy.md) - Pattern effects on energy consumption
