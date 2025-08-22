# Composition Guide

The bio-adapt library provides three composable packages that solve orthogonal coordination problems. This document describes how to compose them effectively.

## The Three Primitives

### Overview

```text
┌─────────────────────────────────────┐
│          bio-adapt library          │
├─────────────────────────────────────┤
│                                     │
│  ┌─────────┐  ┌─────────┐  ┌──────┐│
│  │ emerge  │  │navigate │  │ glue ││
│  └────┬────┘  └────┬────┘  └──┬───┘│
│       │            │           │    │
│    Timing      Resources   Schemas │
│     (When)      (What)      (How)  │
│                                     │
└─────────────────────────────────────┘
```

### Core Questions Each Primitive Answers

| Package      | Problem Domain        | Core Question             | Output              |
| ------------ | --------------------- | ------------------------- | ------------------- |
| **emerge**   | Temporal coordination | "When should agents act?" | Synchronized phases |
| **navigate** | Resource navigation   | "What resources to use?"  | Optimal allocation  |
| **glue**     | Schema discovery      | "How does this API work?" | Inferred contracts  |

## Individual Usage

Each package can be used independently:

```go
// Just synchronization
import "github.com/carlisia/bio-adapt/emerge"
swarm := emerge.NewSwarm()
swarm.Synchronize(ctx, targetState)

// Just resource allocation
import "github.com/carlisia/bio-adapt/navigate"
navigator := navigate.NewNavigator()
navigator.AllocateResources(ctx, resourceTarget)

// Just schema discovery
import "github.com/carlisia/bio-adapt/glue"
network := glue.NewNetwork()
schema := network.SolveSchema(ctx, observations)
```

## Composition Strategies

### 1. Synchronized Resource Allocation

Combine `emerge` + `navigate` for coordinated resource changes:

```go
type SynchronizedAllocator struct {
    Timing    *emerge.Swarm
    Resources *navigate.Navigator
}

func (sa *SynchronizedAllocator) CoordinatedReallocation(ctx context.Context) error {
    // First: Synchronize all agents
    if err := sa.Timing.Synchronize(ctx, emerge.PerfectSync()); err != nil {
        return err
    }

    // Then: Navigate to new resource allocation together
    target := &navigate.ResourceState{
        CPU:    0.8,
        Memory: 0.6,
    }
    return sa.Resources.NavigateToTarget(ctx, target)
}
```

### 2. Adaptive API Integration

Combine `glue` + `emerge` for zero-downtime API migrations:

```go
type AdaptiveIntegrator struct {
    Discovery *glue.Network
    Timing    *emerge.Swarm
}

func (ai *AdaptiveIntegrator) MigrateAPI(ctx context.Context, observations []Observation) error {
    // First: Collectively discover new API schema
    newSchema, err := ai.Discovery.SolveSchema(ctx, observations)
    if err != nil {
        return err
    }

    // Then: Synchronize switchover to new schema
    switchState := &emerge.State{
        Phase:     0,
        Frequency: 100 * time.Millisecond,
    }
    return ai.Timing.Synchronize(ctx, switchState)
}
```

### 3. Full Stack Composition

Combine all three for complete adaptive systems:

```go
type AdaptiveSystem struct {
    Timing    *emerge.Swarm           // When to act
    Resources *navigate.Navigator     // What resources
    Contracts *glue.Network          // How to integrate
}

func (as *AdaptiveSystem) HandleAPIChange(ctx context.Context) error {
    // 1. Discover new API schema collectively
    observations := as.gatherObservations()
    newSchema, err := as.Contracts.SolveSchema(ctx, observations)
    if err != nil {
        return fmt.Errorf("schema discovery failed: %w", err)
    }

    // 2. Calculate resource needs for new API
    resourceTarget := as.calculateResourcesForSchema(newSchema)

    // 3. Navigate to new resource allocation
    if err := as.Resources.NavigateToTarget(ctx, resourceTarget); err != nil {
        return fmt.Errorf("resource allocation failed: %w", err)
    }

    // 4. Synchronize all agents to switch at same time
    switchState := &emerge.State{
        Phase:     0,
        Frequency: 50 * time.Millisecond,
        Coherence: 0.95,
    }
    return as.Timing.Synchronize(ctx, switchState)
}

func (as *AdaptiveSystem) calculateResourcesForSchema(schema *glue.Schema) *navigate.ResourceState {
    // Determine resource needs based on discovered schema
    resources := &navigate.ResourceState{
        Allocations: make(map[string]float64),
    }

    if schema.RequiresBatching() {
        resources.Allocations["memory"] = 0.7  // More memory for batching
        resources.Allocations["cpu"] = 0.3
    } else {
        resources.Allocations["memory"] = 0.3
        resources.Allocations["cpu"] = 0.7     // More CPU for streaming
    }

    return resources
}
```

## Composition Strategies

### Sequential Composition

Execute primitives in sequence when order matters:

```go
func SequentialComposition(ctx context.Context) error {
    // 1. First discover what we're dealing with
    schema := glueNetwork.Discover(ctx)

    // 2. Then allocate appropriate resources
    resources := navigator.AllocateForSchema(ctx, schema)

    // 3. Finally synchronize the change
    return swarm.Synchronize(ctx)
}
```

### Parallel Composition

Execute independent primitives concurrently:

```go
func ParallelComposition(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)

    // Run resource allocation and schema discovery in parallel
    g.Go(func() error {
        return navigator.OptimizeResources(ctx)
    })

    g.Go(func() error {
        return glueNetwork.UpdateSchemas(ctx)
    })

    // Wait for both to complete
    if err := g.Wait(); err != nil {
        return err
    }

    // Then synchronize
    return swarm.Synchronize(ctx)
}
```

### Feedback Loop Composition

Use output from one primitive to drive another:

```go
func FeedbackComposition(ctx context.Context) {
    for {
        // Measure synchronization quality
        coherence := swarm.MeasureCoherence()

        // Adjust resources based on coherence
        if coherence < 0.8 {
            // Need more CPU for better sync
            navigator.AdjustResources(ctx, navigate.BoostCPU())
        }

        // Check if schema assumptions still hold
        if !glueNetwork.ValidateSchema(ctx) {
            // Schema changed, re-discover
            newSchema := glueNetwork.Discover(ctx)
            // Trigger re-synchronization
            swarm.Reset()
        }

        select {
        case <-ctx.Done():
            return
        case <-time.After(100 * time.Millisecond):
            // Continue monitoring
        }
    }
}
```

## Best Practices

### 1. Start Simple

Begin with individual primitives before composing:

- Master `emerge` for synchronization
- Then add `navigate` for resource management
- Finally integrate `glue` for schema discovery

### 2. Monitor Primitive Health

Each primitive provides health metrics:

```go
type SystemHealth struct {
    Synchronization float64  // emerge: coherence metric
    // ResourceUsage   float64  // (not yet implemented) navigate: allocation efficiency
    // SchemaAccuracy  float64  // (not yet implemented) glue: consensus strength
}

func (as *AdaptiveSystem) GetHealth() SystemHealth {
    return SystemHealth{
        Synchronization: as.Timing.MeasureCoherence(),
        // ResourceUsage:   as.Resources.GetEfficiency(), // (not yet implemented)
        // SchemaAccuracy:  as.Contracts.GetConsensusStrength(), //(not yet implemented)
    }
}
```

### 3. Handle Failures Gracefully

Each primitive can fail independently:

```go
func ResilientComposition(ctx context.Context) error {
    // Try primary strategy
    if err := as.HandleAPIChange(ctx); err != nil {
        // Fall back to simpler approach
        log.Printf("Primary composition failed: %v, trying fallback", err)

        // Just synchronize without resource changes
        return as.Timing.Synchronize(ctx, emerge.DefaultState())
    }
    return nil
}
```

### 4. Use Primitive-Specific Configurations

Each primitive has its own tuning parameters:

```go
// emerge: Focus on timing
emergeConfig := emerge.Config{
    SwarmSize:        100,
    CouplingStrength: 0.5,
    UpdateInterval:   50 * time.Millisecond,
}

// navigate: Focus on resource limits
navigateConfig := navigate.Config{
    MaxCPU:          0.8,
    MaxMemory:       0.7,
    ExplorationRate: 0.2,
}

// glue: Focus on consensus
glueConfig := glue.Config{
    ConsensusThreshold: 0.6,
    HypothesisTimeout:  5 * time.Second,
    MaxAgents:         50,
}
```

## Common Use Cases

### Load Balancer Coordination

```go
// emerge: Synchronize request distribution timing
// navigate: Allocate backend resources
// glue: Discover backend API changes
```

### Database Migration

```go
// glue: Discover schema differences
// navigate: Allocate migration resources
// emerge: Coordinate cutover timing
```

### Microservice Mesh

```go
// emerge: Synchronize circuit breaker states
// navigate: Manage resource quotas
// glue: Track service contract evolution
```

### Stream Processing Pipeline

```go
// emerge: Coordinate checkpoint timing
// navigate: Balance processing resources
// glue: Adapt to schema changes in stream
```

## Conclusion

The three bio-adapt primitives are designed to work independently or together. Like biological systems that separate timing (circadian rhythms), resource management (metabolism), and information processing (neural networks), these primitives provide focused solutions that compose into sophisticated adaptive systems.
