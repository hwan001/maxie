package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// idleEvictAfter is how long an IP's limiter may sit unused before it is purged
// to keep the limiter map from growing without bound.
const idleEvictAfter = 10 * time.Minute

// ipRateLimiter holds a per-client-IP token bucket. Entries idle beyond
// idleEvictAfter are evicted by a background sweeper.
type ipRateLimiter struct {
	mu      sync.Mutex
	clients map[string]*ipLimiterEntry
	r       rate.Limit
	burst   int
}

type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	l := &ipRateLimiter{
		clients: make(map[string]*ipLimiterEntry),
		r:       r,
		burst:   burst,
	}
	go l.sweepLoop()
	return l
}

// limiterFor returns the token bucket for an IP, creating it on first use.
func (l *ipRateLimiter) limiterFor(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.clients[ip]
	if !ok {
		e = &ipLimiterEntry{limiter: rate.NewLimiter(l.r, l.burst)}
		l.clients[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

func (l *ipRateLimiter) sweepLoop() {
	ticker := time.NewTicker(idleEvictAfter)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-idleEvictAfter)
		l.mu.Lock()
		for ip, e := range l.clients {
			if e.lastSeen.Before(cutoff) {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

// rateLimit builds a Gin middleware enforcing a per-IP token bucket. All routes
// sharing the returned middleware share one limiter store (e.g. a single "auth
// attempts per IP" budget across login and register).
func rateLimit(r rate.Limit, burst int) gin.HandlerFunc {
	limiter := newIPRateLimiter(r, burst)
	return func(c *gin.Context) {
		if !limiter.limiterFor(c.ClientIP()).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded, try again later"})
			c.Abort()
			return
		}
		c.Next()
	}
}
