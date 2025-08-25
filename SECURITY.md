# Security policy

## Supported versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a vulnerability

We take security seriously. If you discover a security vulnerability, please:

1. **DO NOT** open a public issue
2. Email security concerns to: [maintainer email]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We'll respond within 48 hours and work with you to address the issue.

## Security considerations

When using bio-adapt in production:

### Resource limits

- **Swarm size**: Validate swarm size limits to prevent resource exhaustion
- **Agent creation**: Implement rate limiting for agent creation
- **Memory bounds**: Monitor memory usage, especially for swarms >1000 agents

### Convergence safety

- **Timeouts**: Always use context timeouts for convergence operations
- **Energy levels**: Monitor agent energy to detect potential DoS patterns
- **Coherence thresholds**: Set minimum acceptable coherence levels

### Configuration hardening

```go
// Example: Safe client creation with limits
import (
    "github.com/carlisia/bio-adapt/client/emerge"
    "github.com/carlisia/bio-adapt/emerge/scale"
)

const maxAgentCount = 5000

func CreateClient(desiredScale scale.Size) (*emerge.Client, error) {
    // Validate scale
    agentCount := desiredScale.DefaultAgentCount()
    if agentCount > maxAgentCount {
        return nil, fmt.Errorf("scale %s has %d agents, exceeds limit %d",
            desiredScale, agentCount, maxAgentCount)
    }

    // Create client with safe configuration
    client := emerge.MinimizeAPICalls(desiredScale)

    // Use context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    go client.Start(ctx)
    return client, nil
}
```

### Monitoring recommendations

- Track convergence times - unusually long convergence may indicate issues
- Monitor coherence levels - sudden drops may indicate disruption attacks
- Watch resource consumption - CPU and memory per agent
- Set alerts for anomalous behavior patterns

## Dependencies

Bio-adapt has minimal dependencies. We regularly:

- Audit all dependencies for vulnerabilities
- Keep dependencies up to date
- Use `go mod tidy` to remove unused dependencies

To check for vulnerabilities in dependencies:

```bash
task vuln
```

## Best practices

1. **Input validation**: Always validate swarm configuration parameters
2. **Context usage**: Use contexts with timeouts for all operations
3. **Resource monitoring**: Implement metrics collection for production deployments
4. **Graceful degradation**: Handle partial swarm failures gracefully
5. **Audit logging**: Log significant events for security analysis
