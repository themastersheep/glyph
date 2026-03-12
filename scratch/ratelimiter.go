package scratch

import (
	"sync"
	"time"
)

// TokenBucket is a rate limiter using the token bucket algorithm.
type TokenBucket struct {
	mu       sync.Mutex
	tokens   float64
	capacity float64   // max tokens (burst size)
	rate     float64   // tokens added per second
	lastTick time.Time
}

// NewTokenBucket creates a rate limiter that allows up to burst requests instantly,
// then refills at rate tokens per second.
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		tokens:   float64(burst),
		capacity: float64(burst),
		rate:     rate,
		lastTick: time.Now(),
	}
}

// Allow returns true if a token is available and consumes it, false otherwise.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTick).Seconds()
	tb.lastTick = now

	// refill tokens based on elapsed time, capped at capacity
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.rate)

	if tb.tokens < 1 {
		return false
	}
	tb.tokens--
	return true
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
