package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a token bucket rate limiter: capacity refills at a steady rate
// up to a ceiling, allowing bursts while preventing unlimited accumulation.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64    // max tokens (burst ceiling)
	rate     float64    // tokens added per second
	last     time.Time  // last refill timestamp
}

// New creates a Limiter with the given burst capacity and refill rate (tokens/sec).
func New(capacity float64, rate float64) *Limiter {
	return &Limiter{
		tokens:   capacity,
		capacity: capacity,
		rate:     rate,
		last:     time.Now(),
	}
}

// Allow reports whether a request is permitted, consuming one token if so.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// refill tokens proportional to elapsed time, capped at capacity
	elapsed := now.Sub(l.last).Seconds()
	l.tokens = min(l.capacity, l.tokens+elapsed*l.rate)
	l.last = now

	if l.tokens < 1 {
		return false
	}

	l.tokens--
	return true
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
