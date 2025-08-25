// Package emerge provides an idiomatic Go client for the emerge synchronization framework.
//
// The emerge client enables distributed agents to achieve synchronization through
// emergent behavior, without central coordination. It implements the Kuramoto model
// for phase synchronization, allowing independent agents to coordinate their actions.
//
// Key concepts:
//   - Agents: Autonomous units that synchronize their phase and frequency
//   - Swarm: Collection of agents achieving collective behavior
//   - Coherence: Measure of synchronization (0 = random, 1 = perfect sync)
//   - Goals: High-level objectives (minimize API calls, distribute load)
//
// Basic usage:
//
//	client := emerge.MinimizeAPICalls(scale.Medium)
//	err := client.Start(ctx)
//
// The client provides a clean API over the emerge/swarm framework,
// handling the complexity of distributed synchronization.
package emerge

import (
	"context"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// Client provides access to emerge swarm synchronization.
// It manages a collection of agents that achieve collective behavior
// through local interactions, without central control.
type Client struct {
	swarm  *swarm.Swarm
	config *swarm.Config
}

// Start begins the synchronization process.
// Agents will start adjusting their phases to achieve the target coherence.
// The method blocks until the context is canceled or an error occurs.
func (c *Client) Start(ctx context.Context) error {
	return c.swarm.Run(ctx)
}

// Swarm returns the underlying swarm for advanced operations.
// Most users won't need this - prefer using the client's methods.
func (c *Client) Swarm() *swarm.Swarm {
	return c.swarm
}

// Agents returns all agents in the swarm.
// Each agent has a unique ID and maintains its own phase and frequency.
func (c *Client) Agents() map[string]*agent.Agent {
	return c.swarm.Agents()
}

// Coherence returns the current synchronization coherence (0 to 1).
// Higher values indicate better synchronization among agents.
func (c *Client) Coherence() float64 {
	return c.swarm.CurrentCoherence()
}

// IsConverged returns true if the swarm has achieved its target coherence.
func (c *Client) IsConverged() bool {
	return c.swarm.IsConverged()
}

// Stop gracefully stops the swarm synchronization.
// Note: Currently the swarm stops when its context is canceled.
// This method is here for API completeness.
func (*Client) Stop() {
	// The swarm doesn't have a Stop method - it stops when context is canceled
	// This is a no-op for now but keeps the API consistent
}

// Size returns the number of agents in the swarm.
func (c *Client) Size() int {
	return c.swarm.Size()
}

// Config returns the swarm configuration.
// This includes goal parameters and scale settings.
func (c *Client) Config() *swarm.Config {
	return c.config
}
