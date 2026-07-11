package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthzHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/healthz", nil)

	healthzHandler(c)

	if w.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", w.Code)
	}
}

// TestReadyzHandler verifies the readiness probe reflects DB reachability.
func TestReadyzHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orig := fileDB
	t.Cleanup(func() { fileDB = orig })

	// No DB → not ready (503).
	fileDB = nil
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyzHandler(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz (no db) status = %d, want 503", w.Code)
	}

	// Reachable in-memory DB → ready (200). Driver is registered via filedb.go's
	// blank import in this package.
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	fileDB = db

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyzHandler(c2)
	if w2.Code != http.StatusOK {
		t.Fatalf("readyz (db up) status = %d, want 200", w2.Code)
	}
}
