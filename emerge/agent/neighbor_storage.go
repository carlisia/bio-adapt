package agent

import (
	"sync"
	"sync/atomic"
)

// NeighborStorage provides an optimized storage for agent neighbors.
// It uses a fixed-size slice for better cache locality and reduced allocations.
type NeighborStorage struct {
	// Fixed-size arrays for better cache locality
	neighbors []*Agent // Pre-allocated slice of neighbors
	ids       []string // Neighbor IDs for lookup
	count     int32    // Atomic counter for active neighbors
	capacity  int      // Maximum capacity
	mu        sync.RWMutex
}

// NewNeighborStorage creates an optimized neighbor storage.
func NewNeighborStorage(capacity int) *NeighborStorage {
	if capacity <= 0 {
		capacity = 20 // Default small-world network size
	}
	return &NeighborStorage{
		neighbors: make([]*Agent, capacity),
		ids:       make([]string, capacity),
		capacity:  capacity,
	}
}

// Store adds or updates a neighbor.
func (ns *NeighborStorage) Store(id string, agent *Agent) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	// Check if already exists
	currentCount := int(atomic.LoadInt32(&ns.count))
	for i := 0; i < currentCount; i++ {
		if ns.ids[i] == id {
			ns.neighbors[i] = agent
			return true
		}
	}

	// Add new neighbor if there's space
	if currentCount < ns.capacity {
		ns.ids[currentCount] = id
		ns.neighbors[currentCount] = agent
		atomic.AddInt32(&ns.count, 1)
		return true
	}

	// No space available
	return false
}

// Load retrieves a neighbor by ID.
func (ns *NeighborStorage) Load(id string) (*Agent, bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	currentCount := int(atomic.LoadInt32(&ns.count))
	for i := 0; i < currentCount; i++ {
		if ns.ids[i] == id {
			return ns.neighbors[i], true
		}
	}
	return nil, false
}

// Delete removes a neighbor.
func (ns *NeighborStorage) Delete(id string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	currentCount := int(atomic.LoadInt32(&ns.count))
	for i := 0; i < currentCount; i++ {
		if ns.ids[i] == id {
			// Swap with last element and decrease count
			lastIdx := currentCount - 1
			if i < lastIdx {
				ns.ids[i] = ns.ids[lastIdx]
				ns.neighbors[i] = ns.neighbors[lastIdx]
			}
			// Clear last position
			ns.ids[lastIdx] = ""
			ns.neighbors[lastIdx] = nil
			atomic.AddInt32(&ns.count, -1)
			return
		}
	}
}

// Range iterates over all neighbors.
func (ns *NeighborStorage) Range(f func(id string, agent *Agent) bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	currentCount := int(atomic.LoadInt32(&ns.count))
	for i := 0; i < currentCount; i++ {
		if !f(ns.ids[i], ns.neighbors[i]) {
			break
		}
	}
}

// Count returns the number of neighbors.
func (ns *NeighborStorage) Count() int {
	return int(atomic.LoadInt32(&ns.count))
}

// All returns a slice of all neighbors (for compatibility).
func (ns *NeighborStorage) All() []*Agent {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	currentCount := int(atomic.LoadInt32(&ns.count))
	if currentCount == 0 {
		return nil
	}

	result := make([]*Agent, currentCount)
	copy(result, ns.neighbors[:currentCount])
	return result
}

// Clear removes all neighbors.
func (ns *NeighborStorage) Clear() {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	currentCount := int(atomic.LoadInt32(&ns.count))
	for i := 0; i < currentCount; i++ {
		ns.ids[i] = ""
		ns.neighbors[i] = nil
	}
	atomic.StoreInt32(&ns.count, 0)
}
