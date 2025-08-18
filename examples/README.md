# Biofield Package Examples

This directory contains comprehensive examples demonstrating various aspects of the biofield package's bio-inspired adaptive system capabilities.

## Examples Overview

### 1. Basic Synchronization (`basic_sync/`)

Demonstrates fundamental bioelectric attractor basin synchronization with a small swarm converging to a target state.

**Key concepts:**

- Creating a swarm with target state
- Monitoring convergence progress
- Measuring coherence improvements

**Run:**

```bash
go run examples/basic_sync/main.go
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
go run examples/distributed_swarm/main.go
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
go run examples/energy_management/main.go
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
go run examples/custom_decision/main.go
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
go run examples/disruption_recovery/main.go
```

### 6. Monitoring and Metrics (`monitoring_metrics/`)

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
go run examples/monitoring_metrics/main.go
```

## Key Biofield Concepts

All examples demonstrate these core principles:

1. **Autonomous Agency**: Each agent makes independent decisions based on local information
2. **Emergent Synchronization**: Global coherence emerges from local interactions without central control
3. **Energy-Based Resource Management**: Agents manage limited energy resources for actions
4. **Adaptive Behavior**: Agents adjust their behavior based on context and past experiences
5. **Resilience**: The system recovers from disruptions and maintains functionality
6. **Hierarchical Goal Blending**: Agents balance local preferences with global objectives

## Running All Examples

To run all examples sequentially:

```bash
for dir in examples/*/; do
    if [ -f "$dir/main.go" ]; then
        echo "Running $dir..."
        go run "$dir/main.go"
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

