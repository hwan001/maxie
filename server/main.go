package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	sessionSecret := envOrDefault("SESSION_SECRET", devSessionSecret)
	if err := validateSecrets(sessionSecret); err != nil {
		log.Fatalf("startup aborted: %v", err)
	}

	loadAgents()
	if err := initFileDB(); err != nil {
		log.Printf("filedb init failed: %v", err)
	}
	go runCleanupTask()

	router := gin.New()
	router.Use(gin.Recovery(), requestLogger())
	store := cookie.NewStore([]byte(sessionSecret))
	router.Use(sessions.Sessions("session", store))

	corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Agent-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health/readiness probes — public, unlogged, no auth.
	router.GET("/healthz", healthzHandler)
	router.GET("/readyz", readyzHandler)

	// Shared per-IP budget for session/agent creation: ~10 burst then refill one
	// every 6s (~10/min sustained). Guards against credential-stuffing and
	// registration abuse. Token-authed high-frequency routes (heartbeat, file
	// push) are intentionally left ungated to avoid throttling many agents that
	// share one NAT IP.
	authLimiter := rateLimit(rate.Every(6*time.Second), 10)

	router.POST("/auth/google", authLimiter, handleGoogleAuth)
	router.POST("/auth/guest", authLimiter, handleGuestAuth)

	// Admin console — protected by its own password (separate from user auth).
	// Status is public so the UI can pick setup vs login; login and password
	// changes are rate-limited to blunt brute force.
	router.GET("/admin/status", adminStatusHandler)
	router.POST("/admin/login", authLimiter, adminLoginHandler)
	router.POST("/admin/logout", adminLogoutHandler)
	router.POST("/admin/password", authLimiter, adminSetPasswordHandler)

	// Agent-auth routes (X-Agent-Token, no JWT required)
	router.POST("/agent/register", authLimiter, registerAgent)
	router.POST("/agent/heartbeat", agentHeartbeat)
	router.POST("/agent/drives", agentUpdateDrives)
	router.POST("/agent/files", pushAgentFiles)
	router.GET("/agent/pending-actions", getPendingActionsHandler)
	router.POST("/agent/confirm-action", confirmActionHandler)
	router.PUT("/agent/user", linkAgentUser)

	protected := router.Group("/protected")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/", protectedEndpoint)
		protected.GET("/profile", getProfile)
		protected.POST("/logout", logout)

		// Web UI routes (JWT cookie required)
		protected.GET("/agents", listAgents)
		protected.DELETE("/agents/:id", deleteAgent)
		protected.PUT("/agents/:id/drives", updateAgentDrives)
		protected.PUT("/agents/:id/config", updateAgentConfig)
		protected.GET("/files", listFilesHandler)
		protected.GET("/files/duplicates", getDuplicatesHandler)
		protected.DELETE("/files", deleteFileHandler)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":50000"
	} else if port[0] != ':' {
		port = ":" + port
	}
	router.Run(port)
}
