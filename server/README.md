```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 로그인 처리
    r.POST("/login", func(c *gin.Context) {
        // 로그인 처리 로직
        c.JSON(200, gin.H{"message": "logged in"})
    })

    // 클라이언트 데이터 수신
    r.POST("/client-data", func(c *gin.Context) {
        // 클라이언트로부터 데이터 수신 처리
        c.JSON(200, gin.H{"message": "data received"})
    })

    r.Run()
}
```

```go
package main

import (
    "net/http"
    "log"
)

func downloadClient(w http.ResponseWriter, r *http.Request) {
    // 클라이언트 파일의 경로를 지정
    http.ServeFile(w, r, "./client/client")
}

func main() {
    http.HandleFunc("/download/client", downloadClient)
    log.Println("Server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```


GOOS=linux GOARCH=amd64 go build -o server .      