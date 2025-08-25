# Composition Guide

The bio-adapt library provides three composable packages that solve orthogonal coordination problems. This document describes how to compose them effectively.

> **Current Status**:
>
> - ✅ **emerge** - Production-ready for temporal synchronization
> - 🚧 **navigate** - Coming soon for resource allocation
> - 📋 **glue** - Planned for schema discovery

## The Three Primitives

### Overview

```text
┌─────────────────────────────────────┐
│          bio-adapt library          │
├─────────────────────────────────────┤
│                                     │
│  ┌─────────┐  ┌─────────┐  ┌──────┐ │
│  │ emerge  │  │navigate │  │ glue │ │
│  └────┬────┘  └────┬────┘  └──┬───┘ │
│       │            │           │    │
│    Timing      Resources   Schemas  │
│     (When)      (What)      (How)   │
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
// Just synchronization (currently implemented)
import "github.com/carlisia/bio-adapt/client/emerge"
import "github.com/carlisia/bio-adapt/emerge/scale"

client := emerge.MinimizeAPICalls(scale.Medium)
err := client.Start(ctx)

// Just resource allocation (coming soon)
import "github.com/carlisia/bio-adapt/navigate"
navigator := navigate.NewNavigator()
navigator.AllocateResources(ctx, resourceTarget)

// Just schema discovery (planned)
import "github.com/carlisia/bio-adapt/glue"
network := glue.NewNetwork()
schema := network.SolveSchema(ctx, observations)
```

## Composition Patterns

### 1. Synchronized Resource Allocation (Future)

Combine `emerge` + `navigate` for coordinated resource changes:

```go
type SynchronizedAllocator struct {
    Timing    *emerge.Client      // Using emerge client
    Resources *navigate.Navigator // Coming soon
}

func (sa *SynchronizedAllocator) CoordinatedReallocation(ctx context.Context) error {
    // First: Synchronize all agents
    go sa.Timing.Start(ctx)

    // Wait for convergence
    for !sa.Timing.IsConverged() {
        time.Sleep(100 * time.Millisecond)
    }

    // Then: Navigate to new resource allocation together
    target := &navigate.ResourceState{
        CPU:    0.8,
        Memory: 0.6,
    }
    return sa.Resources.NavigateToTarget(ctx, target)
}
```

### 2. Adaptive API Integration (Future)

Combine `glue` + `emerge` for zero-downtime API migrations:

```go
type AdaptiveIntegrator struct {
    Discovery *glue.Network    // Planned
    Timing    *emerge.Client   // Using emerge client
}

func (ai *AdaptiveIntegrator) MigrateAPI(ctx context.Context, observations []Observation) error {
    // First: Collectively discover new API schema
    newSchema, err := ai.Discovery.SolveSchema(ctx, observations)
    if err != nil {
        return err
    }

    // Then: Synchronize switchover to new schema
    // All agents switch at the same synchronized moment
    return ai.Timing.Start(ctx)
}
```

### 3. Full Stack Composition (Future)

Combine all three for complete adaptive systems:

```go
type AdaptiveSystem struct {
    Timing    *emerge.Client         // When to act (implemented)
    Resources *navigate.Navigator    // What resources (coming soon)
    Contracts *glue.Network          // How to integrate (planned)
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
    // The emerge client handles the synchronization details
    return as.Timing.Start(ctx)
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

## Implementation Strategies

### Sequential Composition

Execute primitives in sequence when order matters:

```go
func SequentialComposition(ctx context.Context) error {
    // 1. First discover what we're dealing with (planned)
    schema := glueNetwork.Discover(ctx)

    // 2. Then allocate appropriate resources (coming soon)
    resources := navigator.AllocateForSchema(ctx, schema)

    // 3. Finally synchronize the change (implemented)
    client := emerge.MinimizeAPICalls(scale.Medium)
    return client.Start(ctx)
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

    // Then synchronize using emerge client
    client := emerge.MinimizeAPICalls(scale.Medium)
    return client.Start(ctx)
}
```

### Feedback Loop Composition

Use output from one primitive to drive another:

```go
func FeedbackComposition(ctx context.Context, client *emerge.Client) {
    for {
        // Measure synchronization quality
        coherence := client.Coherence()

        // Adjust resources based on coherence (coming soon)
        if coherence < 0.8 {
            // Need more CPU for better sync
            // navigator.AdjustResources(ctx, navigate.BoostCPU())
        }

        // Check if schema assumptions still hold (planned)
        // if !glueNetwork.ValidateSchema(ctx) {
        //     // Schema changed, re-discover
        //     newSchema := glueNetwork.Discover(ctx)
        //     // Trigger re-synchronization
        // }

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
    Synchronization float64  // emerge: coherence metric (implemented)
    ResourceUsage   float64  // navigate: allocation efficiency (coming soon)
    SchemaAccuracy  float64  // glue: consensus strength (planned)
}

func (as *AdaptiveSystem) GetHealth() SystemHealth {
    return SystemHealth{
        Synchronization: as.Timing.Coherence(),  // Using emerge client
        ResourceUsage:   0.0, // Coming soon with navigate
        SchemaAccuracy:  0.0, // Planned with glue
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
        return as.Timing.Start(ctx)
    }
    return nil
}
```

### 4. Use Primitive-Specific Configurations

Each primitive has its own tuning parameters:

```go
// emerge: Focus on timing (implemented)
client := emerge.Custom().
    WithGoal(goal.MinimizeAPICalls).
    WithScale(scale.Large).
    WithTargetCoherence(0.85).
    Build()

// navigate: Focus on resource limits (coming soon)
// navigateConfig := navigate.Config{
//     MaxCPU:          0.8,
//     MaxMemory:       0.7,
//     ExplorationRate: 0.2,
// }

// glue: Focus on consensus (planned)
// glueConfig := glue.Config{
//     ConsensusThreshold: 0.6,
//     HypothesisTimeout:  5 * time.Second,
//     MaxAgents:         50,
// }
```

## Common Use Cases

### Currently Available (emerge only)

#### API Batching

```go
// Minimize API calls through synchronized batching
client := emerge.MinimizeAPICalls(scale.Large)
client.Start(ctx)
```

#### Load Distribution

```go
// Distribute load through anti-phase synchronization
client := emerge.DistributeLoad(scale.Medium)
client.Start(ctx)
```

### Future Use Cases (with all three primitives)

#### Load Balancer Coordination

- **emerge**: Synchronize request distribution timing (available now)
- **navigate**: Allocate backend resources (coming soon)
- **glue**: Discover backend API changes (planned)

#### Database Migration

- **glue**: Discover schema differences (planned)
- **navigate**: Allocate migration resources (coming soon)
- **emerge**: Coordinate cutover timing (available now)

#### Microservice Mesh

- **emerge**: Synchronize circuit breaker states (available now)
- **navigate**: Manage resource quotas (coming soon)
- **glue**: Track service contract evolution (planned)

#### Stream Processing Pipeline

- **emerge**: Coordinate checkpoint timing (available now)
- **navigate**: Balance processing resources (coming soon)
- **glue**: Adapt to schema changes in stream (planned)

## Conclusion

The three bio-adapt primitives are designed to work independently or together. Like biological systems that separate timing (circadian rhythms), resource management (metabolism), and information processing (neural networks), these primitives provide focused solutions that compose into sophisticated adaptive systems.
