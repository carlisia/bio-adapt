# Simulations

Interactive demonstrations of bio-adapt's emergent synchronization capabilities for distributed workload optimization.

## Quick Start

```bash
# Run with default settings (MinimizeAPICalls goal, Tiny scale)
go run ./simulations/emerge

# Run with different scales
go run ./simulations/emerge -scale=medium  # 200 workloads
go run ./simulations/emerge -scale=large   # 1000 workloads

# List available scales
go run ./simulations/emerge -list

# Run with custom timeout
go run ./simulations/emerge -timeout=10m
```

## The Simulation

### Workload Optimization through Emergent Behavior

This simulation demonstrates how independent workloads (ETL pipelines, ML training jobs, web servers, IoT sensors, etc.) learn to optimize their operations through emergent synchronization - without any central coordinator.

**Key Features:**

- **8 optimization goals** - Each with goal-specific workload types
- **5 request patterns** - From high-frequency streams to sparse requests
- **5 scale options** - From 20 to 2000 workloads
- **Real-time visualization** - Watch coherence evolve in real time
- **Fully interactive** - Switch goals, patterns, and scales instantly
- **Self-healing** - Automatic recovery from disruptions
- **Cryptographically secure randomization** - For realistic workload behavior

## Interactive Controls

### Navigation

- `Q` - Quit the simulation
- `R` - Reset simulation
- `D` - Disrupt synchronization (test recovery)
- `P` - Pause simulation
- `Space` - Resume simulation

### Scale Selection

- `1` - Tiny (20 workloads)
- `2` - Small (50 workloads)
- `3` - Medium (200 workloads)
- `4` - Large (1000 workloads)
- `5` - Huge (2000 workloads)

### Goal Selection

- `B` - Batch (MinimizeAPICalls) - Combine requests to reduce costs
- `L` - Load (DistributeLoad) - Balance work across servers
- `C` - Consensus (ReachConsensus) - Distributed agreement
- `T` - laTency (MinimizeLatency) - Reduce response time
- `E` - Energy (SaveEnergy) - Minimize power consumption
- `M` - rhythM (MaintainRhythm) - Keep consistent timing
- `F` - Failure (RecoverFromFailure) - Handle disruptions
- `A` - Adapt (AdaptToTraffic) - Respond to load changes

### Pattern Selection

- `H` - High-frequency (continuous stream)
- `U` - Burst (spikes and quiet periods)
- `Y` - Steady (consistent rate)
- `X` - Mixed (combination)
- `Z` - Sparse (infrequent)

## What You'll See

The simulation provides rich visual feedback:

### Display Components

- **Title Bar** - Current goal and scale
- **Configuration Panel** - Active goal, pattern, and scale details
- **Problem/Solution Panel** - Context-specific description
- **Workload List** - Live view of all workloads with icons, phases, and activity
- **Coherence Gauges** - Target vs current synchronization levels
- **Metrics Panel** - Real-time performance metrics
- **Swarm Visualization** - Phase distribution visualization
- **Interactive Menu** - Available keyboard commands

### Visual Indicators

- **Workload Icons** - 🤖 (ML), 📊 (ETL), 🌐 (Web), 🌡️ (Sensor), etc.
- **Activity Levels** - Color-coded: burst (red), active (yellow), steady (green), quiet (blue)
- **Sync Quality** - 🟢 Good (>70%), 🟡 Partial (40-70%), 🔴 Poor (<40%)
- **Special States** - Paused, Disrupted, Reset indicators

## Architecture

````bash
simulations/
├── emerge/                   # Main simulation
│   ├── main.go              # Entry point, CLI handling, goal switching
│   ├── config.go            # Configuration management
│   ├── simulation/          # Core simulation logic
│   │   ├── agent.go         # Workload implementation (secure random)
│   │   ├── batch_manager.go # Batch processing and metrics
│   │   ├── builder.go       # Clean builder using emerge client
│   │   ├── metrics.go       # Metrics collection
│   │   ├── simulation.go    # Main simulation logic
│   │   ├── types.go         # Shared types and interfaces
│   │   └── pattern/         # Request patterns
│   │       └── pattern.go   # Pattern definitions and modifiers
│   └── ui/                  # User interface
│       └── runner.go        # UI coordination
├── display/                 # Shared display components
│   ├── controller.go        # Keyboard input handling
│   ├── display.go          # Terminal UI (termdash)
│   ├── text_display.go     # Text-only display
│   └── interfaces.go       # Display interfaces
└── client/
    └── emerge/              # Clean emerge client API
        ├── emerge.go        # Core client
        ├── builder.go       # Fluent builder API
        └── custom.go        # Convenience methods

## Key Concepts

### Goals

Each goal represents a different optimization objective:

- **MinimizeAPICalls**: Batch requests to reduce API costs
- **DistributeLoad**: Spread work evenly across servers
- **ReachConsensus**: Achieve distributed agreement
- **MinimizeLatency**: Reduce response times
- **SaveEnergy**: Minimize power consumption
- **MaintainRhythm**: Keep consistent timing
- **RecoverFromFailure**: Handle disruptions gracefully
- **AdaptToTraffic**: Respond to load changes

### Workload Types

Different workload types are selected based on the goal:

- Data ETL pipelines (for batching)
- Web servers (for load distribution)
- Consensus nodes (for distributed agreement)
- Game servers (for low latency)
- IoT sensors (for energy saving)
- Cron jobs (for rhythm maintenance)

### Patterns

Request patterns affect how workloads generate tasks:

- **High-frequency**: Continuous stream (>10 requests/sec)
- **Burst**: Sudden spikes followed by quiet periods
- **Steady**: Consistent, predictable rate
- **Mixed**: Combination of patterns
- **Sparse**: Infrequent, irregular (<1 request/sec)

### Scales

Different scales demonstrate various coordination challenges:

- **Tiny** (20): Quick demos, tight coordination
- **Small** (50): Team-sized coordination
- **Medium** (200): Department-scale systems
- **Large** (1000): Enterprise deployments
- **Huge** (2000): Cloud-scale operations

## Optimal Combinations

Some goal/pattern/scale combinations work better than others:

| Goal               | Best Pattern         | Best Scale | Why                          |
| ------------------ | -------------------- | ---------- | ---------------------------- |
| MinimizeAPICalls   | High-frequency/Burst | Medium+    | More requests to batch       |
| DistributeLoad     | Steady               | Large      | Even distribution at scale   |
| ReachConsensus     | Steady               | Small      | Regular participation        |
| MinimizeLatency    | High-frequency       | Tiny       | Quick response, low overhead |
| SaveEnergy         | Sparse               | Large      | Minimal activity             |
| MaintainRhythm     | Steady               | Medium     | Perfect timing               |
| RecoverFromFailure | Mixed                | Any        | Handles variability          |
| AdaptToTraffic     | Burst                | Large      | Simulates real surges        |

## Command-Line Options

```bash
-scale string      # Swarm scale: tiny, small, medium, large, huge (default "tiny")
-interval duration # Display update interval (default 100ms)
-timeout duration  # Simulation timeout, 0 for no timeout (default 5m)
-list             # List available scales with descriptions
````

## Real-World Applications

This simulation pattern applies to:

- **API/Cloud Services**: Batch processing, cost optimization
- **Distributed Systems**: Load balancing, consensus protocols
- **IoT/Edge Computing**: Sensor coordination, power management
- **Microservices**: Service mesh optimization
- **Database Operations**: Query batching, write optimization
- **Content Delivery**: Cache coordination, CDN optimization

## Learn More

See [emerge/README.md](emerge/README.md) for detailed information about the simulation mechanics and underlying theory.
