package main

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// TestRateLimiterBurstThenBlock verifies a per-IP bucket allows up to its burst
// and then blocks, and that each IP has an independent budget.
func TestRateLimiterBurstThenBlock(t *testing.T) {
	// Refill effectively never within the test (1h), so only the burst applies.
	l := newIPRateLimiter(rate.Every(time.Hour), 2)

	const ip = "1.2.3.4"
	if !l.limiterFor(ip).Allow() {
		t.Fatal("1st request should be allowed")
	}
	if !l.limiterFor(ip).Allow() {
		t.Fatal("2nd request should be allowed (within burst)")
	}
	if l.limiterFor(ip).Allow() {
		t.Fatal("3rd request should be blocked (burst exhausted)")
	}

	// A different client IP has its own independent bucket.
	if !l.limiterFor("5.6.7.8").Allow() {
		t.Fatal("a new IP should be allowed")
	}
}
