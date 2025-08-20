# Neighbor Storage Optimization

## Overview

The bio-adapt library now includes optimized neighbor storage for large swarms (>100 agents), providing significant performance improvements for production deployments handling 1000+ concurrent agents.

## Implementation

### Adaptive Storage Selection

The system automatically selects the optimal storage strategy based on swarm size:

| Swarm Size | Storage Type | Implementation |
|------------|--------------|----------------|
| ≤100 agents | sync.Map | Standard concurrent map |
| >100 agents | Fixed arrays | Pre-allocated slices with atomic counters |

### OptimizedAgent Structure

```go
// emerge/agent/agent_optimized.go
type OptimizedAgent struct {
    *Agent // Embed standard agent
    optimizedNeighbors *NeighborStorage // Fixed-size array storage
    useOptimized       bool
}
```

### NeighborStorage Implementation

```go
// emerge/agent/neighbor_storage.go
type NeighborStorage struct {
    neighbors []*Agent      // Pre-allocated slice
    ids       []string      // ID lookup array
    count     int32         // Atomic counter
    capacity  int           // Fixed capacity
    mu        sync.RWMutex  // For thread safety
}
```

## Performance Improvements

### Benchmark Results

| Operation | Size | Standard (sync.Map) | Optimized (arrays) | Improvement |
|-----------|------|--------------------|--------------------|-------------|
| Neighbor iteration | 10 | 2,969 ns/op | 1,814 ns/op | **39% faster** |
| Neighbor iteration | 20 | 7,083 ns/op | 4,130 ns/op | **42% faster** |
| Neighbor iteration | 50 | 21,911 ns/op | 11,610 ns/op | **47% faster** |
| Coherence calc | 20 neighbors | 288 ns/op | 234 ns/op | **19% faster** |

### Memory Characteristics

- **Reduced allocations**: Fixed arrays eliminate dynamic allocation during iteration
- **Better cache locality**: Sequential memory access patterns
- **Lower GC pressure**: Pre-allocated storage reduces garbage collection overhead

## Usage

The optimization is completely transparent to users:

```go
// Automatically uses optimized storage for large swarms
swarm, err := swarm.New(1000, goalState)

// All API methods work identically
coherence := swarm.MeasureCoherence()
agent, found := swarm.Agent("agent-500")
```

## Technical Details

### Key Optimizations

1. **Direct slice iteration**: Eliminates sync.Map Range() overhead
2. **Pre-allocated arrays**: Reduces memory allocations
3. **Atomic operations**: Lock-free neighbor counting
4. **Batch operations**: Groups related updates for cache efficiency

### Compatibility

- Fully backward compatible with existing code
- Seamless interoperability between standard and optimized agents
- Automatic fallback to sync.Map for small swarms

## Testing

```bash
# Run benchmarks
go test -bench=BenchmarkNeighbor -benchmem ./emerge/agent

# Verify correctness
task test

# Check memory profile
go test -bench=. -memprofile=mem.prof ./emerge/agent
go tool pprof -http=:8080 mem.prof
```

## Production Considerations

1. **Neighbor capacity**: Default is 20 for small-world networks, adjustable via configuration
2. **Thread safety**: All operations are thread-safe with minimal locking
3. **Memory bounds**: Fixed capacity prevents unbounded growth
4. **Monitoring**: Performance metrics available through standard profiling tools

## Future Enhancements

- Dynamic resizing of neighbor arrays based on runtime patterns
- SIMD optimizations for coherence calculations
- Lock-free data structures for further concurrency improvements