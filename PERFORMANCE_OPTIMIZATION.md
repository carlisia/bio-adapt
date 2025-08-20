# Performance Architecture

## Storage Strategy

The swarm automatically selects storage based on size:

| Swarm Size | Storage Type | Rationale |
|------------|-------------|-----------|
| ≤100 agents | sync.Map | Simple, good concurrency for small sets |
| >100 agents | Slice + index map | Better cache locality, faster iteration |

## Memory Characteristics

### Baseline Performance
- Small swarm (10 agents): ~58KB, 485 allocations
- Medium swarm (100 agents): ~344KB, 6,496 allocations
- Large swarm (500 agents): ~2.8MB, 56,576 allocations
- Very large swarm (1000 agents): ~3.3MB, 69,535 allocations

Memory scales linearly with swarm size, with optimizations reducing allocations in hot paths.

## Implementation Details

### Object Pooling (`emerge/agent/pool.go`)

```go
type PoolManager struct {
    pool  sync.Pool
    stats *PoolStats
}

// Get retrieves an Agent from the pool or creates a new one
func (p *PoolManager) Get(id string) *Agent
// Put returns an Agent to the pool for reuse
func (p *PoolManager) Put(a *Agent)
```

### Adaptive Storage (`emerge/swarm/swarm.go`)

```go
type Swarm struct {
    // Storage - automatically selected based on size
    agents      sync.Map       // map[string]*Agent - used for small swarms
    agentSlice  []*agent.Agent // Direct access by index - used for large swarms
    agentIndex  map[string]int // ID to index mapping - used for large swarms
    optimized   bool           // Whether using optimized storage
    
    // Performance optimization for large swarms
    workerPool *WorkerPool // Goroutine pool for concurrent updates
}
```

### Worker Pool (`emerge/swarm/swarm.go`)

```go
type WorkerPool struct {
    workers   int
    workQueue chan func()
    quit      chan struct{}
}

// Automatically sized based on swarm size:
// <100 agents: NumCPU
// <1000 agents: NumCPU * 2  
// ≥1000 agents: min(NumCPU * 4, 32)
```

### Optimized Methods

```go
// MeasureCoherence uses optimized path for large swarms
func (s *Swarm) MeasureCoherence() float64 {
    if s.optimized {
        // Direct slice iteration with pre-allocated array
        phases := make([]float64, len(s.agentSlice))
        for i, a := range s.agentSlice {
            phases[i] = a.Phase()
        }
        return core.MeasureCoherence(phases)
    }
    // Standard sync.Map iteration for small swarms
}

// Agent lookup is O(1) for optimized swarms
func (s *Swarm) Agent(id string) (*agent.Agent, bool) {
    if s.optimized {
        if idx, ok := s.agentIndex[id]; ok {
            return s.agentSlice[idx], true
        }
    }
    // Standard sync.Map lookup for small swarms
}
```

## Benchmarking

```bash
# Full benchmark suite
go test -bench=. -benchmem ./emerge/swarm

# Specific size comparison
go test -bench=BenchmarkLargeSwarm -benchmem ./emerge/swarm

# Memory profiling
go test -bench=. -memprofile=mem.prof ./emerge/swarm
go tool pprof -http=:8080 mem.prof
```

