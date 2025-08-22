# Bio-adapt

Goal-directed distributed coordination inspired by biological intelligence.

Drawing from Michael Levin's research on how biological systems reliably achieve goals through multiple pathways, bio-adapt brings these principles to distributed systems—without implementing biological mechanisms directly.

**What:** Goal-directedness, adaptive pathfinding, collective intelligence  
**How:** Distributed algorithms that pursue goals through multiple strategies

**Why:** Instead of programming HOW (procedures), you program WHAT (goals). Go goroutines figure out the HOW through:

- Emerge: Finding when to coordinate (temporal synchronization)
- Navigate: Finding what resources to use (resource allocation)
- Glue: Finding how things work (collective understanding)

## Quick start

```bash
go get github.com/carlisia/bio-adapt
```

```go
import "github.com/carlisia/bio-adapt/emerge"

// Create goal-directed swarm that pursues synchronization target
target := emerge.Pattern{
    Frequency: 200 * time.Millisecond, // Target coordination interval
    Coherence: 0.9,                    // Goal: 90% synchronization
}
swarm, _ := emerge.NewSwarm(20)
swarm.AchieveSynchronization(ctx, target) // Pursues goal through multiple strategies
```

## Features

🎯 **Goal-directed** - Systems maintain target states as invariants, finding alternative paths when defaults fail  
🔄 **Multiple pathways** - Inspired by how biological systems reach goals despite perturbations  
⚡ **Emergent coordination** - Collective intelligence without central control  
🧬 **Bio-inspired principles** - Computational patterns derived from Levin's research on adaptive biological systems

## Patterns

Bio-adapt provides three complementary patterns for distributed coordination:

### 🧲 [Emerge](docs/emerge/pattern.md) - Goal-directed synchronization

**Status:** ✅ Production-ready

Distributed systems that converge on target coordination states through multiple pathways, inspired by how biological systems reliably achieve morphological goals.

- Temporal coordination (when agents act)
- Self-organizing synchronization
- Adaptive strategy switching

### ⚡ [Navigate](docs/navigate/pattern.md) - Goal-directed resource allocation

**Status:** 🚧 Coming soon

Systems that navigate resource configuration spaces to reach target allocations via multiple paths, adapting when direct routes are blocked.

- Dynamic resource allocation (what resources to use)
- Alternative path discovery
- Constraint-aware navigation

### 🔗 [Glue](docs/glue/pattern.md) - Goal-directed collective intelligence

**Status:** 📋 Planned

Collective goal-seeking enables distributed agents to converge on shared understanding through local interactions, achieving insights no individual could reach alone.

- Schema discovery (how APIs work)
- Distributed hypothesis testing
- Emergent consensus

See [patterns overview](docs/patterns.md) for detailed comparison.

## Use cases

- **API batching** - Goal: minimize API calls; emerge finds optimal coordination timing
- **Distributed synchronization** - Goal: achieve coherence; multiple strategies ensure convergence
- **Self-healing systems** - Goal: maintain service levels; alternative paths when failures occur
- **Load balancing** - Goal: optimal resource usage; navigate finds best allocation paths

Perfect for systems that need to maintain goals despite disruptions, with 100+ agents requiring coordination.

## Documentation

### Getting started

- [Patterns overview](docs/patterns.md) - Choose the right pattern
- [Architecture](docs/architecture.md) - System design and principles
- [Examples](examples/) - Production-ready code samples

### Guides

- [Development](docs/development.md) - Build, test, contribute
- [Deployment](docs/deployment.md) - Production guidelines
- [API reference](https://pkg.go.dev/github.com/carlisia/bio-adapt) - Complete API docs

### Pattern-specific docs

- [Emerge documentation](docs/emerge/pattern.md) - Goal-directed synchronization
- [Navigate documentation](docs/navigate/pattern.md) - Goal-directed resource allocation (coming soon)
- [Glue documentation](docs/glue/pattern.md) - Goal-directed collective intelligence (planned)
- [Orchestration guide](docs/orchestration.md) - Composing patterns for complex systems

## Examples

🔄 [Basic synchronization](examples/emerge/basic_sync) - Learn the fundamentals  
📦 [LLM batching](examples/emerge/llm_batching) - Reduce API calls by 80%  
🌐 [Distributed swarm](examples/emerge/distributed_swarm) - Multi-region coordination  
💪 [Disruption recovery](examples/emerge/disruption_recovery) - Self-healing demos

## Development

See [development guide](docs/development.md) for setup, building, testing, and contributing.

## Performance

The emerge pattern is optimized for production with 1000+ agents:

- **Sub-linear convergence** - Better performance at scale
- **~2KB memory per agent** - Efficient resource usage
- **<1ms convergence latency** - Fast coordination
- **Automatic optimization** - Adapts storage strategy by swarm size

See [emerge optimization guide](docs/emerge/optimization.md) for details.

## Research foundation

Inspired by Dr. Michael Levin's research on goal-directedness in biological systems, where cells and tissues achieve target morphologies through multiple pathways despite perturbations.

Key concepts adapted:

- **Goal-directedness** - Systems that maintain target states as invariants
- **Multiple pathways** - Alternative routes to achieve the same outcome
- **Collective intelligence** - Problem-solving that emerges from local interactions
- **Adaptive navigation** - Finding new solutions when defaults are blocked

Implementation foundations:

- Kuramoto model for synchronization dynamics (emerge)
- Pathfinding algorithms for resource navigation (navigate)
- Distributed consensus protocols for collective intelligence (glue)

## License

MIT - See [LICENSE](LICENSE)
