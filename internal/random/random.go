// Package random provides cryptographically secure random number generation.
package random

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"sync"
)

// SecureRandom provides cryptographically secure random number generation
type SecureRandom struct {
	mu sync.Mutex
}

var defaultRandom = &SecureRandom{}

// Float64 returns a cryptographically secure random float64 in [0.0,1.0)
func Float64() float64 {
	return defaultRandom.Float64()
}

// Float64 returns a cryptographically secure random float64 in [0.0,1.0)
func (r *SecureRandom) Float64() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		// This should never happen with crypto/rand
		panic("crypto/rand failed: " + err.Error())
	}

	// Convert to uint64, mask to get 53 bits of precision (same as math/rand)
	u := binary.BigEndian.Uint64(b[:]) & ((1 << 53) - 1)
	// Convert to float64 in [0, 1)
	return float64(u) / float64(1<<53)
}

// Intn returns a cryptographically secure random int in [0,n)
func Intn(n int) int {
	return defaultRandom.Intn(n)
}

// Intn returns a cryptographically secure random int in [0,n)
func (r *SecureRandom) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	if n <= 1 {
		return 0
	}

	// For small n, use simple rejection sampling
	// n must be less than 2^31 to fit in int32
	const maxInt31 = 1<<31 - 1
	if n <= maxInt31 {
		n32 := int32(n)
		maxVal := (maxInt31 / n32) * n32
		for {
			v := r.Int31n()
			if v < maxVal {
				return int(v) % n
			}
		}
	}

	// For large n, use 64-bit
	panic("Intn: n too large")
}

// Int31n returns a cryptographically secure random int32 in [0, 2^31)
func (r *SecureRandom) Int31n() int32 {
	r.mu.Lock()
	defer r.mu.Unlock()

	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}

	// Mask to get 31 bits (removes sign bit)
	u := binary.BigEndian.Uint32(b[:]) & 0x7FFFFFFF
	// Safe conversion since u is guaranteed to be <= 0x7FFFFFFF
	if u > 0x7FFFFFFF {
		panic("unexpected value from random source")
	}
	return int32(u)
}

// Phase returns a random phase value in [0, 2π)
func Phase() float64 {
	return Float64() * 2 * math.Pi
}
