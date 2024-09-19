package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "log"
	"fmt"
	"time"
	"fileoptimizer/common"
)

func getClientID(c *gin.Context) {
    clientID := "client-" + fmt.Sprintf("%d", time.Now().UnixNano())
    c.JSON(http.StatusOK, gin.H{"client_id": clientID})
}

func receiveClientData(c *gin.Context) {
    var data common.ClientData

    if err := c.ShouldBindJSON(&data); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    log.Printf("Received data from client %s\n", data.ClientID)
    log.Printf("Network Interfaces: %+v\n", data.NetworkInterfaces)
    log.Printf("System Info: %+v\n", data.SystemInfo)
    log.Printf("Active Ports: %+v\n", data.ActivePorts)

    c.JSON(http.StatusOK, gin.H{"message": "data received"})
}

func login(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"message": "logged in"})
}

func main() {
    router := gin.Default()
	
    router.GET("/get-client-id", getClientID)
    router.POST("/client-data", receiveClientData)

    router.GET("/download/client-mac", func(c *gin.Context) {
        c.File("/home/user/client-mac")
    })
	router.GET("/download/client-linux", func(c *gin.Context) {
        c.File("/home/user/client-linux")
    })
	router.GET("/download/client-windows", func(c *gin.Context) {
        c.File("/home/user/client-windows")
    })
	
    router.POST("/login", login)
    
    router.Run(":8080")
}
