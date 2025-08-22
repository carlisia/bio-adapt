# Deployment guide

## Overview

This guide covers deploying bio-adapt in production environments, including configuration, monitoring, scaling, and troubleshooting.

## System requirements

### Minimum requirements

- Go 1.21 or higher
- 2 CPU cores
- 1GB RAM (for swarms up to 1000 agents)
- Linux, macOS, or Windows

### Recommended for production

- Go 1.22+ (latest stable)
- 4+ CPU cores (more cores improve concurrent agent updates)
- 4GB+ RAM (for swarms of 5000+ agents)
- Linux with kernel 5.x+ (better scheduling)

### Resource planning

| Swarm size | Memory | CPU cores | Network   |
| ---------- | ------ | --------- | --------- |
| 10-100     | 256MB  | 2         | Minimal   |
| 100-1000   | 1GB    | 4         | Moderate  |
| 1000-5000  | 4GB    | 8         | High      |
| 5000+      | 8GB+   | 16+       | Very high |

## Configuration

### Basic configuration

```go
package main

import (
    "context"
    "time"
    "github.com/carlisia/bio-adapt/emerge"
)

func main() {
    // Production configuration
    config := emerge.SwarmConfig{
        Size: 1000,
        Goal: emerge.State{
            Frequency: 100 * time.Millisecond,
            Coherence: 0.9,
        },
        // Production timeouts
        ConvergenceTimeout: 30 * time.Second,
        UpdateInterval:     10 * time.Millisecond,
    }

    swarm, err := emerge.NewSwarmWithConfig(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    swarm.Run(ctx)
}
```

### Advanced configuration

```go
// Resource limits
const (
    maxSwarmSize     = 5000
    maxAgentEnergy   = 1000
    minEnergyForAction = 10
    energyRecoveryRate = 5.0
)

// Create production swarm with limits
func CreateProductionSwarm(size int) (*emerge.Swarm, error) {
    // Validate size
    if size > maxSwarmSize {
        return nil, fmt.Errorf("swarm size %d exceeds limit", size)
    }

    // Configure agents
    agentConfig := emerge.AgentConfig{
        InitialEnergy:      maxAgentEnergy,
        MinEnergyThreshold: minEnergyForAction,
        RecoveryRate:      energyRecoveryRate,
        Stubbornness:      0.3, // Moderate resistance
        Influence:         0.7, // Strong influence
    }

    // Create swarm
    swarm, err := emerge.NewSwarm(size, goal)
    if err != nil {
        return nil, err
    }

    // Apply agent configuration
    swarm.ConfigureAgents(agentConfig)

    return swarm, nil
}
```

## Deployment patterns

### Standalone service

```go
// main.go - Standalone synchronization service
package main

import (
    "context"
    "net/http"
    "github.com/carlisia/bio-adapt/emerge"
)

func main() {
    // Create swarm
    swarm, _ := emerge.NewSwarm(100, goalState)

    // Run in background
    ctx := context.Background()
    go swarm.Run(ctx)

    // Expose metrics endpoint
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := swarm.GetMetrics()
        json.NewEncoder(w).Encode(metrics)
    })

    http.ListenAndServe(":8080", nil)
}
```

### Embedded in application

```go
// Embed bio-adapt in existing service
type Service struct {
    swarm *emerge.Swarm
    // ... other fields
}

func (s *Service) Initialize() error {
    // Create swarm for request batching
    swarm, err := emerge.NewSwarm(50, emerge.State{
        Frequency: 200 * time.Millisecond, // Batch window
        Coherence: 0.9,
    })
    if err != nil {
        return err
    }

    s.swarm = swarm
    go s.swarm.Run(context.Background())

    return nil
}

func (s *Service) ProcessRequest(req Request) {
    // Use swarm synchronization for batching
    agent := s.swarm.Agent(req.ID)
    if agent.Phase() < 0.1 { // In batch window
        s.batchQueue.Add(req)
    }
}
```

### Kubernetes deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bio-adapt-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bio-adapt
  template:
    metadata:
      labels:
        app: bio-adapt
    spec:
      containers:
        - name: bio-adapt
          image: your-registry/bio-adapt:latest
          resources:
            requests:
              memory: "1Gi"
              cpu: "2"
            limits:
              memory: "2Gi"
              cpu: "4"
          env:
            - name: SWARM_SIZE
              value: "1000"
            - name: TARGET_COHERENCE
              value: "0.9"
            - name: CONVERGENCE_TIMEOUT
              value: "30s"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

## Monitoring

### Health checks

```go
// Health check endpoint
func (s *Service) HealthCheck() HealthStatus {
    coherence := s.swarm.MeasureCoherence()
    activeAgents := s.swarm.ActiveAgentCount()

    status := HealthStatus{
        Healthy:      coherence > 0.6,
        Coherence:    coherence,
        ActiveAgents: activeAgents,
        Uptime:       time.Since(s.startTime),
    }

    if coherence < 0.3 {
        status.Status = "DEGRADED"
        status.Message = "Low coherence detected"
    }

    return status
}
```

### Metrics collection

```go
// Prometheus metrics (future feature)
var (
    swarmCoherence = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "bioadapt_swarm_coherence",
            Help: "Current swarm coherence level",
        },
        []string{"swarm_id"},
    )

    convergenceTime = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "bioadapt_convergence_duration_seconds",
            Help: "Time to reach target coherence",
        },
        []string{"swarm_id"},
    )
)

// Collect metrics
func (s *Service) CollectMetrics() {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        for range ticker.C {
            coherence := s.swarm.MeasureCoherence()
            swarmCoherence.WithLabelValues(s.swarmID).Set(coherence)
        }
    }()
}
```

### Logging

```go
import "log/slog"

// Structured logging
func (s *Service) logSwarmState() {
    slog.Info("swarm state",
        "coherence", s.swarm.MeasureCoherence(),
        "agents", s.swarm.Size(),
        "convergence_time", s.convergenceTime,
        "disruptions", s.disruptionCount,
    )
}
```

## Performance tuning

### CPU optimization

```go
import "runtime"

// Optimize for CPU cores
func OptimizeForCPU() {
    // Use all available cores
    runtime.GOMAXPROCS(runtime.NumCPU())

    // Adjust worker pool based on cores
    workers := runtime.NumCPU() * 2
    if swarmSize > 1000 {
        workers = runtime.NumCPU() * 4
    }
}
```

### Memory optimization

```go
// Pre-allocate for large swarms
func CreateOptimizedSwarm(size int) *emerge.Swarm {
    if size > 1000 {
        // Pre-allocate agent storage
        emerge.PreallocateAgentPool(size)
    }

    swarm, _ := emerge.NewSwarm(size, goal)
    return swarm
}
```

### Network optimization

```go
// Batch network updates
type BatchedUpdater struct {
    updates chan Update
    batch   []Update
    ticker  *time.Ticker
}

func (b *BatchedUpdater) Run() {
    for {
        select {
        case update := <-b.updates:
            b.batch = append(b.batch, update)
        case <-b.ticker.C:
            if len(b.batch) > 0 {
                b.sendBatch(b.batch)
                b.batch = b.batch[:0]
            }
        }
    }
}
```

## Scaling strategies

### Horizontal scaling

```go
// Multi-swarm coordination
type SwarmCluster struct {
    swarms []*emerge.Swarm
    router LoadBalancer
}

func (sc *SwarmCluster) AddSwarm() {
    swarm, _ := emerge.NewSwarm(100, sc.goal)
    sc.swarms = append(sc.swarms, swarm)
    go swarm.Run(context.Background())
}

func (sc *SwarmCluster) Route(agentID string) *emerge.Swarm {
    // Consistent hashing for agent assignment
    return sc.router.SelectSwarm(agentID)
}
```

### Vertical scaling

```go
// Dynamic swarm resizing
func (s *Service) AutoScale() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        load := s.getCurrentLoad()

        if load > 0.8 && s.swarm.Size() < maxSize {
            // Add agents
            s.swarm.AddAgents(10)
        } else if load < 0.3 && s.swarm.Size() > minSize {
            // Remove agents
            s.swarm.RemoveAgents(10)
        }
    }
}
```

## Troubleshooting

### Common issues

**Low coherence**

- Check energy levels - agents may be depleted
- Verify network connectivity between agents
- Review stubbornness settings - too high prevents convergence
- Check for disruptions or failing agents

**High memory usage**

- Verify swarm size is within limits
- Check for memory leaks in custom strategies
- Review neighbor storage settings
- Enable memory profiling

**Slow convergence**

- Increase coupling strength
- Reduce stubbornness
- Check network latency
- Verify CPU resources

### Debug logging

```go
// Enable debug mode
func EnableDebug() {
    slog.SetLogLoggerLevel(slog.LevelDebug)

    // Log agent decisions
    emerge.SetDebugCallback(func(event DebugEvent) {
        slog.Debug("agent event",
            "agent_id", event.AgentID,
            "action", event.Action,
            "phase", event.Phase,
            "energy", event.Energy,
        )
    })
}
```

### Profiling

```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine analysis
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

## Security considerations

### Resource limits

```go
// Enforce resource limits
type LimitedSwarm struct {
    *emerge.Swarm
    maxMemory   int64
    maxCPU      float64
    rateLimit   rate.Limiter
}

func (ls *LimitedSwarm) AddAgent(id string) error {
    // Check rate limit
    if !ls.rateLimit.Allow() {
        return errors.New("rate limit exceeded")
    }

    // Check memory
    if getCurrentMemory() > ls.maxMemory {
        return errors.New("memory limit exceeded")
    }

    return ls.Swarm.AddAgent(id)
}
```

### Input validation

```go
// Validate configuration
func ValidateConfig(config SwarmConfig) error {
    if config.Size > 10000 {
        return errors.New("swarm size too large")
    }

    if config.Coherence < 0 || config.Coherence > 1 {
        return errors.New("invalid coherence value")
    }

    if config.Frequency < time.Millisecond {
        return errors.New("frequency too high")
    }

    return nil
}
```

## Best practices

1. **Start small**: Begin with small swarms and scale up
2. **Monitor continuously**: Track coherence and resource usage
3. **Set timeouts**: Always use context timeouts
4. **Handle failures**: Implement graceful degradation
5. **Test disruptions**: Regularly test failure scenarios
6. **Document thresholds**: Record optimal settings for your use case
7. **Version carefully**: Test thoroughly before upgrading

