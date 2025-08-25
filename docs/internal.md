# Internal Packages

## Overview

This document describes the internal utility packages used across the bio-adapt system. These packages are internal and not part of the public API - they may change without notice.

## Package Structure

### analysis/

Diagnostic and quality analysis tools for monitoring system behavior.

- `diagnostics.go` - System diagnostics
- `quality.go` - Quality metrics

### config/

Configuration management and validation utilities.

- `agent.go` - Agent configuration
- `swarm.go` - Swarm configuration
- `validation.go` - Configuration validation

### display/

Display formatting and visualization utilities.

- `format.go` - Output formatting
- `phaseviz.go` - Phase visualization
- `progress.go` - Progress indicators
- `results.go` - Result formatting

### emerge/

Internal emerge-specific utilities.

- `basin.go` - Attractor basin calculations
- `convergence.go` - Convergence detection
- `metrics.go` - Metric collection
- `pattern.go` - Pattern recognition

### random/

Cryptographically secure random number generation.

- `random.go` - Secure random utilities

### resource/

Resource management primitives.

- `token.go` - Token-based resource control

### sim/

Simulation and mathematical utilities.

- `mathx.go` - Mathematical helper functions

### testutil/

Testing utilities and helpers.

- `emerge/testutil.go` - Emerge-specific test utilities

### topology/

Network topology builders for agent connections.

- `builder.go` - Topology builder interface
- `full.go` - Full mesh topology
- `ring.go` - Ring topology
- `star.go` - Star topology

## Usage

These packages are internal and should not be imported by external code. They are used internally by the public bio-adapt packages.

```go
// ❌ Don't do this in external code
import "github.com/carlisia/bio-adapt/internal/topology"

// ✅ Use public APIs instead
import "github.com/carlisia/bio-adapt/emerge/swarm"
```

## Testing

Internal packages have their own tests:

```bash
# Run all internal package tests
go test ./internal/...

# Run specific package tests
go test ./internal/topology
go test ./internal/analysis
```

## Contributing

When modifying internal packages:

1. Consider if the functionality should be public
2. Maintain compatibility with existing internal users
3. Add comprehensive tests
4. Document complex algorithms
5. Keep packages focused and cohesive

## Note

Internal packages are not covered by semantic versioning guarantees. They may change, move, or be removed in any release.
