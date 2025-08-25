// Package emerge provides an idiomatic Go client for the emerge synchronization framework.
//
// The emerge framework enables distributed agents to achieve collective behavior through
// local interactions, without central coordination. It's based on the Kuramoto model of
// coupled oscillators, where agents adjust their phase and frequency to synchronize.
//
// # Quick Start
//
// The simplest way to create a client:
//
//	client := emerge.MinimizeAPICalls(scale.Medium)
//	err := client.Start(ctx)
//
// # Core Concepts
//
// Agents: Autonomous units with phase and frequency that interact locally.
//
// Swarm: Collection of agents achieving emergent synchronization.
//
// Coherence: Measure of synchronization quality (0 = chaos, 1 = perfect sync).
//
// Goals: High-level objectives that determine synchronization patterns:
//   - MinimizeAPICalls: Agents synchronize to batch operations
//   - DistributeLoad: Agents anti-synchronize to spread load
//
// Scales: Predefined swarm sizes with optimized parameters:
//   - Tiny: 10-20 agents (high coherence possible)
//   - Small: 50 agents (balanced)
//   - Medium: 100 agents (good for most cases)
//   - Large: 1000+ agents (enterprise scale)
//
// # Usage Patterns
//
// Convenience constructors for common scenarios:
//
//	// API optimization scenario
//	client := emerge.MinimizeAPICalls(scale.Large)
//
//	// Load distribution scenario
//	client := emerge.DistributeLoad(scale.Medium)
//
//	// Just use defaults
//	client := emerge.Default()
//
// Builder pattern for custom configuration:
//
//	client := emerge.New().
//	    WithGoal(goal.MinimizeAPICalls).
//	    WithScale(scale.Large).
//	    WithTargetCoherence(0.95).
//	    Build()
//
// Functional options pattern:
//
//	client := emerge.NewWithOptions(
//	    emerge.WithGoalOption(goal.DistributeLoad),
//	    emerge.WithScaleOption(scale.Medium),
//	    emerge.WithCoherenceOption(0.90),
//	)
//
// # Client Lifecycle
//
//	// Create client
//	client, err := emerge.MinimizeAPICalls(scale.Medium)
//	if err != nil {
//	    return err
//	}
//
//	// Start synchronization
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	go func() {
//	    if err := client.Start(ctx); err != nil {
//	        log.Printf("swarm error: %v", err)
//	    }
//	}()
//
//	// Monitor synchronization
//	for {
//	    if client.IsConverged() {
//	        fmt.Printf("Converged! Coherence: %.2f\n", client.Coherence())
//	        break
//	    }
//	    time.Sleep(100 * time.Millisecond)
//	}
//
//	// Clean shutdown
//	client.Stop()
//
// # Advanced Usage
//
// Access underlying swarm for advanced operations:
//
//	swarm := client.Swarm()
//	agents := client.Agents()
//
//	for id, agent := range agents {
//	    // Custom agent operations
//	}
//
// # Performance Considerations
//
// The framework automatically optimizes based on swarm size:
//   - Small swarms (<100 agents): Uses sync.Map for flexibility
//   - Large swarms (≥100 agents): Uses slice-based storage for cache locality
//   - Very large swarms (≥1000 agents): Enables additional optimizations
//
// Target coherence should be realistic for the swarm size:
//   - Tiny swarms can achieve 0.95+ coherence
//   - Medium swarms typically achieve 0.85-0.90
//   - Large swarms may max out at 0.70-0.80
//
// # Thread Safety
//
// All client methods are thread-safe and can be called concurrently.
// The underlying swarm handles synchronization internally.
package emerge
