package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("session", store))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://goserver.666lab.org"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	default_endpoint := "/api"

	// POST
	router.POST(default_endpoint+"/auth/google", handleGoogleAuth)
	router.POST(default_endpoint+"/client-data", receiveClientData)

	// GET
	router.GET(default_endpoint+"/get-client-id", getClientID)
	router.GET(default_endpoint+"/data", getData)
	router.GET(default_endpoint+"/download/client/mac", func(c *gin.Context) {
		log.Println("Request for mac client")
		c.FileAttachment("./client/client-mac", "client")
	})
	router.GET(default_endpoint+"/download/client/linux", func(c *gin.Context) {
		log.Println("Request for linux client")
		c.FileAttachment("./client/client-linux", "client")
	})
	router.GET(default_endpoint+"/download/client/windows", func(c *gin.Context) {
		log.Println("Request for windows client")
		c.FileAttachment("./client/client-windows", "client.exe")
	})

	protected := router.Group(default_endpoint + "/protected")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/", protectedEndpoint)
		protected.GET("/profile", getProfile)
		protected.POST("/logout", logout)
	}

	// 관리자만 접근 가능한 경로 그룹
	// admin := router.Group(default_endpoint + "/admin")
	// admin.Use(AuthMiddleware(), AdminMiddleware()) // 두 개의 미들웨어 적용
	// {
	//     admin.GET("/dashboard", getAdminDashboard)
	//     admin.POST("/user", createUser)
	// }

	router.Run(":50000")
}
