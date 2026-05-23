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

	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
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

	api := "/api"

	router.POST(api+"/auth/google", handleGoogleAuth)

	router.POST(api+"/agent/register", registerAgent)
	router.GET(api+"/agents", listAgents)
	router.PUT(api+"/agents/:id/drives", updateAgentDrives)
	router.PUT(api+"/agents/:id/config", updateAgentConfig)
	router.POST(api+"/agent/heartbeat", agentHeartbeat)
	router.POST(api+"/agent/drives", agentUpdateDrives)
	router.POST(api+"/agent/files", pushAgentFiles)
	router.GET(api+"/agent/pending-actions", getPendingActionsHandler)
	router.POST(api+"/agent/confirm-action", confirmActionHandler)

	router.GET(api+"/files", listFilesHandler)
	router.GET(api+"/files/duplicates", getDuplicatesHandler)
	router.DELETE(api+"/files", deleteFileHandler)

	protected := router.Group(api + "/protected")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/", protectedEndpoint)
		protected.GET("/profile", getProfile)
		protected.POST("/logout", logout)
	}

	router.Run(":50000")
}
