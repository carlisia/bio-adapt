# Biofield Package

A bio-inspired adaptive system that implements distributed agent synchronization through emergent behavior, inspired by bioelectric patterns in biological systems.

## Features

- **Autonomous Agents**: Agents with genuine autonomy that can refuse adjustments and pursue local objectives
- **Hierarchical Goal Blending**: Local and global goals are blended based on agent influence
- **Energy-Based Resource Management**: Metabolic-like constraints where actions have energy costs
- **Context-Sensitive Behavior**: Agents adapt strategies based on environmental awareness
- **Emergent Synchronization**: No central orchestrator - coordination emerges from local interactions

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/carlisia/bio-adapt/biofield"
)

func main() {
    // Define target state
    goal := biofield.State{
        Phase:     0,
        Frequency: 100 * time.Millisecond,
        Coherence: 0.9,
    }
    
    // Create swarm of autonomous agents
    swarm := biofield.NewSwarm(100, goal)
    
    // Measure initial coherence
    fmt.Printf("Initial coherence: %.3f\n", swarm.MeasureCoherence())
    
    // Run autonomous synchronization
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

- **Agent**: Autonomous entity with local goals and decision-making capability
- **State**: System configuration that agents work toward (like a morphological target)
- **DecisionMaker**: Interface for autonomous decision-making (upgradeable to neural networks)
- **GoalManager**: Blends local and global objectives hierarchically
- **ResourceManager**: Handles energy/resource allocation with metabolic-like constraints
- **Swarm**: Collection of agents achieving synchronization through emergent behavior

### Key Properties

1. **Multi-Scale Goal Structure**: Individual preferences blend with collective objectives
2. **Genuine Agency**: Agents can refuse adjustments based on energy, stubbornness, and goal alignment
3. **Local Interactions**: Agents only know about their immediate neighbors
4. **Emergent Behavior**: Global synchronization emerges from local decisions

## Testing

Run tests with the `nogossip` build tag to exclude distributed gossip functionality:

```bash
go test -tags nogossip -v ./biofield/...
```

For full tests including convergence:

```bash
go test -tags nogossip ./biofield/...
```

## Performance

- ~800ms convergence for 100 agents (vs 500ms centralized)
- O(log N * log N) convergence with gossip protocol
- Probabilistic but robust convergence
- Graceful degradation under agent failures

## Upgrade Paths

The modular design allows upgrading components without architectural changes:

1. **Neural Decision Making**: Replace `SimpleDecisionMaker` with neural network implementation
2. **Evolutionary Strategies**: Add strategy evolution based on success
3. **Full Metabolism**: Implement complete metabolic model with production, consumption, and trade
4. **Advanced Gossip**: Upgrade to more sophisticated gossip protocols when needed

## Biological Inspiration

This implementation is inspired by bioelectric patterns that act as "attractor states" in biological systems, similar to how a salamander limb bud "knows" to become a complete limb through bioelectric pattern memory.

## License

See the main project LICENSE file.