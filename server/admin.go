package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	adminCookieName      = "admin_jwt"
	adminSessionDuration = 12 * time.Hour
	minAdminPasswordLen  = 8
)

// AdminClaims is the JWT payload for an authenticated admin console session.
// Admin auth is intentionally separate from user login (Claims) — the console
// is protected by its own password, not by a user's OAuth/guest identity.
type AdminClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// adminConfigured reports whether an admin password has been set yet.
func adminConfigured() (bool, error) {
	_, found, err := getConfig(cfgKeyAdminPasswordHash)
	return found, err
}

// issueAdminCookie signs a short-lived admin JWT and sets it as an HttpOnly cookie.
func issueAdminCookie(c *gin.Context) error {
	expiry := time.Now().Add(adminSessionDuration)
	claims := AdminClaims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	if err != nil {
		return err
	}
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     adminCookieName,
		Value:    token,
		Expires:  expiry,
		HttpOnly: true,
		Secure:   cookieDomain != "",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   cookieDomain,
	})
	return nil
}

// clearAdminCookie expires the admin session cookie.
func clearAdminCookie(c *gin.Context) {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieDomain != "",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   cookieDomain,
	})
}

// isAdminAuthenticated verifies the admin JWT cookie is present and valid.
func isAdminAuthenticated(c *gin.Context) bool {
	tokenString, err := c.Cookie(adminCookieName)
	if err != nil {
		return false
	}
	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	return err == nil && token.Valid && claims.Role == "admin"
}

// adminStatusHandler (GET /admin/status, public) tells the UI whether to render
// first-time setup, a login prompt, or the authenticated console.
func adminStatusHandler(c *gin.Context) {
	configured, err := adminConfigured()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read admin config"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configured":    configured,
		"authenticated": isAdminAuthenticated(c),
	})
}

// adminSetPasswordHandler (POST /admin/password) bootstraps the password on
// first run (no auth) or changes it later (requires an admin session AND the
// current password). The password is stored only as a bcrypt hash.
func adminSetPasswordHandler(c *gin.Context) {
	var body struct {
		Password        string `json:"password"`
		CurrentPassword string `json:"current_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(body.Password) < minAdminPasswordLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	configured, err := adminConfigured()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read admin config"})
		return
	}

	// Changing an existing password requires proof: a valid admin session and
	// the current password. Only first-time bootstrap is open.
	if configured {
		if !isAdminAuthenticated(c) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "admin authentication required"})
			return
		}
		hash, _, _ := getConfig(cfgKeyAdminPasswordHash)
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.CurrentPassword)) != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "current password is incorrect"})
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	if err := setConfig(cfgKeyAdminPasswordHash, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save password"})
		return
	}

	// Log the admin in right after bootstrap so first-run flows straight into
	// the console without a second prompt.
	if !configured {
		if err := issueAdminCookie(c); err != nil {
			log.Printf("issueAdminCookie after bootstrap: %v", err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "password saved"})
}

// adminLoginHandler (POST /admin/login) verifies the password and starts an
// admin session. Rate-limited at the route to blunt brute force.
func adminLoginHandler(c *gin.Context) {
	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	hash, found, err := getConfig(cfgKeyAdminPasswordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read admin config"})
		return
	}
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin password not set"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}
	if err := issueAdminCookie(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged in"})
}

// adminLogoutHandler (POST /admin/logout) clears the admin session cookie.
func adminLogoutHandler(c *gin.Context) {
	clearAdminCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// AdminAuthMiddleware gates admin-only data endpoints behind a valid admin session.
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdminAuthenticated(c) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "admin authentication required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// agentOnlineWindow is how long after the last heartbeat an agent is still
// considered online. Mirrors the 5-minute threshold used in the web UI.
const agentOnlineWindow = 5 * time.Minute

type adminTopologyDevice struct {
	AgentID  string    `json:"agent_id"`
	Name     string    `json:"name"`
	OS       string    `json:"os"`
	LastSeen time.Time `json:"last_seen"`
	Online   bool      `json:"online"`
	Files    int       `json:"files"`
}

type adminTopologyUser struct {
	UserID      string                `json:"user_id"`
	Name        string                `json:"name"`
	Email       string                `json:"email"`
	IsGuest     bool                  `json:"is_guest"`
	DeviceCount int                   `json:"device_count"`
	Devices     []adminTopologyDevice `json:"devices"`
}

// adminTopologyHandler (GET /admin/topology, admin-gated) returns every user
// with the devices (agents) they own, built from the users table + agentStore.
// Agents whose owner has no user row are returned under "unassigned" (legacy
// agents registered before user isolation).
func adminTopologyHandler(c *gin.Context) {
	users, err := listUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load users"})
		return
	}

	agentMu.RLock()
	byUser := make(map[string][]adminTopologyDevice)
	for _, a := range agentStore {
		byUser[a.UserID] = append(byUser[a.UserID], adminTopologyDevice{
			AgentID:  a.AgentID,
			Name:     a.Name,
			OS:       a.ClientData.SystemInfo.OS,
			LastSeen: a.LastSeen,
			Online:   time.Since(a.LastSeen) < agentOnlineWindow,
			Files:    a.FileStats.TotalFiles,
		})
	}
	agentMu.RUnlock()

	result := make([]adminTopologyUser, 0, len(users))
	for _, u := range users {
		devices := byUser[u.ID]
		delete(byUser, u.ID)
		if devices == nil {
			devices = []adminTopologyDevice{}
		}
		result = append(result, adminTopologyUser{
			UserID:      u.ID,
			Name:        u.Name,
			Email:       u.Email,
			IsGuest:     u.IsGuest,
			DeviceCount: len(devices),
			Devices:     devices,
		})
	}

	// Whatever remains in byUser has no matching user row.
	orphan := []adminTopologyDevice{}
	for _, ds := range byUser {
		orphan = append(orphan, ds...)
	}

	c.JSON(http.StatusOK, gin.H{"users": result, "unassigned": orphan})
}
