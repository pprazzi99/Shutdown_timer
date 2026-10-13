// Package countdown tracks the time left until a deadline and decides when
// warnings should be shown. It holds no goroutines or timers of its own, so the
// caller drives it with whatever clock it likes.
package countdown

import "time"

// DefaultWarnings are the reminders shown before shutdown.
var DefaultWarnings = []time.Duration{5 * time.Minute, 30 * time.Second}

// Countdown is a deadline plus the set of warnings that have not fired yet.
// It is not safe for concurrent use.
type Countdown struct {
	start    time.Time
	deadline time.Time
	pending  []time.Duration
}

// State is a snapshot of the countdown at a given moment.
type State struct {
	Remaining time.Duration
	// Progress is the elapsed fraction of the countdown, from 0 to 1.
	Progress float64
	// Warnings holds the thresholds crossed since the previous call to Tick.
	Warnings []time.Duration
	Done     bool
}

// New creates a countdown from now until deadline. Warnings at or beyond the
// full length of the countdown are dropped, since they could never be crossed.
//
// Monotonic clock readings are stripped so that the deadline follows the wall
// clock: if the computer sleeps past it, the countdown ends on wake-up.
func New(now, deadline time.Time, warnings []time.Duration) *Countdown {
	now, deadline = now.Round(0), deadline.Round(0)
	total := deadline.Sub(now)

	pending := make([]time.Duration, 0, len(warnings))
	for _, w := range warnings {
		if w > 0 && w < total {
			pending = append(pending, w)
		}
	}
	return &Countdown{start: now, deadline: deadline, pending: pending}
}

// Deadline returns the moment the countdown ends.
func (c *Countdown) Deadline() time.Time { return c.deadline }

// Tick reports the state at now. Each warning is returned exactly once, by the
// first Tick at or after its threshold, unless the countdown is already done.
func (c *Countdown) Tick(now time.Time) State {
	remaining := c.deadline.Sub(now.Round(0))
	if remaining < 0 {
		remaining = 0
	}

	var fired []time.Duration
	kept := c.pending[:0]
	for _, w := range c.pending {
		if remaining <= w {
			fired = append(fired, w)
		} else {
			kept = append(kept, w)
		}
	}
	c.pending = kept

	done := remaining == 0
	if done {
		// Warnings are pointless once the action is about to run.
		fired = nil
	}

	progress := 1.0
	if total := c.deadline.Sub(c.start); total > 0 {
		progress = 1 - float64(remaining)/float64(total)
	}
	return State{Remaining: remaining, Progress: progress, Warnings: fired, Done: done}
}
