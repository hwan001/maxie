package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func handleGoogleAuth(c *gin.Context) {
	var request struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request, code not provided"})
		return
	}
	if request.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not provided"})
		return
	}

	ctx := context.Background()
	token, err := googleOAuthConfig.Exchange(ctx, request.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to exchange token (%s)", err)})
		return
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token found in token response"})
		return
	}

	tokenInfoURL := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	tokenInfoResp, err := http.Get(tokenInfoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get token info"})
		return
	}
	defer tokenInfoResp.Body.Close()

	var profile map[string]interface{}
	if err := json.NewDecoder(tokenInfoResp.Body).Decode(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode token info"})
		return
	}

	sub, _ := profile["sub"].(string)
	name, _ := profile["name"].(string)
	email, _ := profile["email"].(string)
	picture, _ := profile["picture"].(string)

	user, err := upsertOAuthUser("google", sub, name, email, picture)
	if err != nil {
		log.Printf("upsertOAuthUser: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store user"})
		return
	}

	claims := Claims{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Picture: user.Picture,
		IsGuest: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "jwt",
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   cookieDomain != "",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   cookieDomain,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logged in", "profile": user})
}

func handleGuestAuth(c *gin.Context) {
	user, err := createGuestUser()
	if err != nil {
		log.Printf("createGuestUser: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guest session"})
		return
	}

	expiry := *user.ExpiresAt

	claims := Claims{
		ID:      user.ID,
		Name:    "Guest",
		IsGuest: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "jwt",
		Value:    jwtToken,
		Expires:  expiry,
		HttpOnly: true,
		Secure:   cookieDomain != "",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   cookieDomain,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Guest session created", "profile": user})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("jwt")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie not found"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}

func protectedEndpoint(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"message": "Protected endpoint", "user": user})
}

func getProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User information not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": user})
}

func logout(c *gin.Context) {
	token, err := c.Cookie("jwt")
	if err == nil {
		jwtBlacklist[token] = time.Now().Add(time.Hour * 24)
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieDomain != "",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   cookieDomain,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func registerAgent(c *gin.Context) {
	var req struct {
		Name      string     `json:"name"`
		ServerURL string     `json:"server_url"`
		Data      ClientData `json:"client_data"`
		UserID    string     `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	agentID := "agent-" + fmt.Sprintf("%d", time.Now().UnixNano())
	token := generateToken()

	rec := &AgentRecord{
		AgentID:      agentID,
		Name:         req.Name,
		Token:        token,
		ServerURL:    req.ServerURL,
		RegisteredAt: time.Now(),
		LastSeen:     time.Now(),
		ClientData:   req.Data,
		UserID:       req.UserID,
	}

	agentMu.Lock()
	agentStore[agentID] = rec
	agentMu.Unlock()

	go saveAgents()
	log.Printf("Agent registered: %s (%s) user=%s", req.Name, agentID, req.UserID)
	c.JSON(http.StatusOK, gin.H{"agent_id": agentID, "token": token})
}

// getUserAgentIDs returns all agent IDs owned by the given user.
// Must NOT be called while agentMu is held.
func getUserAgentIDs(userID string) []string {
	agentMu.RLock()
	defer agentMu.RUnlock()
	var ids []string
	for id, a := range agentStore {
		if a.UserID == userID {
			ids = append(ids, id)
		}
	}
	return ids
}

func listAgents(c *gin.Context) {
	claims := c.MustGet("user").(*Claims)

	agentMu.RLock()
	defer agentMu.RUnlock()

	list := make([]*AgentRecord, 0)
	for _, a := range agentStore {
		if a.UserID == claims.ID {
			list = append(list, a)
		}
	}
	c.JSON(http.StatusOK, gin.H{"agents": list})
}

func deleteAgent(c *gin.Context) {
	claims := c.MustGet("user").(*Claims)
	agentID := c.Param("id")

	agentMu.Lock()
	agent, ok := agentStore[agentID]
	if !ok {
		agentMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	if agent.UserID != claims.ID {
		agentMu.Unlock()
		c.JSON(http.StatusForbidden, gin.H{"error": "not your agent"})
		return
	}
	delete(agentStore, agentID)
	agentMu.Unlock()

	go saveAgents()
	if _, err := fileDB.Exec(`DELETE FROM files WHERE agent_id = ?`, agentID); err != nil {
		log.Printf("deleteAgent: failed to delete files for %s: %v", agentID, err)
	}
	if _, err := fileDB.Exec(`DELETE FROM pending_actions WHERE agent_id = ?`, agentID); err != nil {
		log.Printf("deleteAgent: failed to delete pending_actions for %s: %v", agentID, err)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func agentHeartbeat(c *gin.Context) {
	token := c.GetHeader("X-Agent-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	var req struct {
		ClientData          ClientData   `json:"client_data"`
		Drives              []DriveEntry `json:"drives"`
		FileStats           FileStats    `json:"file_stats"`
		ScanIntervalMinutes int          `json:"scan_interval_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentMu.Lock()
	defer agentMu.Unlock()

	for _, a := range agentStore {
		if a.Token == token {
			a.ClientData = req.ClientData

			// Build a set of explicitly-deleted paths so agent heartbeats
				// cannot silently re-add drives removed via the web UI.
				deletedSet := make(map[string]struct{}, len(a.DeletedPaths))
				for _, p := range a.DeletedPaths {
					deletedSet[p] = struct{}{}
				}

				// Start from the current server drive list and update metadata.
				merged := make([]DriveEntry, len(a.Drives))
				copy(merged, a.Drives)
				for i, serverDrive := range merged {
					for _, agentDrive := range req.Drives {
						if serverDrive.Path == agentDrive.Path {
							merged[i].Label = agentDrive.Label
							merged[i].DriveType = agentDrive.DriveType
							break
						}
					}
				}

				// Agent may report paths not yet on the server — add them unless
				// they were explicitly removed through the web UI.
				serverPathSet := make(map[string]struct{}, len(a.Drives))
				for _, d := range a.Drives {
					serverPathSet[d.Path] = struct{}{}
				}
				for _, agentDrive := range req.Drives {
					if _, onServer := serverPathSet[agentDrive.Path]; onServer {
						continue // already handled above
					}
					if _, deleted := deletedSet[agentDrive.Path]; deleted {
						continue // explicitly removed — do not re-add
					}
					merged = append(merged, agentDrive)
				}

			a.Drives = merged
			a.FileStats = req.FileStats
			// Update scan interval only if agent sends a non-zero value
			// (server's configured value takes precedence if different)
			if req.ScanIntervalMinutes > 0 && a.ScanIntervalMinutes == 0 {
				a.ScanIntervalMinutes = req.ScanIntervalMinutes
			}
			a.LastSeen = time.Now()
			go saveAgents()
			c.JSON(http.StatusOK, gin.H{
				"ok":                    true,
				"drives":                merged,
				"scan_interval_minutes": a.ScanIntervalMinutes,
			})
			return
		}
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
}

func updateAgentConfig(c *gin.Context) {
	claims := c.MustGet("user").(*Claims)
	agentID := c.Param("id")
	var req struct {
		ScanIntervalMinutes int `json:"scan_interval_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentMu.Lock()
	defer agentMu.Unlock()

	agent, ok := agentStore[agentID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	if agent.UserID != claims.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if req.ScanIntervalMinutes > 0 {
		agent.ScanIntervalMinutes = req.ScanIntervalMinutes
	}
	go saveAgents()
	c.JSON(http.StatusOK, gin.H{"ok": true, "scan_interval_minutes": agent.ScanIntervalMinutes})
}

func pushAgentFiles(c *gin.Context) {
	token := c.GetHeader("X-Agent-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	agentMu.RLock()
	var agentID string
	for id, a := range agentStore {
		if a.Token == token {
			agentID = id
			break
		}
	}
	agentMu.RUnlock()

	if agentID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req struct {
		Files []AgentFileRecord `json:"files"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := upsertFileBatch(agentID, req.Files); err != nil {
		log.Printf("upsertFileBatch failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store files"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "count": len(req.Files)})
}

func listFilesHandler(c *gin.Context) {
	claims := c.MustGet("user").(*Claims)
	userAgentIDs := getUserAgentIDs(claims.ID)

	// No agents linked to this user → nothing to show.
	if len(userAgentIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"files": []interface{}{}, "total": 0, "page": 1, "limit": 50})
		return
	}

	requestedAgentID := c.Query("agent_id")
	if requestedAgentID != "" {
		// Verify the requested agent belongs to the authenticated user.
		allowed := false
		for _, id := range userAgentIDs {
			if id == requestedAgentID {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	q := FileQuery{
		AgentID:   requestedAgentID,
		AgentIDs:  userAgentIDs,
		DriveType: c.Query("drive_type"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortDir:   c.Query("sort_dir"),
	}
	fmt.Sscanf(c.Query("page"), "%d", &q.Page)
	fmt.Sscanf(c.Query("limit"), "%d", &q.Limit)

	result, err := queryFiles(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	agentMu.RLock()
	nameMap := map[string]string{}
	for id, a := range agentStore {
		nameMap[id] = a.Name
	}
	agentMu.RUnlock()

	for i := range result.Files {
		result.Files[i].AgentName = nameMap[result.Files[i].AgentID]
	}

	c.JSON(http.StatusOK, gin.H{
		"files": result.Files,
		"total": result.Total,
		"page":  q.Page,
		"limit": q.Limit,
	})
}

func getDuplicatesHandler(c *gin.Context) {
	claims := c.MustGet("user").(*Claims)
	userAgentIDs := getUserAgentIDs(claims.ID)

	// No agents linked to this user → nothing to show.
	if len(userAgentIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"groups": []interface{}{}})
		return
	}

	requestedAgentID := c.Query("agent_id")
	if requestedAgentID != "" {
		allowed := false
		for _, id := range userAgentIDs {
			if id == requestedAgentID {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	groups, err := getDuplicateGroups(requestedAgentID, userAgentIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	agentMu.RLock()
	nameMap := map[string]string{}
	for id, a := range agentStore {
		nameMap[id] = a.Name
	}
	agentMu.RUnlock()

	for gi := range groups {
		for fi := range groups[gi].Files {
			groups[gi].Files[fi].AgentName = nameMap[groups[gi].Files[fi].AgentID]
		}
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func deleteFileHandler(c *gin.Context) {
	claims := c.MustGet("user").(*Claims)

	var req struct {
		AgentID  string `json:"agent_id"`
		FullPath string `json:"fullpath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.AgentID == "" || req.FullPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id and fullpath required"})
		return
	}

	// Verify the agent belongs to the authenticated user.
	agentMu.RLock()
	a, ok := agentStore[req.AgentID]
	agentMu.RUnlock()
	if !ok || a.UserID != claims.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := addPendingAction(req.AgentID, req.FullPath, "delete"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func getPendingActionsHandler(c *gin.Context) {
	token := c.GetHeader("X-Agent-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	agentMu.RLock()
	var agentID string
	for id, a := range agentStore {
		if a.Token == token {
			agentID = id
			break
		}
	}
	agentMu.RUnlock()

	if agentID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	actions, err := getPendingActions(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"actions": actions})
}

func confirmActionHandler(c *gin.Context) {
	token := c.GetHeader("X-Agent-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	agentMu.RLock()
	var agentID string
	for id, a := range agentStore {
		if a.Token == token {
			agentID = id
			break
		}
	}
	agentMu.RUnlock()

	if agentID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var req struct {
		ID       int64  `json:"id"`
		FullPath string `json:"fullpath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := clearPendingAction(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	deleteFileRecord(agentID, req.FullPath)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// updateAgentDrives performs a full replacement of the agent's drive list (web UI).
// Drives present in the old list but absent from the new list have their
// associated file records deleted from the database synchronously before responding,
// so the next page load reflects the removal immediately.
func updateAgentDrives(c *gin.Context) {
	claims := c.MustGet("user").(*Claims)
	agentID := c.Param("id")
	var req struct {
		Drives []DriveEntry `json:"drives"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentMu.Lock()

	agent, ok := agentStore[agentID]
	if !ok {
		agentMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	if agent.UserID != claims.ID {
		agentMu.Unlock()
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Build a set of new paths for O(1) lookup.
	newPathSet := make(map[string]struct{}, len(req.Drives))
	for _, d := range req.Drives {
		newPathSet[d.Path] = struct{}{}
	}

	// Determine which paths are being removed.
	oldDrives := agent.Drives
	deletedSet := make(map[string]struct{}, len(agent.DeletedPaths))
	for _, p := range agent.DeletedPaths {
		deletedSet[p] = struct{}{}
	}

	var removedPaths []string
	for _, oldDrive := range oldDrives {
		if _, kept := newPathSet[oldDrive.Path]; !kept {
			// Drive removed via web — record in DeletedPaths.
			deletedSet[oldDrive.Path] = struct{}{}
			removedPaths = append(removedPaths, oldDrive.Path)
		}
	}

	// If a previously-deleted path is being re-added explicitly via web, un-delete it.
	for _, d := range req.Drives {
		delete(deletedSet, d.Path)
	}

	// Persist the updated deleted-paths list.
	agent.DeletedPaths = make([]string, 0, len(deletedSet))
	for p := range deletedSet {
		agent.DeletedPaths = append(agent.DeletedPaths, p)
	}

	// Full replacement. For existing paths, preserve server-side excludes unless
	// the request explicitly provides them.
	newDrives := make([]DriveEntry, len(req.Drives))
	for i, d := range req.Drives {
		newDrives[i] = d
		for _, old := range oldDrives {
			if old.Path == d.Path {
				if len(d.ExcludeDirs) == 0 {
					newDrives[i].ExcludeDirs = old.ExcludeDirs
				}
				if len(d.ExcludeExts) == 0 {
					newDrives[i].ExcludeExts = old.ExcludeExts
				}
				break
			}
		}
	}

	agent.Drives = newDrives
	agentMu.Unlock()
	go saveAgents()

	// Delete file records for removed drives synchronously so the next page
	// load reflects the removal without a race window.
	for _, path := range removedPaths {
		if err := deleteFilesForDrivePrefix(agentID, path); err != nil {
			log.Printf("deleteFilesForDrivePrefix(%s, %s): %v", agentID, path, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "drives": newDrives})
}

// agentUpdateDrives is called by the agent itself (X-Agent-Token auth) when
// its local drive list changes (drive added or removed via the tray menu).
// Drives absent from the agent's report are removed from the server's list
// and their file records are deleted synchronously.
func agentUpdateDrives(c *gin.Context) {
	token := c.GetHeader("X-Agent-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	var req struct {
		Drives []DriveEntry `json:"drives"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentMu.Lock()
	var agentID string
	var agent *AgentRecord
	for id, a := range agentStore {
		if a.Token == token {
			agentID = id
			agent = a
			break
		}
	}
	if agent == nil {
		agentMu.Unlock()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Build set of new paths.
	newPathSet := make(map[string]struct{}, len(req.Drives))
	for _, d := range req.Drives {
		newPathSet[d.Path] = struct{}{}
	}

	// Detect drives the agent has removed.
	var removedPaths []string
	for _, oldDrive := range agent.Drives {
		if _, kept := newPathSet[oldDrive.Path]; !kept {
			removedPaths = append(removedPaths, oldDrive.Path)
		}
	}

	// Full replacement — preserve server-side excludes for existing drives.
	oldDrives := agent.Drives
	newDrives := make([]DriveEntry, len(req.Drives))
	for i, d := range req.Drives {
		newDrives[i] = d
		for _, old := range oldDrives {
			if old.Path == d.Path {
				if len(d.ExcludeDirs) == 0 {
					newDrives[i].ExcludeDirs = old.ExcludeDirs
				}
				if len(d.ExcludeExts) == 0 {
					newDrives[i].ExcludeExts = old.ExcludeExts
				}
				break
			}
		}
	}

	agent.Drives = newDrives
	capturedAgentID := agentID
	agentMu.Unlock()
	go saveAgents()

	// Synchronously delete file records for removed drives.
	for _, path := range removedPaths {
		if err := deleteFilesForDrivePrefix(capturedAgentID, path); err != nil {
			log.Printf("agentUpdateDrives deleteFilesForDrivePrefix(%s, %s): %v", capturedAgentID, path, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "drives": newDrives})
}

func generateToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateUUID returns a random UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("generateUUID: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// runCleanupTask runs on a ticker and removes expired guest users with all
// their agents and file records. Intended to be started as a goroutine.
func runCleanupTask() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		cleanupExpiredGuests()
	}
}

func agentsFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".fileoptimizer", "agents.json")
}

func loadAgents() {
	data, err := os.ReadFile(agentsFilePath())
	if err != nil {
		return
	}
	var store map[string]*AgentRecord
	if err := json.Unmarshal(data, &store); err != nil {
		log.Printf("failed to load agents from disk: %v", err)
		return
	}
	agentMu.Lock()
	agentStore = store
	agentMu.Unlock()
	log.Printf("loaded %d agent(s) from disk", len(store))
}

func saveAgents() {
	agentMu.RLock()
	data, err := json.MarshalIndent(agentStore, "", "  ")
	agentMu.RUnlock()
	if err != nil {
		return
	}
	path := agentsFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		log.Printf("failed to save agents: %v", err)
	}
}

// linkAgentUser links an already-registered agent to a user account.
// The agent authenticates with its token via X-Agent-Token header.
// Body: { "user_id": "<uuid>" }
func linkAgentUser(c *gin.Context) {
	token := c.GetHeader("X-Agent-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	agentMu.Lock()
	defer agentMu.Unlock()

	for _, a := range agentStore {
		if a.Token == token {
			a.UserID = req.UserID
			go saveAgents()
			log.Printf("agent %s linked to user %s", a.AgentID, req.UserID)
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
}
