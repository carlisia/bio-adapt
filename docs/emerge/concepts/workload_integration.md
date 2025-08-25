# Workload Integration

## Overview

In bio-adapt, **your workload wraps around emerge agents**. Emerge agents provide the synchronization mechanism (the "when"), while your application code decides what work to perform (the "what").

## Architecture Pattern

```
Your Application
    ↓
Workload Logic (what to do)
    ↓
Emerge Agent (when to do it)
    ↓
Synchronization Physics
```

## Key Concept: Separation of Concerns

### Emerge Agent Provides

- **Phase synchronization** - Coordinated timing
- **Frequency control** - Rate of actions
- **Energy management** - Resource constraints
- **Strategy switching** - Adaptive convergence

### Your Workload Provides

- **Business logic** - What work needs to be done
- **Action execution** - The actual operations
- **Result handling** - Processing outcomes
- **Domain-specific decisions** - Application behavior

## Implementation Pattern

### Basic Structure

```go
// Your workload wraps an emerge agent
type MyWorkload struct {
    id          string
    emergeAgent *agent.Agent  // The emerge agent for synchronization

    // Your application state
    pendingWork []Task
    results     []Result
}

// Start your workload
func (w *MyWorkload) Start(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-time.After(100 * time.Millisecond):
            // Check emerge agent's phase
            phase := w.emergeAgent.Phase()

            // Decide whether to act based on phase
            if w.shouldAct(phase) {
                w.performWork()
            }
        }
    }
}
```

### Real Example: API Batching

```go
type APIWorkload struct {
    emergeAgent  *agent.Agent
    pendingCalls []APICall
    batchManager *BatchManager
}

func (w *APIWorkload) Run(ctx context.Context) {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Emerge tells us WHEN we're synchronized
            coherence := w.emergeAgent.Coherence()

            // We decide WHAT to do when synchronized
            if coherence > 0.8 && len(w.pendingCalls) > 0 {
                // We're synchronized - batch our calls!
                w.batchManager.AddBatch(w.pendingCalls)
                w.pendingCalls = nil
            }
        }
    }
}
```

## Design Philosophy

### Why This Architecture?

1. **Clean separation** - Synchronization logic stays separate from business logic
2. **Flexibility** - Any workload type can use emerge synchronization
3. **Composability** - Multiple emerge strategies work with any workload
4. **Testability** - Can test synchronization and business logic independently

### Analogy: Dancers and Metronome

Think of it like dancers (workloads) following a metronome (emerge agents):

- The metronome provides the beat (phase synchronization)
- Each dancer performs their own moves (business logic)
- When all dancers sync to the same beat, they can coordinate (batch operations)

## Common Patterns

### Pattern 1: Threshold-Based Action

```go
func (w *Workload) shouldAct() bool {
    phase := w.emergeAgent.Phase()
    // Act when phase crosses zero (once per cycle)
    return phase < 0.1 && w.lastPhase > 6.0
}
```

### Pattern 2: Coherence-Based Batching

```go
func (w *Workload) processBatch() {
    coherence := w.emergeAgent.Coherence()
    if coherence > 0.85 {
        // High synchronization - safe to batch
        w.sendBatch(w.pendingWork)
    } else {
        // Low synchronization - process individually
        w.processIndividually(w.pendingWork)
    }
}
```

### Pattern 3: Energy-Aware Processing

```go
func (w *Workload) performWork() {
    energy := w.emergeAgent.Energy()
    if energy > 50 {
        // High energy - can do expensive operations
        w.doExpensiveWork()
    } else {
        // Low energy - only critical work
        w.doCriticalWorkOnly()
    }
}
```

## Integration with Client API

When using the high-level client API:

```go
// Create emerge client for synchronization
client := emerge.MinimizeAPICalls(scale.Medium)

// Start the client (manages internal agents)
go client.Start(ctx)

// Your workload checks client state
for !client.IsConverged() {
    time.Sleep(100 * time.Millisecond)
}

// Now synchronized - your workload can batch
workload.BatchOperations()
```

## Best Practices

### DO:

- Keep workload logic separate from synchronization
- Let emerge handle the timing, you handle the work
- Use coherence/phase to trigger workload actions
- Design workloads to benefit from synchronization

### DON'T:

- Try to control emerge agents directly from workload
- Mix synchronization physics with business logic
- Ignore energy constraints in workload decisions
- Assume synchronization is instant

## Examples by Goal

### MinimizeAPICalls

Workloads accumulate API calls and batch them when synchronized:

```go
if client.IsConverged() {
    workload.SendBatchedAPICalls()
}
```

### DistributeLoad

Workloads spread their operations when agents are anti-synchronized:

```go
if client.Coherence() < 0.3 {
    workload.ProcessNextTask() // Good distribution
}
```

### ReachConsensus

Workloads participate in voting when partially synchronized:

```go
if client.Coherence() > 0.5 && client.Coherence() < 0.7 {
    workload.CastVote()
}
```

## Advanced Integration

### Custom Workload Types

```go
type CustomWorkload interface {
    // Called when synchronization state changes
    OnSynchronized()
    OnDesynchronized()

    // Called periodically with current phase
    OnPhaseUpdate(phase float64)

    // Energy-aware callbacks
    OnEnergyLow()
    OnEnergyRecovered()
}
```

### Workload Orchestration

```go
type WorkloadOrchestrator struct {
    emergeClient *emerge.Client
    workloads    []Workload
}

func (o *WorkloadOrchestrator) Coordinate() {
    // Check global synchronization
    if o.emergeClient.IsConverged() {
        // Coordinate all workloads
        for _, w := range o.workloads {
            w.ExecuteSynchronizedAction()
        }
    }
}
```

## See Also

- [Goals](goals.md) - Optimization objectives for workloads
- [Phase](phase.md) - Understanding timing coordination
- [Coherence](coherence.md) - Measuring synchronization
- [Energy](energy.md) - Resource constraints for workloads
- [Strategies](strategies.md) - How synchronization is achieved
