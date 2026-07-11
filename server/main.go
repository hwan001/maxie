package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	loadAgents()
	if err := initFileDB(); err != nil {
		log.Printf("filedb init failed: %v", err)
	}
	go runCleanupTask()

	sessionSecret := envOrDefault("SESSION_SECRET", "dev-session-secret-change-me")

	router := gin.Default()
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

	router.POST("/auth/google", handleGoogleAuth)
	router.POST("/auth/guest", handleGuestAuth)

	// Agent-auth routes (X-Agent-Token, no JWT required)
	router.POST("/agent/register", registerAgent)
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
