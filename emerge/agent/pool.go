package agent

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/decision"
	"github.com/carlisia/bio-adapt/internal/random"
)

// Pool manages a pool of reusable Agent objects to reduce allocations.
var Pool = &PoolManager{
	pool: sync.Pool{
		New: func() interface{} {
			// Create a new agent with minimal initialization
			a := &Agent{}
			// neighbors is already a sync.Map field, not a pointer
			return a
		},
	},
	stats: &PoolStats{},
}

// PoolManager provides pooled Agent objects to reduce GC pressure.
type PoolManager struct {
	pool  sync.Pool
	stats *PoolStats
}

// PoolStats tracks pool usage metrics.
type PoolStats struct {
	gets  atomic.Uint64
	puts  atomic.Uint64
	news  atomic.Uint64
	inUse atomic.Int64
}

// Get retrieves an Agent from the pool or creates a new one.
func (p *PoolManager) Get(id string) *Agent {
	p.stats.gets.Add(1)
	p.stats.inUse.Add(1)

	val := p.pool.Get()
	a, ok := val.(*Agent)
	if !ok {
		panic("pool returned non-Agent type")
	}

	// Check if this is a newly created agent
	if a.ID == "" {
		p.stats.news.Add(1)
	}

	// Initialize/reset the agent
	a.Reset(id)
	return a
}

// Put returns an Agent to the pool for reuse.
func (p *PoolManager) Put(a *Agent) {
	if a == nil {
		return
	}

	p.stats.puts.Add(1)
	p.stats.inUse.Add(-1)

	// Clear sensitive data but keep allocated structures
	a.Clear()
	p.pool.Put(a)
}

// Stats returns current pool statistics.
func (p *PoolManager) Stats() (gets, puts, news uint64, inUse int64) {
	return p.stats.gets.Load(),
		p.stats.puts.Load(),
		p.stats.news.Load(),
		p.stats.inUse.Load()
}

// Reset initializes an agent with the given ID, preserving allocated memory.
func (a *Agent) Reset(id string) {
	a.ID = id

	// Reset atomic values to defaults
	a.phase.Store(random.Phase())
	a.frequency.Store(100 * time.Millisecond)
	a.energy.Store(100.0)
	a.localGoal.Store(random.Phase())
	a.influence.Store(0.1 + random.Float64()*0.1)
	a.stubbornness.Store(random.Float64() * 0.3)

	// Clear neighbors but keep the map allocated
	a.neighbors.Range(func(key, _ interface{}) bool {
		a.neighbors.Delete(key)
		return true
	})

	// Reset context
	a.context.Store(core.Context{})

	// Keep decision maker if set, otherwise use default
	if a.decider == nil {
		a.decider = &decision.SimpleDecisionMaker{}
	}
}

// Clear removes all data from the agent but keeps allocated structures.
func (a *Agent) Clear() {
	a.ID = ""

	// Clear neighbors
	a.neighbors.Range(func(key, _ interface{}) bool {
		a.neighbors.Delete(key)
		return true
	})

	// Don't nil out structures, just clear them
	a.context.Store(core.Context{})
}
