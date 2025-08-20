package topology

import (
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// Builder defines how to construct network topologies for agent connections.
type Builder func(*swarm.Swarm) error

// BuilderWithParams allows topology builders to accept configuration parameters.
type BuilderWithParams interface {
	Build(s *swarm.Swarm) error
	WithParams(params map[string]interface{}) BuilderWithParams
}
