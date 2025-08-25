# Documentation Index

## Overview

- [Architecture](architecture.md) - System design and principles
- [Primitives](primitives.md) - Overview of the three coordination primitives
- [Scale Definitions](emerge/scales.md) - Agent counts, performance, and resource requirements
- [Composition](composition.md) - How to combine primitives
- [Development](development.md) - Development guide and contribution guidelines
- [Deployment](deployment.md) - Production deployment guide
- [Internal Packages](internal.md) - Internal utilities documentation

## Core Concepts

- [Agents](concepts/agents.md) - The fundamental units of coordination
- [Swarm](concepts/swarm.md) - Collections of agents working toward goals
- [Synchronization](concepts/synchronization.md) - How agents coordinate behavior
- [Workload Integration](concepts/workload-integration.md) - How your application wraps emerge agents
- [Request Patterns](concepts/request-patterns.md) - Workload activity patterns (burst, steady, sparse, etc.)
- [Coherence](concepts/coherence.md) - Understanding synchronization measurement
- [Phase](concepts/phase.md) - Agent positions in oscillation cycles
- [Frequency](concepts/frequency.md) - Rate of phase change
- [Energy](concepts/energy.md) - Resource constraints and management
- [Goals](concepts/goals.md) - Optimization objectives (MinimizeAPICalls, DistributeLoad, etc.)
- [Strategies](concepts/strategies.md) - Synchronization approaches (PhaseNudge, PulseCoupling, etc.)

## Primitives

### Emerge (Production-Ready)

- [Use Cases](emerge/use-cases.md) - Real-world applications and examples
- [FAQ](emerge/faq.md) - Frequently asked questions
- [Algorithm](emerge/algorithm.md) - The emerge synchronization algorithm
- [Protocol](emerge/protocol.md) - The emerge-specific synchronization protocol
- [Goal-Directed](emerge/goal-directed.md) - How emerge pursues and maintains goals
- [Disruption](emerge/disruption.md) - How emerge handles failures and perturbations
- [Alternatives](emerge/alternatives.md) - Comparison with other coordination approaches
- [Decentralization](emerge/decentralization.md) - How emerge achieves coordination without central control
- [Concurrency](emerge/concurrency.md) - Go concurrency patterns and implementation
- [Security](emerge/security.md) - Security considerations and best practices
- [Primitive Guide](emerge/primitive.md) - High-level usage patterns
- [Architecture](emerge/architecture.md) - Detailed architecture
- [Optimization](emerge/optimization.md) - Performance optimizations
- [Package Reference](emerge/package.md) - Package structure and API
- [Scales](emerge/scales.md) - Agent count configurations
- [Glossary](emerge/glossary.md) - Terms and definitions

### Navigate (Coming Soon)

- [Primitive Guide](navigate/primitive.md) - Planned features
- [Architecture](navigate/architecture.md) - Design documentation
- [Optimization](navigate/optimization.md) - Planned optimizations

### Glue (Planned)

- [Primitive Guide](glue/primitive.md) - Concept overview
- [Architecture](glue/architecture.md) - Design documentation
- [Optimization](glue/optimization.md) - Planned optimizations

## Client Libraries

- [Overview](client/overview.md) - Client packages overview
- [Emerge Client](client/emerge.md) - Production-ready emerge client

## Simulations

- [Overview](simulations/overview.md) - Interactive demonstrations
- [Emerge Simulation](simulations/emerge.md) - Distributed workload optimization demo

## Testing

- [End-to-End Tests](testing/e2e.md) - E2E test documentation

## Quick Links

### Getting Started

1. Read the [Primitives Overview](primitives.md)
2. Understand [Core Concepts](concepts/goals.md)
3. Try the [Interactive Simulation](simulations/overview.md)
4. Use the [Emerge Client](client/emerge.md)

### For Contributors

1. [Development Guide](development.md)
2. [Architecture Overview](architecture.md)
3. [Internal Packages](internal.md)

### For Production

1. [Deployment Guide](deployment.md)
2. [Performance Optimization](emerge/optimization.md)
3. [E2E Testing](testing/e2e.md)
