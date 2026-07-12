package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newAdminRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/status", adminStatusHandler)
	r.POST("/admin/login", adminLoginHandler)
	r.POST("/admin/logout", adminLogoutHandler)
	r.POST("/admin/password", adminSetPasswordHandler)
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
