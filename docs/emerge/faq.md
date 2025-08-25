# Frequently Asked Questions (FAQ)

## General Questions

### What is emerge?

Emerge is a decentralized synchronization algorithm that enables independent agents to coordinate their behavior without central control. Think of it like fireflies synchronizing their flashing - no firefly is in charge, but they all end up flashing together.

### How is emerge different from a scheduler?

**Schedulers** tell you exactly when to run (like cron: "run at 3pm").  
**Emerge** helps you figure out when to run together (like "let's all batch our API calls when we're ready").

Emerge is dynamic and adaptive, while schedulers are static and predetermined.

### Do I need to understand the math to use emerge?

No! The client API is simple:

```go
client := emerge.MinimizeAPICalls(scale.Medium)
client.Start(ctx)
```

The complex math is handled internally. You just specify what you want to achieve.

### Is emerge a consensus algorithm like Raft?

No. Consensus algorithms help you agree on a single value (like "who is the leader?"). Emerge helps you synchronize continuous behavior (like "when should we all act?").

- **Raft**: "Let's vote on value X"
- **Emerge**: "Let's coordinate our timing"

See [Alternatives](alternatives.md) for detailed comparisons.

## Getting Started

### How do I choose the right scale?

Start small and scale up:

- **Testing**: Use Tiny (20 agents)
- **Production start**: Use Small (50 agents) or Medium (200 agents)
- **Scale as needed**: Move to Large (1000) or Huge (2000+) when ready

See [Scales](scales.md) for detailed configurations and resource requirements.

```go
// Start small
client := emerge.MinimizeAPICalls(scale.Small)

// Scale up later
client := emerge.MinimizeAPICalls(scale.Large)
```

### How do I know which goal to use?

Ask yourself what you're trying to optimize:

- **Want to batch operations?** → `MinimizeAPICalls`
- **Want to spread load?** → `DistributeLoad`
- **Want agreement?** → `ReachConsensus`
- **Want speed?** → `MinimizeLatency`
- **Want to save resources?** → `SaveEnergy`

See [Goals](../concepts/goals.md) for detailed descriptions and [Use Cases](use_cases.md) for real-world examples.

### How long does synchronization take?

Convergence time depends on many factors:

- **Scale** - More agents generally take longer
- **Goal** - Different goals have different convergence characteristics
- **Network topology** - Full mesh converges faster than sparse networks
- **Initial conditions** - Random start vs partially synchronized
- **Parameters** - Coupling strength, update frequency, etc.

**Rough estimates** (actual times vary):

- **Tiny** (20 agents): Seconds
- **Small** (50 agents): Several seconds
- **Medium** (200 agents): Tens of seconds
- **Large** (1000 agents): Up to minutes
- **Huge** (2000+ agents): Several minutes

These are approximations. Always test with your specific configuration and workload to determine actual convergence times.

### Can I use emerge with my existing system?

Yes! Emerge is designed to integrate with existing systems without major rewrites. Here's how:

**Step 1: Add emerge client to your service**

```go
// In your existing service
type MyService struct {
    // Your existing fields stay the same
    database *sql.DB
    cache    *redis.Client

    // Add emerge client
    emergeClient *emerge.Client
}
```

**Step 2: Initialize emerge alongside your existing setup**

```go
func NewMyService() *MyService {
    s := &MyService{
        database: connectDB(),
        cache:    connectRedis(),
        // Add emerge with appropriate goal
        emergeClient: emerge.MinimizeAPICalls(scale.Small),
    }

    // Start emerge in background
    go s.emergeClient.Start(context.Background())
    return s
}
```

**Step 3: Use emerge to coordinate existing operations**

```go
type MyService struct {
    // ... existing fields ...
    emergeClient  *emerge.Client
    pendingItems  []Item // Add a field to accumulate items
    mu            sync.Mutex
}

// Your existing method now accumulates items
func (s *MyService) ProcessItem(item Item) {
    s.mu.Lock()
    s.pendingItems = append(s.pendingItems, item)
    s.mu.Unlock()
}

// Add a background goroutine to handle batching
func (s *MyService) RunBatchProcessor(ctx context.Context) {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    wasConverged := false

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            isConverged := s.emergeClient.IsConverged()

            // Batch when we transition into converged state
            // This prevents batching continuously while converged
            if isConverged && !wasConverged {
                s.mu.Lock()
                if len(s.pendingItems) > 0 {
                    // Now synchronized - batch all pending items
                    s.database.BatchInsert(s.pendingItems)
                    s.pendingItems = nil
                    // Items will accumulate again until next convergence
                }
                s.mu.Unlock()
            }

            wasConverged = isConverged
        }
    }
}
```

The key is that emerge doesn't replace your existing logic - it helps coordinate WHEN batched operations happen. Items accumulate until emerge says "now we're all synchronized, safe to batch!"

## Technical Questions

### How many agents can emerge handle?

Tested up to 2000 agents (Huge scale). Theoretically can handle more, but you'll need to:

- Increase memory (roughly 5KB per agent)
- Adjust parameters for larger scales
- Consider hierarchical organization for 10,000+ agents

### Does emerge require network communication?

No, emerge can work in-process with goroutines. For distributed systems, you'll need to implement neighbor communication, but emerge doesn't mandate any specific network protocol.

### What happens if agents fail?

Emerge is resilient to failures:

- Up to 50% of agents can fail without breaking synchronization
- Remaining agents continue coordinating
- No leader election or recovery protocol needed
- System self-heals as new agents join

### Is emerge deterministic?

The convergence is deterministic (will always reach the goal), but the exact path and time to convergence may vary based on initial conditions and random factors.

### Can agents have different capabilities?

While agents should follow the same protocol, they can have:

- Different natural frequencies
- Different stubbornness levels
- Different energy constraints
- Different workloads

The key is they all participate in the same synchronization protocol.

## Performance Questions

### Is emerge faster than using a message queue?

It's not about speed, it's about coordination:

- **Message queue**: Handles message delivery
- **Emerge**: Coordinates when to send messages

They solve different problems. You can use both together:

```go
if emerge.IsConverged() {
    queue.PublishBatch(messages)  // Use both!
}
```

### How much overhead does emerge add?

Minimal overhead:

- **Memory**: ~5KB per agent
- **CPU**: ~50ns per agent update (atomic operations)
- **Network**: Only neighbor communication (not all-to-all)

The coordination benefits usually outweigh the overhead.

### Does emerge scale linearly?

Better than linear for many operations:

- **Communication**: O(k) where k = neighbors (constant, not O(N))
- **Convergence time**: O(log N) in many cases
- **Memory**: O(N) - linear with agent count

### How do I optimize emerge performance?

1. **Choose the right scale** - Don't use Huge if Medium works
2. **Select appropriate goals** - Match goal to use case
3. **Use proper patterns** - High-frequency for batching, sparse for energy saving
4. **Monitor coherence** - Don't over-synchronize

## Troubleshooting

### Why aren't my agents synchronizing?

Common causes:

1. **Coupling too weak** - Increase coupling strength
2. **Network partitioned** - Check agent connectivity
3. **Energy depleted** - Increase recovery rate
4. **Wrong goal** - Verify goal matches your needs

### Why is coherence oscillating?

Usually means:

- Coupling strength too high (agents over-correcting)
- Energy depletion cycles (agents run out of energy)
- Conflicting influences (check network topology)

Solution: Reduce coupling strength or increase energy recovery.

### Why is convergence slow?

Could be:

- Scale too large for the goal
- Pattern doesn't match goal (e.g., sparse pattern with batching goal)
- Natural frequencies too diverse
- Network topology too sparse

### Can I force immediate synchronization?

No, and you shouldn't try. Emerge is about emergent coordination. Forcing it defeats the purpose and benefits. If you need immediate coordination, consider a different tool.

## Common Misconceptions

### "Emerge is just a fancy timer"

No. Timers are static and predetermined. Emerge dynamically adapts to system conditions, load, failures, and other factors. It's like comparing a sundial to a smart watch.

### "I can just use a database lock"

Database locks provide mutual exclusion (one at a time). Emerge provides coordination (all together). They solve different problems:

- **Lock**: "Only I can access this"
- **Emerge**: "Let's all act together"

### "This is too complex for my simple use case"

The complexity is hidden. Using emerge is as simple as:

```go
client := emerge.MinimizeAPICalls(scale.Small)
client.Start(ctx)
if client.IsConverged() {
    // Do your thing
}
```

That's simpler than implementing your own coordination logic.

### "Emerge requires all agents to be identical"

Agents need to follow the same synchronization protocol, but can have:

- Different workloads
- Different processing speeds
- Different resource constraints
- Different business logic

## Best Practices

### Should I use emerge for everything?

No. Use emerge when you need:

- Distributed coordination without central control
- Adaptive synchronization
- Resilience to failures
- Scalable coordination

Don't use emerge for:

- Simple task distribution (use work queues)
- Fixed scheduling (use cron)
- Mutual exclusion (use locks)
- Message passing (use message queues)

### How do I debug emerge?

1. **Monitor coherence** - Track synchronization level
2. **Check energy levels** - Ensure agents have resources
3. **Verify topology** - Confirm agents can see neighbors
4. **Use small scale** - Test with Tiny scale first
5. **Enable logging** - Add coherence/phase logging

### How do I test with emerge?

```go
func TestWithEmerge(t *testing.T) {
    // Use tiny scale for tests
    client := emerge.MinimizeAPICalls(scale.Tiny)

    // Use shorter timeouts
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    client.Start(ctx)

    // Wait for convergence
    require.Eventually(t, client.IsConverged, 5*time.Second, 100*time.Millisecond)
}
```

### Can I change goals dynamically?

Yes, but allow time for reconvergence:

```go
// Start with one goal
client := emerge.MinimizeAPICalls(scale.Medium)
client.Start(ctx)

// Later, switch goals
client.Stop()
client = emerge.DistributeLoad(scale.Medium)
client.Start(ctx)
```

## Getting Help

### Where can I find examples?

- **Simulation**: `simulations/emerge/` - Interactive demo
- **Tests**: `emerge/swarm/*_test.go` - Test cases
- **Documentation**: `docs/emerge/` - Detailed guides

### How do I report issues?

1. Check this FAQ first
2. Search existing issues on GitHub
3. Provide minimal reproduction case
4. Include coherence logs and agent counts

### Can emerge do X?

If X involves coordinating multiple independent entities without central control, probably yes! Emerge is quite flexible. Check if one of the existing goals matches your needs, or consider combining emerge with other tools.

### Is emerge production-ready?

Yes, the emerge primitive is production-ready. It has:

- Comprehensive test coverage
- Performance optimizations
- Proven algorithm (Kuramoto model)
- Resilience to failures

Always test with your specific use case and scale before production deployment.

## See Also

- [Getting Started](primitive.md) - Quick start guide
- [Algorithm](emerge_algorithm.md) - How emerge works
- [Goal-Directed](goal-directed.md) - How emerge pursues goals
- [Disruption](disruption.md) - Handling failures
- [Use Cases](use_cases.md) - Real-world applications
- [Glossary](glossary.md) - Term definitions
