package topology

import (
	"fmt"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// Star creates a star network topology.
// One central agent is connected to all others.
func Star(s *swarm.Swarm) error {
	var agents []*agent.Agent
	s.ForEachAgent(func(a *agent.Agent) bool {
		agents = append(agents, a)
		return true
	})

	if len(agents) < 2 {
		return fmt.Errorf("%w for star topology: got %d, need at least 2", core.ErrInsufficientAgents, len(agents))
	}

	hub := agents[0]

	for i, a := range agents {
		if i == 0 {
			// Hub connects to everyone
			for j, neighbor := range agents {
				if j != 0 {
					hub.ConnectTo(neighbor)
				}
			}
		} else {
			// Everyone else connects only to hub
			a.ConnectTo(hub)
		}
	}

	return nil
}
