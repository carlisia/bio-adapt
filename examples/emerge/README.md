# emerge package's bio-inspired adaptive system capabilities

## Examples Overview

### 1. Basic Synchronization (`basic_sync/`)

Demonstrates fundamental bioelectric attractor basin synchronization with a small swarm converging to a target state.

**Key concepts:**

- Creating a swarm with target state
- Monitoring convergence progress
- Measuring coherence improvements

**Run:**

```bash
task run:example -- basic_sync
```

### 2. Distributed Swarm (`distributed_swarm/`)

Shows how multiple sub-swarms can operate independently yet achieve global coherence through local interactions.

**Key concepts:**

- Multiple sub-swarms with regional biases
- Inter-swarm bridge connections
- Global coherence measurement
- Distributed coordination without central control

**Run:**

```bash
task run:example -- distributed_swarm
```

### 3. Energy Management (`energy_management/`)

Demonstrates how agents manage their energy resources and how energy constraints affect synchronization behavior.

**Key concepts:**

- Energy-constrained decision making
- Different energy profiles (high/medium/low)
- Energy replenishment simulation
- Energy-based behavior adaptation

**Run:**

```bash
task run:example -- energy_management
```

### 4. Custom Decision Makers (`custom_decision/`)

Shows how to implement custom decision-making strategies for agents, including risk-averse, aggressive, and adaptive strategies.

**Key concepts:**

- Implementing custom DecisionMaker interface
- Risk-averse strategy (avoids high costs)
- Aggressive strategy (takes risks for rewards)
- Adaptive strategy (learns from past decisions)
- Strategy performance comparison

**Run:**

```bash
task run:example -- custom_decision
```

### 5. Disruption Recovery (`disruption_recovery/`)

Demonstrates the system's resilience to various disruptions and its ability to recover and maintain coherence.

**Key concepts:**

- Random phase disruptions
- Energy depletion attacks
- Network partitions
- Stubborn agent introduction
- Cascade failures
- Recovery monitoring and analysis

**Run:**

```bash
task run:example -- disruption_recovery
```

### 6. LLM Request Batching (`llm_batching/`)

Demonstrates how bio-inspired synchronization can efficiently batch LLM API requests from multiple workloads, reducing API calls and improving system efficiency.

**Key concepts:**

- Workload synchronization for API batching
- Natural batch window formation
- Resilience to workload disruptions
- Performance metrics and efficiency gains
- Self-organizing without central coordinator

**Run:**

```bash
task run:example -- llm_batching
```

### 7. Monitoring and Metrics (`monitoring_metrics/`)

Comprehensive monitoring of swarm behavior, including real-time metrics, performance analysis, and visualization data.

**Key concepts:**

- Real-time metrics collection
- Statistical analysis (mean, std dev, variance)
- Agent-level behavior analysis
- Network topology analysis
- Visualization data export
- Performance benchmarking

**Run:**

```bash
task run:example -- monitoring_metrics
```

## Key Attractor Concepts

All examples demonstrate these core principles:

1. **Autonomous Agency**: Each agent makes independent decisions based on local information
2. **Emergent Synchronization**: Global coherence emerges from local interactions without central control
3. **Energy-Based Resource Management**: Agents manage limited energy resources for actions
4. **Adaptive Behavior**: Agents adjust their behavior based on context and past experiences
5. **Resilience**: The system recovers from disruptions and maintains functionality
6. **Hierarchical Goal Blending**: Agents balance local preferences with global objectives

## Running Examples

### Using Task (Recommended for quick runs)

```bash
# Build all examples first
task build:examples

# Run specific example using CLI args (simple syntax)
task run:example -- basic_sync
task run:example -- llm_batching
task run:example -- distributed_swarm
```

### Using Go Run (Alternative method)

```bash
# Run any example directly with go run
go run ./examples/emerge/basic_sync
go run ./examples/emerge/llm_batching
go run ./examples/emerge/distributed_swarm
```

### Running All Examples Sequentially

```bash
# Using task
for example in basic_sync llm_batching distributed_swarm disruption_recovery energy_management custom_decision monitoring_metrics; do
    echo "Running $example..."
    task run:example -- $example
    echo "---"
    sleep 2
done

# Or using go run
for dir in examples/emerge/*/; do
    if [ -f "$dir/main.go" ]; then
        echo "Running $dir..."
        go run "$dir"
        echo "---"
        sleep 2
    fi
done
```

## Extending the Examples

These examples provide templates for:

- Creating custom decision strategies
- Implementing monitoring systems
- Testing disruption scenarios
- Building distributed systems
- Managing resource constraints
- Analyzing emergent behavior

Feel free to modify and extend these examples for your specific use cases!
