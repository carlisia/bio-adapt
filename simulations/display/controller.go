// controller.go implements user input handling.
// This can be swapped for different input methods (keyboard, network, etc.)

package display

import (
	"time"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// KeyboardController implements Controller for keyboard input
type KeyboardController struct {
	events chan ControlEvent
}

// NewKeyboardController creates a new keyboard controller
func NewKeyboardController() Controller {
	return &KeyboardController{
		events: make(chan ControlEvent, 10),
	}
}

// Events returns the event channel
func (c *KeyboardController) Events() <-chan ControlEvent {
	return c.events
}

// ProcessKey processes a keyboard event
func (c *KeyboardController) ProcessKey(k *terminalapi.Keyboard) {
	var eventType EventType

	//nolint:exhaustive // We handle specific keys and ignore others
	switch k.Key {
	case 'q', 'Q':
		eventType = EventQuit
	case 'r', 'R':
		eventType = EventReset
	case 'd', 'D':
		eventType = EventDisrupt
	case 'p', 'P':
		// P toggles pause/resume
		eventType = EventPause
	case ' ':
		eventType = EventResume
	case keyboard.KeyEsc:
		eventType = EventQuit
	// Scale switching keys
	case '1':
		eventType = ScaleTiny
	case '2':
		eventType = ScaleSmall
	case '3':
		eventType = ScaleMedium
	case '4':
		eventType = ScaleLarge
	case '5':
		eventType = ScaleHuge
	// Goal optimization keys
	case 'b', 'B':
		eventType = GoalBatch
	case 'l', 'L':
		eventType = GoalLoad
	case 'c', 'C':
		eventType = GoalConsensus
	case 't', 'T':
		eventType = GoalLatency
	case 'e', 'E':
		eventType = GoalEnergy
	case 'm', 'M': // Using M for rhythM since R is Reset
		eventType = GoalRhythm
	case 'f', 'F':
		eventType = GoalFailure
	case 'a', 'A':
		eventType = GoalTraffic
	// Pattern switching keys
	case 'h', 'H':
		eventType = PatternHighFreq
	case 'u', 'U':
		eventType = PatternBurst
	case 'y', 'Y':
		eventType = PatternSteady
	case 'x', 'X':
		eventType = PatternMixed
	case 'z', 'Z':
		eventType = PatternSparse
	default:
		return // Ignore other keys
	}

	select {
	case c.events <- ControlEvent{
		Type: eventType,
	}:
	default:
		// Drop event if channel is full
	}
}

// NetworkController could implement Controller for remote control
type NetworkController struct {
	events chan ControlEvent
	// Add network handling fields
}

// NewNetworkController creates a controller that accepts commands over network
func NewNetworkController(_ string) Controller {
	// Implementation for network-based control
	return &NetworkController{
		events: make(chan ControlEvent, 10),
	}
}

// Events returns the event channel
func (c *NetworkController) Events() <-chan ControlEvent {
	return c.events
}

// AutomatedController implements Controller for automated testing
type AutomatedController struct {
	events chan ControlEvent
	script []ControlEvent
	delays []time.Duration
}

// NewAutomatedController creates a controller that runs a scripted sequence
func NewAutomatedController(script []ControlEvent, delays []time.Duration) Controller {
	c := &AutomatedController{
		events: make(chan ControlEvent, 10),
		script: script,
		delays: delays,
	}
	go c.run()
	return c
}

// Events returns the event channel
func (c *AutomatedController) Events() <-chan ControlEvent {
	return c.events
}

// run executes the scripted events
func (c *AutomatedController) run() {
	for i, event := range c.script {
		if i < len(c.delays) {
			time.Sleep(c.delays[i])
		}
		c.events <- event
	}
}
