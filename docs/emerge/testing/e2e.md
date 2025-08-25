# End-to-End Tests

## Overview

This directory contains end-to-end integration and performance tests for the bio-adapt system. These tests verify that the complete system works correctly under realistic conditions.

## Test Files

### performance_test.go

Benchmarks and performance tests to ensure the system meets performance requirements at various scales.

### swarm_convergence_test.go

Tests that verify swarms achieve target coherence levels across different configurations and scales.

### system_integration_test.go

Integration tests that verify the complete system works correctly when all components are used together.

## Running the Tests

```bash
# Run all e2e tests
task test:e2e

# Run with specific timeout
go test -timeout 10m ./e2e

# Run specific test
go test -run TestSwarmConvergence ./e2e

# Run benchmarks
go test -bench=. ./e2e
```

## Test Coverage

The e2e tests cover:

- **Convergence behavior** - Verifies swarms reach target coherence
- **Scale testing** - Tests with Tiny (20) through Huge (2000) agent counts
- **Performance validation** - Ensures operations complete within time limits
- **Resource usage** - Monitors memory and CPU consumption
- **Disruption recovery** - Tests self-healing capabilities
- **Goal achievement** - Verifies different optimization goals work correctly

## Performance Expectations

See [Scale Definitions](../emerge/scales.md) for detailed performance characteristics. The e2e tests verify these expectations are met.

## Writing New E2E Tests

When adding new e2e tests:

1. Use realistic configurations and scales
2. Set appropriate timeouts for convergence
3. Verify both functional correctness and performance
4. Clean up resources properly
5. Use subtests for different configurations

Example:

```go
func TestNewFeature(t *testing.T) {
    scales := []scale.Size{scale.Tiny, scale.Small, scale.Medium}

    for _, s := range scales {
        t.Run(s.String(), func(t *testing.T) {
            client := emerge.MinimizeAPICalls(s)

            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            err := client.Start(ctx)
            require.NoError(t, err)

            assert.True(t, client.IsConverged())
            assert.Greater(t, client.Coherence(), 0.8)
        })
    }
}
```
