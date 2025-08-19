//go:build !nogossip
// +build !nogossip

package attractor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/hashicorp/memberlist"
)

// GossipAgent extends Agent with distributed coordination capabilities.
// Instead of central control, agents gossip with neighbors to achieve
// emergent synchronization - biological behavior.
type GossipAgent struct {
	*Agent

	// Gossip protocol for distributed coordination
	config *memberlist.Config
	list   *memberlist.Memberlist

	// Local state tracking
	updates  chan StateUpdate
	shutdown chan struct{}
}

// NewGossipAgent creates an agent capable of distributed coordination.
// Uses HashiCorp's memberlist for efficient gossip protocol.
func NewGossipAgent(id string, bindPort int) (*GossipAgent, error) {
	ga := &GossipAgent{
		Agent:    NewAgent(id),
		updates:  make(chan StateUpdate, 100),
		shutdown: make(chan struct{}),
	}

	// Configure gossip protocol
	config := memberlist.DefaultLocalConfig()
	config.Name = id
	config.BindPort = bindPort
	config.AdvertisePort = bindPort

	// Create gossip network
	list, err := memberlist.Create(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create memberlist: %w", err)
	}

	ga.config = config
	ga.list = list

	return ga, nil
}

// Run executes the agent's autonomous lifecycle - no central control needed.
// Agent continuously:
//  1. Observes neighbors through gossip
//  2. Updates local context
//  3. Makes autonomous decisions
//  4. Adjusts toward blended goals
//
// This creates emergent behavior from local interactions.
func (ga *GossipAgent) Run(ctx context.Context, globalGoal State) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ga.shutdown:
			return

		case update := <-ga.updates:
			// Process neighbor update
			ga.processNeighborUpdate(update)

		case <-ticker.C:
			// Autonomous decision cycle
			ga.UpdateContext()

			// Propose adjustment based on blended goals
			action, accepted := ga.ProposeAdjustment(globalGoal)

			if accepted {
				success, _ := ga.ApplyAction(action)
				if success {
					// Gossip new state to neighbors
					ga.broadcastState()

					// Regenerate some energy over time
					ga.regenerateEnergy(0.5)
				}
			}
		}
	}
}

// processNeighborUpdate integrates information from gossip network.
func (ga *GossipAgent) processNeighborUpdate(update StateUpdate) {
	// Update neighbor registry
	if neighbor, exists := ga.neighbors.Load(update.FromID); exists {
		n := neighbor.(*Agent)
		n.phase.Store(update.Phase)
		n.frequency.Store(update.Frequency)
		n.energy.Store(update.Energy)
	}
}

// broadcastState shares current state with neighbors via gossip.
func (ga *GossipAgent) broadcastState() {
	update := StateUpdate{
		FromID:    ga.ID,
		Phase:     ga.phase.Load(),
		Frequency: ga.frequency.Load(),
		Energy:    ga.energy.Load(),
	}

	// In real implementation, would use memberlist delegate
	// For simplicity, just show the structure
	for _, member := range ga.list.Members() {
		if member.Name != ga.ID {
			// Would send update to member
			_ = update
		}
	}
}

// regenerateEnergy slowly restores energy over time (metabolic process).
func (ga *GossipAgent) regenerateEnergy(rate float64) {
	current := ga.energy.Load()
	max := 100.0
	if current < max {
		ga.energy.Store(math.Min(current+rate, max))
	}
}

// Stop shuts down the gossip agent.
func (ga *GossipAgent) Stop() error {
	close(ga.shutdown)
	if ga.list != nil {
		if err := ga.list.Leave(time.Second); err != nil {
			return fmt.Errorf("failed to leave gossip network: %w", err)
		}
		if err := ga.list.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown gossip network: %w", err)
		}
	}
	return nil
}
