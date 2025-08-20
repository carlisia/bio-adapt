# Performance Optimization Plan

## Current State Analysis

### Memory Usage (from benchmarks)

- Small swarm (10 agents): ~58KB, 485 allocations
- Medium swarm (100 agents): ~344KB, 6,496 allocations
- Large swarm (500 agents): ~2.8MB, 56,576 allocations
- Very large swarm (1000 agents): ~3.3MB, 69,535 allocations

### Key Observations

1. Memory scales roughly linearly with swarm size (good)
2. High allocation count suggests many small objects
3. sync.Map usage in both Agent and Swarm may cause overhead
4. atomic.Value and atomic.Float64 have memory alignment requirements

## Optimization Strategies

### Phase 1: Reduce Allocations (Quick Wins)

#### 1.1 Pool Agent Objects

```go
var agentPool = sync.Pool{
    New: func() interface{} {
        return &Agent{}
    },
}

func GetAgent() *Agent {
    return agentPool.Get().(*Agent)
}

func PutAgent(a *Agent) {
    a.Reset() // Clear agent state
    agentPool.Put(a)
}
```

#### 1.2 Pre-allocate Neighbor Maps

```go
// Instead of sync.Map, use pre-sized regular map with RWMutex
type Agent struct {
    neighbors    map[string]*Agent
    neighborsMux sync.RWMutex
}

// Pre-allocate based on expected neighbors
func NewAgent(id string, expectedNeighbors int) *Agent {
    return &Agent{
        neighbors: make(map[string]*Agent, expectedNeighbors),
    }
}
```

#### 1.3 Batch Operations

```go
// Instead of individual updates
func (s *Swarm) BatchUpdatePhases(updates map[string]float64) {
    for id, phase := range updates {
        if agent, ok := s.agents.Load(id); ok {
            agent.(*Agent).SetPhase(phase)
        }
    }
}
```

### Phase 2: Optimize Data Structures

#### 2.1 Replace sync.Map for Known Size Collections

```go
// For swarm with known size, use slice + index map
type Swarm struct {
    agents      []*Agent           // Direct access by index
    agentIndex  map[string]int     // ID to index mapping
    agentsMux   sync.RWMutex
}
```

#### 2.2 Reduce Atomic Operations

```go
// Group related fields to reduce cache line bouncing
type AgentState struct {
    phase        float64
    frequency    time.Duration
    energy       float64
}

type Agent struct {
    state atomic.Value // Store entire state atomically
}
```

#### 2.3 Use Value Types Where Possible

```go
// Instead of pointers for small structs
type Context struct {
    LocalCoherence  float64
    NeighborCount   int
    AveragePhase    float64
}
// Pass by value instead of *Context
```

### Phase 3: Goroutine Management

#### 3.1 Worker Pool Pattern

```go
type WorkerPool struct {
    workers   int
    workChan  chan func()
    wg        sync.WaitGroup
}

func (wp *WorkerPool) Submit(work func()) {
    wp.workChan <- work
}

// Use for agent updates
pool := NewWorkerPool(runtime.NumCPU())
for _, agent := range agents {
    pool.Submit(func() {
        agent.Update(ctx)
    })
}
```

#### 3.2 Batch Processing

```go
// Process agents in batches to reduce goroutine overhead
func (s *Swarm) UpdateInBatches(batchSize int) {
    agents := s.GetAgents()
    for i := 0; i < len(agents); i += batchSize {
        end := min(i+batchSize, len(agents))
        batch := agents[i:end]

        var wg sync.WaitGroup
        wg.Add(1)
        go func(batch []*Agent) {
            defer wg.Done()
            for _, a := range batch {
                a.Update()
            }
        }(batch)
        wg.Wait()
    }
}
```

### Phase 4: Memory-Aware Algorithms

#### 4.1 Lazy Evaluation

```go
// Calculate expensive metrics only when needed
type Agent struct {
    coherenceCache     float64
    coherenceCacheTime time.Time
}

func (a *Agent) GetCoherence() float64 {
    if time.Since(a.coherenceCacheTime) > 100*time.Millisecond {
        a.coherenceCache = a.calculateCoherence()
        a.coherenceCacheTime = time.Now()
    }
    return a.coherenceCache
}
```

#### 4.2 Streaming Calculations

```go
// Instead of storing all history
type RunningAverage struct {
    sum   float64
    count int64
}

func (ra *RunningAverage) Add(value float64) {
    ra.sum += value
    ra.count++
}

func (ra *RunningAverage) Average() float64 {
    if ra.count == 0 {
        return 0
    }
    return ra.sum / float64(ra.count)
}
```

### Phase 5: Monitoring and Limits

#### 5.1 Memory Limits

```go
func (s *Swarm) CheckMemoryUsage() error {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    limitMB := float64(s.size) * 0.5 // 0.5 MB per agent
    currentMB := float64(m.Alloc) / (1024 * 1024)

    if currentMB > limitMB {
        return fmt.Errorf("memory limit exceeded: %.2f MB > %.2f MB",
            currentMB, limitMB)
    }
    return nil
}
```

#### 5.2 Metrics Collection

```go
type PerformanceMetrics struct {
    AllocatedBytes   uint64
    NumGoroutines    int
    NumAgents        int
    UpdatesPerSecond float64
}

func (s *Swarm) GetMetrics() PerformanceMetrics {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return PerformanceMetrics{
        AllocatedBytes:   m.Alloc,
        NumGoroutines:    runtime.NumGoroutine(),
        NumAgents:        s.size,
        UpdatesPerSecond: s.calculateUpdateRate(),
    }
}
```

## Implementation Priority

1. **Week 1**: Implement object pooling and pre-allocation (Phase 1)
   - Expected improvement: 30-40% reduction in allocations
2. **Week 2**: Optimize data structures (Phase 2)
   - Expected improvement: 20-30% memory reduction
3. **Week 3**: Implement goroutine management (Phase 3)
   - Expected improvement: Better scaling for large swarms
4. **Week 4**: Add monitoring and testing (Phase 5)
   - Ensure optimizations don't break functionality

## Benchmarking Strategy

### Before Each Change

```bash
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./emerge/swarm
go tool pprof -http=:8080 mem.prof
```

### Key Metrics to Track

- Allocations per operation
- Memory usage per agent
- Goroutine count
- Time to convergence
- CPU usage patterns

## Testing Plan

1. Create benchmark suite comparing before/after
2. Add memory leak detection tests
3. Stress test with very large swarms (10,000+ agents)
4. Long-running stability tests (24+ hours)
5. Profile under different workloads

## Success Criteria

- [ ] Reduce memory usage by 50% for large swarms
- [ ] Reduce allocations by 60%
- [ ] Support 10,000+ agent swarms
- [ ] No memory leaks in 24-hour test
- [ ] Maintain or improve convergence speed

