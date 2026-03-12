package scratch

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucket_Burst(t *testing.T) {
	tb := NewTokenBucket(1, 5) // 1/s, burst of 5

	// should allow 5 back-to-back immediately
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Fatalf("expected Allow() == true on request %d", i+1)
		}
	}
	// bucket is now empty
	if tb.Allow() {
		t.Fatal("expected Allow() == false after burst exhausted")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(10, 1) // 10 tokens/s, burst 1

	if !tb.Allow() {
		t.Fatal("first request should be allowed")
	}
	if tb.Allow() {
		t.Fatal("should be denied immediately after burst")
	}

	time.Sleep(150 * time.Millisecond) // ~1.5 tokens should refill

	if !tb.Allow() {
		t.Fatal("should be allowed after refill")
	}
}

func TestTokenBucket_Concurrent(t *testing.T) {
	const (
		burst       = 100
		goroutines  = 50
		callsEach   = 10
	)
	tb := NewTokenBucket(0, burst) // no refill — fixed budget

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < callsEach; j++ {
				if tb.Allow() {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	// exactly burst tokens available, never more
	if got := allowed.Load(); got > burst {
		t.Fatalf("allowed %d requests but burst cap is %d", got, burst)
	}
}
