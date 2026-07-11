package main

import "testing"

// TestFindAgentIDByToken verifies the O(1) token→agentID lookup that replaced
// the previous O(n) linear scan over agentStore, including the miss case.
func TestFindAgentIDByToken(t *testing.T) {
	const (
		id  = "agent-test-index-1"
		tok = "tok-index-abc"
	)

	agentMu.Lock()
	agentStore[id] = &AgentRecord{AgentID: id, Token: tok}
	tokenIndex[tok] = id
	agentMu.Unlock()
	t.Cleanup(func() {
		agentMu.Lock()
		delete(agentStore, id)
		delete(tokenIndex, tok)
		agentMu.Unlock()
	})

	if got := findAgentIDByToken(tok); got != id {
		t.Errorf("findAgentIDByToken(%q) = %q, want %q", tok, got, id)
	}
	if got := findAgentIDByToken("no-such-token"); got != "" {
		t.Errorf("findAgentIDByToken(unknown) = %q, want empty string", got)
	}
}
