package main

import (
	"testing"
	"time"
)

// TestBlacklistToken verifies that revoking a token via logout makes it appear
// blacklisted, so AuthMiddleware can reject it before its natural expiry.
func TestBlacklistToken(t *testing.T) {
	tok := "test-token-blacklist-1"
	t.Cleanup(func() {
		jwtBlacklistMu.Lock()
		delete(jwtBlacklist, tok)
		jwtBlacklistMu.Unlock()
	})

	if isTokenBlacklisted(tok) {
		t.Fatal("token should not be blacklisted before revocation")
	}
	blacklistToken(tok, time.Now().Add(time.Hour))
	if !isTokenBlacklisted(tok) {
		t.Fatal("token should be blacklisted after revocation")
	}
}

// TestCleanupBlacklist verifies that only entries past their expiry are purged,
// keeping the blacklist from growing without bound while preserving still-valid
// revocations.
func TestCleanupBlacklist(t *testing.T) {
	expired := "expired-token"
	active := "active-token"
	t.Cleanup(func() {
		jwtBlacklistMu.Lock()
		delete(jwtBlacklist, expired)
		delete(jwtBlacklist, active)
		jwtBlacklistMu.Unlock()
	})

	blacklistToken(expired, time.Now().Add(-time.Minute)) // already past
	blacklistToken(active, time.Now().Add(time.Hour))

	cleanupBlacklist()

	if isTokenBlacklisted(expired) {
		t.Error("expired token should have been purged")
	}
	if !isTokenBlacklisted(active) {
		t.Error("active token should remain blacklisted")
	}
}
