package topology

import (
	"fmt"

	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// Ring creates a ring network topology.
// Each agent is connected to its immediate neighbors in a circle.
func Ring(s *swarm.Swarm) error {
	var agents []*agent.Agent
	s.ForEachAgent(func(a *agent.Agent) bool {
		agents = append(agents, a)
		return true
	})

	n := len(agents)
	if n < 2 {
		return fmt.Errorf("%w for ring topology: got %d, need at least 2", core.ErrInsufficientAgents, n)
	}

	for i, a := range agents {
		// Connect to previous neighbor
		prev := agents[(i-1+n)%n]
		a.ConnectTo(prev)

		// Connect to next neighbor
		next := agents[(i+1)%n]
		a.ConnectTo(next)
	}

	return nil
}
