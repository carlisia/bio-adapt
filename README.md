# Bio-adapt

Goal-directed coordination for concurrent and distributed systems, inspired by biological intelligence.

Drawing from [Dr. Michael Levin](https://grokkingtech.io/people/michael-levin)'s research on how biological systems reliably achieve goals through multiple pathways, bio-adapt brings these principles to software systems—from single-process concurrency to distributed architectures.

**What:** Goal-directedness, adaptive pathfinding, collective intelligence  
**How:** Decentralized algorithms that pursue goals through multiple strategies

**Why:** Instead of programming HOW (procedures), you program WHAT (goals). Go goroutines figure out the HOW through:

- Emerge: Finding when to coordinate (temporal synchronization)
- Navigate: Finding what resources to use (resource allocation)
- Glue: Finding how things work (collective understanding)

## Installation

```bash
go get github.com/carlisia/bio-adapt
```

### Quick Start - Use the Emerge Client

```go
import (
    "github.com/carlisia/bio-adapt/client/emerge"
    "github.com/carlisia/bio-adapt/emerge/scale"
)

// One-liner for API batching optimization
client := emerge.MinimizeAPICalls(scale.Medium)
err := client.Start(ctx)

// Check synchronization
if client.IsConverged() {
    // System is synchronized - safe to batch operations
}
```

📖 **[See more examples](docs/client/emerge.md)** | 🎮 **[Try the interactive demo](#quick-start-with-interactive-simulation)**

## Features

🎯 **Goal-directed** - Systems maintain target states as invariants, finding alternative paths when defaults fail  
🔄 **Multiple pathways** - Inspired by how biological systems reach goals despite perturbations  
⚡ **Emergent coordination** - Collective intelligence without central control  
🧬 **Bio-inspired principles** - Computational primitives derived from Levin's research on adaptive biological systems

## Coordination primitives

Bio-adapt provides three complementary primitives for system coordination:

### 🧲 [Emerge](docs/emerge/primitive.md) - Goal-directed synchronization

**Status:** ✅ Production-ready

Systems (concurrent or distributed) that converge on target coordination states through multiple pathways, inspired by how biological systems reliably achieve morphological goals.

- Temporal coordination (when agents act)
- Self-organizing synchronization
- Adaptive strategy switching
- Optimized for 20-2000+ agents

### ⚡ [Navigate](docs/navigate/primitive.md) - Goal-directed resource allocation

**Status:** 🚧 Coming soon

Systems that navigate resource configuration spaces to reach target allocations via multiple paths, adapting when direct routes are blocked.

- Dynamic resource allocation (what resources to use)
- Alternative path discovery
- Constraint-aware navigation

### 🔗 [Glue](docs/glue/primitive.md) - Goal-directed collective intelligence

**Status:** 📋 Planned

Collective goal-seeking enables independent agents to converge on shared understanding through local interactions, achieving insights no individual could reach alone.

- Schema discovery (how APIs work)
- Distributed hypothesis testing
- Emergent consensus

See [primitives overview](docs/primitives.md) for detailed comparison.

## Real-World Use Case Examples

### Emerge

- **API batching** - Reduce API costs by 80% through synchronized batching
- **Load distribution** - Balance work across servers without central control
- **Distributed cron** - Prevent thundering herd in scheduled tasks
- **Connection pooling** - Optimize database connections adaptively
- **Rate limiting** - Coordinate request rates across services

### Coming Soon (Navigate & Glue)

- **Dynamic resource allocation** - Navigate to optimal resource distributions
- **Failure recovery** - Find alternative resource paths when failures occur
- **Schema discovery** - Collectively understand API contracts
- **Distributed consensus** - Achieve agreement without voting

Perfect for systems with 20-2000+ concurrent agents requiring coordination.

## Documentation

📚 **[Full documentation index](docs/README.md)** - Complete documentation guide

### Quick Links

- [Primitives overview](docs/primitives.md) - Choose the right primitive
- [Client libraries](docs/client/overview.md) - Simple APIs for common use cases
- [Interactive simulation](docs/simulations/overview.md) - Try it yourself
- [Architecture](docs/architecture.md) - System design and principles

### For Developers

- [Development guide](docs/development.md) - Build, test, contribute
- [Deployment guide](docs/deployment.md) - Production guidelines
- [Testing guide](docs/testing/e2e.md) - End-to-end testing
- [API reference](https://pkg.go.dev/github.com/carlisia/bio-adapt) - Complete API docs

## Quick Start with Interactive Simulation

```bash
# Clone and run the interactive demo
git clone https://github.com/carlisia/bio-adapt
cd bio-adapt
go run ./simulations/emerge

# Try different scales
go run ./simulations/emerge -scale=large  # 1000 agents

# See all options
go run ./simulations/emerge -list
```

🎮 **[Learn more about the simulation](docs/simulations/emerge.md)** - 8 optimization goals to explore interactively!

## Contributing

We welcome contributions! See our [development guide](docs/development.md) for:

- Setting up your environment
- Running tests and benchmarks
- Submitting pull requests
- Code style guidelines

## Performance

The emerge primitive is production-optimized:

| Scale  | Agents | Convergence Time | Memory/Agent |
| ------ | ------ | ---------------- | ------------ |
| Tiny   | 20     | ~800ms           | ~5KB         |
| Small  | 50     | ~1s              | ~4KB         |
| Medium | 200    | ~2s              | ~3KB         |
| Large  | 1000   | ~5s              | ~3KB         |
| Huge   | 2000   | ~10s             | ~2KB         |

**Key optimizations:**

- Automatic storage strategy selection based on swarm size
- Grouped atomic fields for 62% faster access
- Fixed-size arrays for 45% faster neighbor iteration

See [optimization guide](docs/emerge/optimization.md) for benchmarks and details.

## Research foundation

Inspired by [Dr. Michael Levin](https://grokkingtech.io/people/michael-levin)'s research on goal-directedness in biological systems, where cells and tissues achieve target morphologies through multiple pathways despite perturbations.

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
