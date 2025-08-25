# Concurrency Patterns in Emerge

## Overview

Emerge leverages Go's concurrency primitives to enable thousands of [agents](../concepts/agents.md) to synchronize efficiently. This document explains the concurrency patterns used, how they support the [emerge algorithm](emerge_algorithm.md), and the design decisions that enable scalable, lock-free [synchronization](../concepts/synchronization.md).

## Core Concurrency Model

### Agent Independence

Each [agent](../concepts/agents.md) operates as an independent entity:

- No shared mutable state between agents
- All inter-agent communication through atomic operations
- Agents can be updated in parallel without locks

```go
// Each agent is self-contained
type Agent struct {
    id    string
    state AtomicState  // Lock-free state
    // No direct references to other agents
}
```

### Swarm Orchestration

The [swarm](../concepts/swarm.md) coordinates agents without central locking (see [Decentralization](decentralization.md)):

```go
type Swarm struct {
    agents    []*Agent      // Independent agents
    topology  Topology      // Read-only after creation
    done      chan struct{} // Graceful shutdown
    // No mutex needed for normal operations
}
```

## Atomic Operations

### Lock-Free State Management

Agent state uses atomic operations for concurrent access:

```go
type AtomicState struct {
    phase     atomic.Uint64  // Phase as fixed-point integer
    frequency atomic.Uint64  // Frequency as fixed-point
    energy    atomic.Uint64  // Energy as fixed-point
}

// Lock-free phase update
func (s *AtomicState) UpdatePhase(delta float64) {
    for {
        oldBits := s.phase.Load()
        oldPhase := math.Float64frombits(oldBits)
        newPhase := normalizePhase(oldPhase + delta)
        newBits := math.Float64bits(newPhase)

        if s.phase.CompareAndSwap(oldBits, newBits) {
            break
        }
        // Retry on conflict
    }
}
```

### Why Atomics?

1. **No Lock Contention**: Agents never wait for locks
2. **Cache-Friendly**: Each agent's state fits in cache lines
3. **Scalability**: Performance doesn't degrade with agent count
4. **Predictable Latency**: No unpredictable lock wait times

## Parallel Update Pattern

### Concurrent Agent Updates

Agents update in parallel using worker pools:

```go
func (s *Swarm) UpdateAgents(ctx context.Context) {
    numWorkers := runtime.NumCPU()
    chunkSize := len(s.agents) / numWorkers

    var wg sync.WaitGroup
    wg.Add(numWorkers)

    for i := 0; i < numWorkers; i++ {
        start := i * chunkSize
        end := start + chunkSize
        if i == numWorkers-1 {
            end = len(s.agents)
        }

        go func(agents []*Agent) {
            defer wg.Done()
            for _, agent := range agents {
                agent.Update(ctx)
            }
        }(s.agents[start:end])
    }

    wg.Wait()
}
```

### Work Stealing for Load Balance

For uneven workloads, use work stealing:

```go
type WorkQueue struct {
    agents chan *Agent
    done   chan struct{}
}

func (s *Swarm) UpdateWithWorkStealing(ctx context.Context) {
    queue := &WorkQueue{
        agents: make(chan *Agent, len(s.agents)),
        done:   make(chan struct{}),
    }

    // Fill queue
    for _, agent := range s.agents {
        queue.agents <- agent
    }
    close(queue.agents)

    // Workers steal from queue
    var wg sync.WaitGroup
    for i := 0; i < runtime.NumCPU(); i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for agent := range queue.agents {
                agent.Update(ctx)
            }
        }()
    }

    wg.Wait()
}
```

## Channel Patterns

### Event Broadcasting

Swarm events use fan-out pattern:

```go
type EventBroadcaster struct {
    listeners []chan Event
    mu        sync.RWMutex
}

func (b *EventBroadcaster) Broadcast(event Event) {
    b.mu.RLock()
    defer b.mu.RUnlock()

    for _, listener := range b.listeners {
        select {
        case listener <- event:
        default:
            // Non-blocking send, drop if full
        }
    }
}
```

### Graceful Shutdown

Coordinated shutdown using channels:

```go
func (s *Swarm) Run(ctx context.Context) error {
    ticker := time.NewTicker(s.updateInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return s.shutdown()

        case <-ticker.C:
            s.UpdateAgents(ctx)

        case cmd := <-s.commands:
            s.handleCommand(cmd)
        }
    }
}

func (s *Swarm) shutdown() error {
    // Signal all goroutines
    close(s.done)

    // Wait for graceful termination
    s.wg.Wait()

    return nil
}
```

## Memory Management

### Object Pooling

Reduce GC pressure with sync.Pool:

```go
var statePool = sync.Pool{
    New: func() interface{} {
        return &AgentState{
            Neighbors: make([]int32, 0, 10),
        }
    },
}

func (a *Agent) Update() {
    // Get temporary state from pool
    state := statePool.Get().(*AgentState)
    defer statePool.Put(state)

    // Use state for calculations
    state.Reset()
    state.CalculatePhaseAdjustment(a)

    // Apply results atomically
    a.ApplyAdjustment(state.adjustment)
}
```

### Neighbor Storage Optimization

Compact neighbor storage for cache efficiency:

```go
type OptimizedNeighbors struct {
    // Use int32 indices instead of pointers
    indices []int32  // 4 bytes per neighbor vs 8 bytes for pointer

    // Pre-allocated buffer pool
    bufferPool *sync.Pool
}

func (n *OptimizedNeighbors) GetNeighborStates(agents []*Agent) []float64 {
    // Get buffer from pool
    buf := n.bufferPool.Get().([]float64)
    defer n.bufferPool.Put(buf)

    // Gather neighbor phases efficiently
    for i, idx := range n.indices {
        buf[i] = agents[idx].GetPhase()
    }

    return buf[:len(n.indices)]
}
```

## Synchronization Primitives

### Read-Copy-Update (RCU) Pattern

For rarely-changing topology:

```go
type Topology struct {
    connections atomic.Value  // *ConnectionMap
}

func (t *Topology) Update(newConnections *ConnectionMap) {
    // Atomic swap - readers see old or new, never partial
    t.connections.Store(newConnections)
}

func (t *Topology) GetNeighbors(agentID string) []string {
    // Lock-free read
    connections := t.connections.Load().(*ConnectionMap)
    return connections.GetNeighbors(agentID)
}
```

### Barrier Synchronization

For phase-locked updates:

```go
type PhaseBarrier struct {
    count  atomic.Int32
    target int32
    done   chan struct{}
}

func (b *PhaseBarrier) Wait() {
    if b.count.Add(1) == b.target {
        close(b.done)  // Release all waiters
    } else {
        <-b.done  // Wait for last agent
    }
}

// Usage in synchronized updates
func (s *Swarm) SynchronizedUpdate() {
    barrier := &PhaseBarrier{
        target: int32(len(s.agents)),
        done:   make(chan struct{}),
    }

    for _, agent := range s.agents {
        go func(a *Agent) {
            a.CalculateUpdate()
            barrier.Wait()  // Wait for all to calculate
            a.ApplyUpdate() // All apply together
        }(agent)
    }
}
```

## Context Propagation

### Cancellation and Timeouts

Proper context handling throughout:

```go
func (a *Agent) UpdateWithTimeout(ctx context.Context) error {
    // Create timeout for this update
    updateCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
    defer cancel()

    // Check context before expensive operations
    select {
    case <-updateCtx.Done():
        return updateCtx.Err()
    default:
    }

    // Perform update
    return a.performUpdate(updateCtx)
}
```

### Context Values for Debugging

Pass trace information through context:

```go
type traceKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceKey{}, traceID)
}

func (a *Agent) Update(ctx context.Context) {
    if traceID, ok := ctx.Value(traceKey{}).(string); ok {
        // Log with trace ID for debugging
        log.Printf("[%s] Agent %s updating", traceID, a.id)
    }
    // ... update logic
}
```

## Performance Patterns

### Batching for Efficiency

Batch operations to reduce overhead:

```go
type UpdateBatcher struct {
    updates chan Update
    batch   []Update
    ticker  *time.Ticker
}

func (b *UpdateBatcher) Run(ctx context.Context) {
    for {
        select {
        case update := <-b.updates:
            b.batch = append(b.batch, update)

        case <-b.ticker.C:
            if len(b.batch) > 0 {
                b.processBatch(b.batch)
                b.batch = b.batch[:0]  // Reuse slice
            }

        case <-ctx.Done():
            return
        }
    }
}
```

### CPU Cache Optimization

Align data structures for cache efficiency:

```go
type CacheAlignedAgent struct {
    _ [0]func() // Prevents false sharing

    // Frequently accessed together
    phase     atomic.Uint64
    frequency atomic.Uint64

    // Padding to next cache line
    _ [40]byte

    // Less frequently accessed
    energy    atomic.Uint64
    neighbors []int32
}
```

## Testing Concurrent Code

### Deterministic Testing

Make concurrency deterministic for tests:

```go
type TestScheduler struct {
    agents []*Agent
    order  []int  // Deterministic update order
}

func (s *TestScheduler) UpdateDeterministic() {
    for _, idx := range s.order {
        s.agents[idx].Update()
    }
}

func TestConvergence(t *testing.T) {
    scheduler := &TestScheduler{
        agents: createAgents(100),
        order:  []int{0, 1, 2, ...},  // Fixed order
    }

    // Test with deterministic scheduling
    scheduler.UpdateDeterministic()

    // Verify convergence properties
    assert.True(t, checkConvergence(scheduler.agents))
}
```

### Race Detection

Use Go's race detector:

```bash
# Run tests with race detection
go test -race ./emerge/...

# Run simulation with race detection
go run -race ./simulations/emerge
```

### Stress Testing

Test under high concurrency:

```go
func TestHighConcurrency(t *testing.T) {
    swarm := NewSwarm(1000)

    // Create high contention
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 1000; j++ {
                swarm.Update()
            }
        }()
    }

    wg.Wait()

    // Verify consistency
    assert.True(t, swarm.IsConsistent())
}
```

## Common Pitfalls

### 1. False Sharing

**Problem**: Multiple goroutines updating nearby memory

```go
// Bad: agents in same cache line
type BadLayout struct {
    agent1Phase atomic.Uint64
    agent2Phase atomic.Uint64  // False sharing!
}
```

**Solution**: Add padding or restructure

```go
// Good: separate cache lines
type GoodLayout struct {
    agent1Phase atomic.Uint64
    _           [56]byte  // Padding to 64 bytes
    agent2Phase atomic.Uint64
}
```

### 2. Goroutine Leaks

**Problem**: Goroutines not terminating

```go
// Bad: goroutine leaks if ctx never cancels
go func() {
    for {
        time.Sleep(1 * time.Second)
        update()
    }
}()
```

**Solution**: Always handle context

```go
// Good: goroutine exits on context cancel
go func() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            update()
        case <-ctx.Done():
            return
        }
    }
}()
```

### 3. Deadlocks

**Problem**: Circular channel dependencies

```go
// Bad: potential deadlock
func bad() {
    ch1 := make(chan int)
    ch2 := make(chan int)

    go func() {
        ch1 <- <-ch2  // Waits for ch2
    }()

    ch2 <- <-ch1  // Waits for ch1 - DEADLOCK!
}
```

**Solution**: Use buffered channels or timeouts

```go
// Good: buffered channels prevent deadlock
func good() {
    ch1 := make(chan int, 1)
    ch2 := make(chan int, 1)

    // ... same logic works now
}
```

## Best Practices

### 1. Start Simple

- Begin with mutexes, optimize to atomics if needed
- Profile before optimizing
- Measure impact of concurrency changes

### 2. Design for Concurrency

- Minimize shared state
- Make data structures immutable where possible
- Use message passing over shared memory

### 3. Test Thoroughly

- Always use race detector during development
- Test with varying GOMAXPROCS
- Stress test with high agent counts

### 4. Monitor in Production

- Track goroutine counts
- Monitor channel buffer usage
- Watch for lock contention

## Performance Metrics

### Concurrency Overhead

Measured overhead for different patterns:

| Pattern       | Overhead per Agent | Scalability              |
| ------------- | ------------------ | ------------------------ |
| Mutex-based   | ~500ns             | Poor (lock contention)   |
| Atomic-based  | ~50ns              | Excellent (lock-free)    |
| Channel-based | ~200ns             | Good (depends on buffer) |
| RCU pattern   | ~20ns read         | Excellent for read-heavy |

### Optimal Concurrency Levels

| Agent Count | Optimal Workers | Update Time |
| ----------- | --------------- | ----------- |
| 10-100      | 2-4             | <1ms        |
| 100-1000    | 4-8             | <10ms       |
| 1000-10000  | 8-16            | <100ms      |
| 10000+      | 16-32           | <1s         |

See [Scales](scales.md) for standard agent count configurations.

## See Also

### Core Documentation
- [Algorithm](emerge_algorithm.md) - How emerge algorithm works
- [Architecture](architecture.md) - System design
- [Protocol](protocol.md) - Synchronization protocol
- [Optimization](optimization.md) - Performance tuning
- [Decentralization](decentralization.md) - No central control

### Concepts
- [Agents](../concepts/agents.md) - Fundamental units
- [Swarm](../concepts/swarm.md) - Agent collections
- [Synchronization](../concepts/synchronization.md) - Coordination

### Related Topics
- [Disruption](disruption.md) - Handling failures
- [Security](security.md) - Security considerations
- [Scales](scales.md) - Configuration sizes
- [Testing](../testing/e2e.md) - Testing strategies
