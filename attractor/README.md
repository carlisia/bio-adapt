# Attractor Package

Implements **Attractor Basin Synchronization** - a bio-inspired pattern for distributed coordination where systems naturally converge to stable states through local interactions, like a ball rolling into a valley.

## Features

- **Attractor Basins**: Stable states that systems naturally converge toward
- **Autonomous Agents**: Agents with genuine autonomy that can refuse adjustments and pursue local objectives
- **Emergent Synchronization**: No central orchestrator - coordination emerges from local interactions
- **Kuramoto Dynamics**: Based on the Kuramoto model of coupled oscillators
- **Self-Healing**: Systems automatically recover from disruptions
- **Energy-Based Constraints**: Metabolic-like resource management

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/carlisia/bio-adapt/attractor"
)

func main() {
    // Define target attractor state
    goal := attractor.State{
        Phase:     0,
        Frequency: 100 * time.Millisecond,
        Coherence: 0.9,
    }
    
    // Create swarm that will converge to attractor
    swarm, err := attractor.NewSwarm(100, goal)
    if err != nil {
        panic(err)
    }
    
    // Measure initial coherence
    fmt.Printf("Initial coherence: %.3f\n", swarm.MeasureCoherence())
    
    // Run attractor basin synchronization
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    go swarm.Run(ctx)
    
    // Wait and measure final coherence
    time.Sleep(3 * time.Second)
    fmt.Printf("Final coherence: %.3f\n", swarm.MeasureCoherence())
}
```

## Architecture

### Core Components

- **AttractorBasin**: Defines stable states and attraction forces
- **Agent**: Autonomous entity with phase, frequency, and local goals
- **Swarm**: Collection of agents achieving synchronization
- **State**: Target configuration (the attractor point)
- **ConvergenceMonitor**: Tracks and predicts convergence
- **SyncStrategy**: Different synchronization strategies (phase nudge, frequency lock, etc.)

### Key Concepts

1. **Attractor Dynamics**: States naturally "fall into" stable configurations
2. **Local Coupling**: Agents only interact with neighbors
3. **Phase Coherence**: Measured using Kuramoto order parameter
4. **Emergent Behavior**: Global synchronization from local rules
5. **Basin of Attraction**: Region where states converge to attractor

## Use Cases

- **LLM Request Batching**: Coordinate API calls into natural batches
- **Distributed Consensus**: Achieve agreement without central authority
- **Resource Scheduling**: Synchronize access to shared resources
- **IoT Coordination**: Align sensor readings and device actions
- **Microservice Orchestration**: Coordinate service interactions

## Testing

Run tests with the `nogossip` build tag to exclude distributed gossip functionality:

```bash
go test -tags nogossip -v ./attractor/...
```

For full tests including convergence:

```bash
go test -tags nogossip ./attractor/...
```

## Performance

- ~800ms convergence for 100 agents (vs 500ms centralized)
- O(log N * log N) convergence with gossip protocol
- Probabilistic but robust convergence
- Graceful degradation under agent failures
- Self-healing after disruptions

## Theory

Based on:
- **Kuramoto Model**: Mathematical model of synchronization
- **Dynamical Systems Theory**: Attractor basins and stability
- **Swarm Intelligence**: Emergent behavior from simple rules
- **Biological Synchronization**: Fireflies, cardiac pacemakers, circadian rhythms

## Upgrade Paths

The modular design allows upgrading components:

1. **Neural Decision Making**: Replace simple strategies with neural networks
2. **Adaptive Basins**: Basins that learn and adapt over time
3. **Multi-Basin Systems**: Multiple attractors with transitions
4. **Hierarchical Synchronization**: Nested levels of coordination

## License

See the main project LICENSE file.