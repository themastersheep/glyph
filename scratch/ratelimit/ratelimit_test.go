package ratelimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBurstAllowed(t *testing.T) {
	// capacity of 5 should allow 5 immediate requests
	l := New(5, 1)
	for i := range 5 {
		if !l.Allow() {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestBurstExceeded(t *testing.T) {
	l := New(3, 1)
	for range 3 {
		l.Allow()
	}
	if l.Allow() {
		t.Fatal("request beyond burst capacity should be denied")
	}
}

func TestRefill(t *testing.T) {
	// drain entirely, then wait for one token to refill
	l := New(2, 10) // 10 tokens/sec = one token every 100ms
	l.Allow()
	l.Allow()

	time.Sleep(120 * time.Millisecond)

	if !l.Allow() {
		t.Fatal("token should have refilled after wait")
	}
}

func TestCeiling(t *testing.T) {
	l := New(3, 100) // very fast refill
	// let it "overfill" time-wise
	time.Sleep(200 * time.Millisecond)

	// should still only allow capacity, not more
	allowed := 0
	for range 10 {
		if l.Allow() {
			allowed++
		}
	}
	if allowed > 3 {
		t.Fatalf("allowed %d requests, expected at most capacity (3)", allowed)
	}
}

func TestConcurrentSafety(t *testing.T) {
	l := New(100, 0) // fixed pool, no refill
	var allowed atomic.Int64
	var wg sync.WaitGroup

	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got != 100 {
		t.Fatalf("expected exactly 100 allowed, got %d", got)
	}
}
