package topology

import (
	"github.com/carlisia/bio-adapt/emerge/agent"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// FullyConnected creates a fully connected network topology.
// Every agent is connected to every other agent.
func FullyConnected(s *swarm.Swarm) error {
	var agents []*agent.Agent
	s.ForEachAgent(func(a *agent.Agent) bool {
		agents = append(agents, a)
		return true
	})

	for i, a := range agents {
		for j, neighbor := range agents {
			if i != j {
				a.ConnectTo(neighbor.ID, neighbor)
			}
		}
	}

	return nil
}
