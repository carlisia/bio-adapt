# Examples

Production-ready examples demonstrating bio-adapt's capabilities for distributed coordination.

## Quick start

```bash
# Build all examples
task build:examples

# Run an example
task run:example -- basic_sync
task run:example -- llm_batching
```

## Examples by use case

### Learning the basics

- **[basic_sync](emerge/basic_sync)** - Fundamental attractor basin synchronization
- **[energy_management](emerge/energy_management)** - Resource-aware coordination

### Production patterns

- **[llm_batching](emerge/llm_batching)** - Reduce LLM API calls by 80% through natural batching
- **[distributed_swarm](emerge/distributed_swarm)** - Multi-region coordination without central control
- **[disruption_recovery](emerge/disruption_recovery)** - Self-healing and resilience patterns

### Advanced customization

- **[custom_decision](emerge/custom_decision)** - Implement custom decision strategies
- **[monitoring_metrics](emerge/monitoring_metrics)** - Real-time monitoring and analysis

## Which example should I start with?

| Your goal                  | Start with            | Then try            |
| -------------------------- | --------------------- | ------------------- |
| Understand core concepts   | `basic_sync`          | `energy_management` |
| Reduce API costs           | `llm_batching`        | `distributed_swarm` |
| Build resilient systems    | `disruption_recovery` | `custom_decision`   |
| Monitor production systems | `monitoring_metrics`  | `distributed_swarm` |

## Running examples

### Using task (recommended)

```bash
task run:example -- basic_sync
```

### Using go run

```bash
go run ./examples/emerge/basic_sync
```

### Run all examples

```bash
for example in basic_sync llm_batching distributed_swarm; do
    task run:example -- $example
    sleep 2
done
```

## Example details

See [emerge/README.md](emerge/README.md) for detailed descriptions of each example and the concepts they demonstrate.

