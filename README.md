# Bio-Adapt

**Bio-inspired adaptive systems for Go** - Self-organizing coordination patterns that achieve goals despite disruptions.

## Overview

Bio-Adapt brings biological intelligence principles to distributed systems. Just as cells self-organize to heal wounds or regenerate limbs, your systems can adaptively coordinate to maintain desired states even when disrupted.

Inspired by Dr. Michael Levin's bioelectric research showing how cellular networks achieve reliable outcomes through goal-directed behavior rather than fixed instruction sequences.

## Key Benefits

🎯 **Goal-Directed** - Systems pursue target states, not rigid procedures
🔄 **Self-Healing** - Automatic recovery from disruptions
⚡ **Emergent Coordination** - No central orchestrator needed
🧬 **Biologically Inspired** - Proven patterns from nature

## Core Patterns

### 🧲 Attractor Basin Synchronization

**Package:** `emerge`
**Use Case:** Coordinate timing across distributed workloads
**Example:** Batch LLM API calls naturally without central control

### ⚡ Morphospace Navigation

**Package:** `bioelectric` _(coming soon)_
**Use Case:** Dynamic resource allocation around bottlenecks
**Example:** Reroute processing when nodes fail

### 🔗 Cognitive Glue Networks

**Package:** `glue` _(coming soon)_
**Use Case:** Emergent consensus through collective problem-solving
**Example:** Distributed schema evolution

## When to Use Bio-Adapt

✅ **Perfect for:**

- Systems with 100+ concurrent workloads needing coordination
- API rate limiting and request batching
- Self-healing distributed systems
- Workload synchronization without central control
- Natural load balancing across resources

❌ **Not ideal for:**

- Systems requiring strict deterministic guarantees
- Simple request-response patterns
- Tightly coupled synchronous operations

## Quick Start

```bash
# Clone the repository
git clone https://github.com/carlisia/bio-adapt
cd bio-adapt

# Build and run examples
task build:examples
task run:example -- llm_batching
```

```go
// Synchronize workloads naturally
import "github.com/carlisia/bio-adapt/emerge"

// Define target state
goal := emerge.State{
    Phase:     0,                      // Alignment point
    Frequency: 200 * time.Millisecond, // Batch window
    Coherence: 0.9,                    // 90% sync target
}

// Create self-organizing swarm
swarm, _ := emerge.NewSwarm(20, goal)
swarm.Run(ctx)
```

## Examples

🔄 **[Basic Synchronization](examples/emerge/basic_sync)** - Learn the fundamentals
📦 **[LLM Batching](examples/emerge/llm_batching)** - Reduce API calls by 80%
🌐 **[Distributed Swarm](examples/emerge/distributed_swarm)** - Multi-region coordination
💪 **[Disruption Recovery](examples/emerge/disruption_recovery)** - Self-healing demos

See [examples/](examples/) for all available examples.

## Documentation

- [Emerge Package Guide](emerge/README.md) - Deep dive into synchronization
- [Examples Overview](examples/emerge/README.md) - Hands-on tutorials
- [API Reference](https://pkg.go.dev/github.com/carlisia/bio-adapt) - Complete API docs

## Development

```bash
# Build everything
task build:all

# Run tests
task test

# Run linter
task lint

# Format code
task fmt

# Check for vulnerabilities
task vuln

# Development mode (auto-rebuild)
task dev  # requires entr

# Clean build artifacts
task clean
```

## Contributing

Contributions welcome! Areas of interest:

- Additional synchronization strategies
- Performance optimizations
- New bio-inspired patterns
- Real-world use cases

## Research Foundation

Based on groundbreaking research:

- Dr. Michael Levin's work on bioelectric networks and regeneration
- Kuramoto model of coupled oscillators
- Swarm intelligence and emergent behavior
- Dynamical systems and attractor theory

## License

MIT - See [LICENSE](LICENSE) file
