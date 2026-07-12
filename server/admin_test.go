package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newAdminRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/status", adminStatusHandler)
	r.POST("/admin/login", adminLoginHandler)
	r.POST("/admin/logout", adminLogoutHandler)
	r.POST("/admin/password", adminSetPasswordHandler)
	gated := r.Group("/admin")
	gated.Use(AdminAuthMiddleware())
	gated.GET("/users", adminUsersHandler)
	gated.GET("/topology", adminTopologyHandler)
	return r
}

// do issues a JSON request, replaying any provided cookie, and returns the
// recorder so callers can inspect status, body, and Set-Cookie.
func do(r *gin.Engine, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func adminCookie(w *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == adminCookieName && c.Value != "" {
			return c
		}
	}
	return nil
}

func TestAdminAuthFlow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := initFileDB(); err != nil {
		t.Fatalf("initFileDB: %v", err)
	}
	r := newAdminRouter()

	// Fresh install: not configured, not authenticated.
	w := do(r, "GET", "/admin/status", nil, nil)
	var st struct{ Configured, Authenticated bool }
	json.Unmarshal(w.Body.Bytes(), &st)
	if st.Configured || st.Authenticated {
		t.Fatalf("fresh status: got %+v, want both false", st)
	}

	// Too-short password is rejected.
	if w := do(r, "POST", "/admin/password", map[string]string{"password": "short"}, nil); w.Code != http.StatusBadRequest {
		t.Fatalf("short password: got %d, want 400", w.Code)
	}

	// Bootstrap: sets the password and logs the admin in (returns a cookie).
	w = do(r, "POST", "/admin/password", map[string]string{"password": "supersecret"}, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("bootstrap: got %d, want 200", w.Code)
	}
	boot := adminCookie(w)
	if boot == nil {
		t.Fatal("bootstrap did not set an admin cookie")
	}

	// Now configured; the bootstrap cookie authenticates.
	w = do(r, "GET", "/admin/status", nil, boot)
	json.Unmarshal(w.Body.Bytes(), &st)
	if !st.Configured || !st.Authenticated {
		t.Fatalf("post-bootstrap status: got %+v, want both true", st)
	}

	// Wrong password is rejected; correct password issues a session.
	if w := do(r, "POST", "/admin/login", map[string]string{"password": "nope"}, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("bad login: got %d, want 401", w.Code)
	}
	w = do(r, "POST", "/admin/login", map[string]string{"password": "supersecret"}, nil)
	if w.Code != http.StatusOK || adminCookie(w) == nil {
		t.Fatalf("good login: code=%d cookie=%v", w.Code, adminCookie(w))
	}

	// Changing the password without a session is blocked even with the right
	// current password.
	if w := do(r, "POST", "/admin/password",
		map[string]string{"password": "brandnewpass", "current_password": "supersecret"}, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauth change: got %d, want 401", w.Code)
	}

	// With a session but a wrong current password → 403.
	if w := do(r, "POST", "/admin/password",
		map[string]string{"password": "brandnewpass", "current_password": "wrong"}, boot); w.Code != http.StatusForbidden {
		t.Fatalf("wrong current: got %d, want 403", w.Code)
	}

	// With a session and correct current password → change succeeds.
	if w := do(r, "POST", "/admin/password",
		map[string]string{"password": "brandnewpass", "current_password": "supersecret"}, boot); w.Code != http.StatusOK {
		t.Fatalf("valid change: got %d, want 200", w.Code)
	}
}

func TestAdminUsers(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := initFileDB(); err != nil {
		t.Fatalf("initFileDB: %v", err)
	}
	r := newAdminRouter()

	// Gated: no admin session → 401.
	if w := do(r, "GET", "/admin/users", nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("users gate: got %d, want 401", w.Code)
	}

	cookie := adminCookie(do(r, "POST", "/admin/password", map[string]string{"password": "supersecret"}, nil))
	u, err := upsertOAuthUser("google", "pid-users", "Bob", "bob@example.com", "")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	agentMu.Lock()
	agentStore["a-users-1"] = &AgentRecord{AgentID: "a-users-1", Name: "Desk", UserID: u.ID, LastSeen: time.Now()}
	agentMu.Unlock()
	defer func() {
		agentMu.Lock()
		delete(agentStore, "a-users-1")
		agentMu.Unlock()
	}()

	w := do(r, "GET", "/admin/users", nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("users: got %d, want 200", w.Code)
	}
	var resp struct {
		Users []struct {
			UserID      string `json:"user_id"`
			Email       string `json:"email"`
			IsGuest     bool   `json:"is_guest"`
			DeviceCount int    `json:"device_count"`
		} `json:"users"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, us := range resp.Users {
		if us.UserID == u.ID {
			found = true
			if us.Email != "bob@example.com" || us.IsGuest || us.DeviceCount != 1 {
				t.Fatalf("user row wrong: %+v", us)
			}
		}
	}
	if !found {
		t.Fatal("seeded user missing from users list")
	}
}

func TestAdminTopology(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := initFileDB(); err != nil {
		t.Fatalf("initFileDB: %v", err)
	}
	r := newAdminRouter()

	// Gated: no admin session → 401.
	if w := do(r, "GET", "/admin/topology", nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("topology gate: got %d, want 401", w.Code)
	}

	// Bootstrap to obtain an admin session cookie.
	cookie := adminCookie(do(r, "POST", "/admin/password", map[string]string{"password": "supersecret"}, nil))
	if cookie == nil {
		t.Fatal("no admin cookie from bootstrap")
	}

	// Seed a user and an owned agent.
	u, err := upsertOAuthUser("google", "pid-1", "Alice", "alice@example.com", "")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	agentMu.Lock()
	agentStore["a-topo-1"] = &AgentRecord{AgentID: "a-topo-1", Name: "Laptop", UserID: u.ID, LastSeen: time.Now()}
	agentMu.Unlock()
	defer func() {
		agentMu.Lock()
		delete(agentStore, "a-topo-1")
		agentMu.Unlock()
	}()

	w := do(r, "GET", "/admin/topology", nil, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("topology: got %d, want 200", w.Code)
	}
	var resp struct {
		Users []struct {
			UserID      string `json:"user_id"`
			DeviceCount int    `json:"device_count"`
			Devices     []struct {
				AgentID string `json:"agent_id"`
				Online  bool   `json:"online"`
			} `json:"devices"`
		} `json:"users"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, us := range resp.Users {
		if us.UserID != u.ID {
			continue
		}
		found = true
		if us.DeviceCount != 1 || len(us.Devices) != 1 || us.Devices[0].AgentID != "a-topo-1" || !us.Devices[0].Online {
			t.Fatalf("alice topology wrong: %+v", us)
		}
	}
	if !found {
		t.Fatal("seeded user missing from topology")
	}
}
