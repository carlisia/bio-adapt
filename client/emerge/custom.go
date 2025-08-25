package emerge

import (
	"fmt"
	"time"

	"github.com/carlisia/bio-adapt/emerge/core"
	"github.com/carlisia/bio-adapt/emerge/goal"
	"github.com/carlisia/bio-adapt/emerge/scale"
	"github.com/carlisia/bio-adapt/emerge/swarm"
)

// Builder provides a fluent API for configuring emerge clients.
// It follows the builder pattern for complex configuration.
type Builder struct {
	goal            goal.Type
	scale           scale.Size
	targetCoherence float64
	// Future: could add more options like custom coupling, topology, etc.
}

// New creates a new builder with sensible defaults.
// This is the starting point for custom configuration.
func New() *Builder {
	return &Builder{
		goal:  goal.MinimizeAPICalls, // Most common use case
		scale: scale.Tiny,            // Start small by default
		// targetCoherence: 0 means use scale's default
	}
}

// WithGoal sets the synchronization goal.
func (b *Builder) WithGoal(g goal.Type) *Builder {
	b.goal = g
	return b
}

// WithScale sets the swarm scale (determines agent count and default coherence).
func (b *Builder) WithScale(s scale.Size) *Builder {
	b.scale = s
	return b
}

// WithTargetCoherence sets a custom target coherence.
// If not called, uses the scale's default coherence.
func (b *Builder) WithTargetCoherence(coherence float64) *Builder {
	b.targetCoherence = coherence
	return b
}

// Build creates the emerge client with the configured options.
func (b *Builder) Build() (*Client, error) {
	agentCount := b.scale.DefaultAgentCount()
	config := swarm.For(b.goal).WithSize(agentCount)

	// Determine target coherence
	targetCoherence := b.targetCoherence
	if targetCoherence <= 0 {
		targetCoherence = b.scale.DefaultTargetCoherence()
	}

	// Create target state
	targetState := core.State{
		Phase:     0,
		Frequency: 1000 * time.Millisecond,
		Coherence: targetCoherence,
	}

	// Create swarm
	sw, err := swarm.New(agentCount, targetState, swarm.WithGoalConfig(config))
	if err != nil {
		return nil, fmt.Errorf("failed to create swarm: %w", err)
	}

	return &Client{
		swarm:  sw,
		config: config,
	}, nil
}

// Custom creates a client with full configuration control.
// This is an alias for New() to make intent clear.
func Custom() *Builder {
	return New()
}

// Functional Options Pattern (alternative approach)
// This pattern is very idiomatic in Go for optional configuration.

// Option configures an emerge Client.
type Option func(*Builder) error

// WithGoalOption returns an option that sets the goal.
func WithGoalOption(g goal.Type) Option {
	return func(b *Builder) error {
		b.goal = g
		return nil
	}
}

// WithScaleOption returns an option that sets the scale.
func WithScaleOption(s scale.Size) Option {
	return func(b *Builder) error {
		b.scale = s
		return nil
	}
}

// WithCoherenceOption returns an option that sets the target coherence.
func WithCoherenceOption(coherence float64) Option {
	return func(b *Builder) error {
		if coherence <= 0 || coherence > 1 {
			return fmt.Errorf("coherence must be between 0 and 1, got %f", coherence)
		}
		b.targetCoherence = coherence
		return nil
	}
}

// NewWithOptions creates a client using functional options.
// This is another idiomatic pattern in Go.
func NewWithOptions(opts ...Option) (*Client, error) {
	b := New()

	for _, opt := range opts {
		if err := opt(b); err != nil {
			return nil, err
		}
	}

	return b.Build()
}

// Examples of usage:
//
// 1. Builder pattern:
//    client := emerge.New().
//        WithGoal(goal.DistributeLoad).
//        WithScale(scale.Large).
//        WithTargetCoherence(0.95).
//        Build()
//
// 2. With functional options:
//    client := emerge.NewWithOptions(
//        emerge.WithGoalOption(goal.MinimizeAPICalls),
//        emerge.WithScaleOption(scale.Large),
//        emerge.WithCoherenceOption(0.90),
//    )
//
// 3. Using Custom alias for clarity:
//    client := emerge.Custom().
//        WithGoal(goal.MinimizeAPICalls).
//        WithScale(scale.Medium).
//        Build()
