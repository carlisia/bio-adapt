# Scale Definitions

Bio-adapt uses predefined scales to simplify configuration and ensure optimal performance at different sizes.

## Available Scales

| Scale      | Agents | Memory/Agent | Convergence Time | Use Case                |
| ---------- | ------ | ------------ | ---------------- | ----------------------- |
| **Tiny**   | 20     | ~5KB         | ~1 second        | Testing, quick demos    |
| **Small**  | 50     | ~4KB         | ~2 seconds       | Team-sized coordination |
| **Medium** | 200    | ~3KB         | ~5 seconds       | Department scale        |
| **Large**  | 1000   | ~3KB         | ~15 seconds      | Enterprise deployments  |
| **Huge**   | 2000   | ~2KB         | ~30 seconds      | Cloud-scale operations  |

## Resource Requirements

| Scale  | Min Memory | Recommended Memory | CPU Cores | Network   |
| ------ | ---------- | ------------------ | --------- | --------- |
| Tiny   | 256MB      | 512MB              | 2         | Minimal   |
| Small  | 512MB      | 1GB                | 2         | Minimal   |
| Medium | 1GB        | 2GB                | 4         | Moderate  |
| Large  | 2GB        | 4GB                | 8         | High      |
| Huge   | 4GB        | 8GB                | 16        | Very high |

## Performance Characteristics

### Optimization Thresholds

- **≤100 agents**: Uses sync.Map for flexible neighbor storage
- **>100 agents**: Switches to fixed-size arrays for better cache locality
- **≥1000 agents**: Enables additional memory optimizations

### Target Coherence Defaults

Each scale has an optimized default target coherence:

- **Tiny**: 0.95 (very tight synchronization)
- **Small**: 0.90 (tight synchronization)
- **Medium**: 0.85 (good synchronization)
- **Large**: 0.80 (moderate synchronization)
- **Huge**: 0.75 (loose synchronization)

## Usage Examples

### Using with Client API

```go
import (
    "github.com/carlisia/bio-adapt/client/emerge"
    "github.com/carlisia/bio-adapt/emerge/scale"
)

// Simple usage with predefined scales
client := emerge.MinimizeAPICalls(scale.Small)   // 50 agents
client := emerge.MinimizeAPICalls(scale.Medium)  // 200 agents
client := emerge.MinimizeAPICalls(scale.Large)   // 1000 agents
```

### Custom Configuration

```go
// Override defaults if needed
client := emerge.Custom().
    WithScale(scale.Large).
    WithTargetCoherence(0.95).  // Override default 0.80
    Build()
```

## Choosing the Right Scale

### Guidelines

- **Development/Testing**: Use Tiny or Small
- **Production Start**: Begin with Medium
- **Scaling Up**: Move to Large when you have 500+ concurrent operations
- **Maximum Scale**: Use Huge for 1000+ concurrent operations

### Scale Selection Factors

1. **Number of concurrent operations** - Primary factor
2. **Available memory** - Check resource requirements
3. **Convergence time requirements** - Larger scales take longer
4. **Network capacity** - Larger scales need more bandwidth

## See Also

- [Performance Optimization](emerge/optimization.md)
- [Deployment Guide](deployment.md)
- [E2E Testing](testing/e2e.md)
