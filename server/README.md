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

scp server user@t8p:/home/user/server

go run .


curl -i -X POST http://localhost:8080/api/auth/google \
-H "Content-Type: application/x-www-form-urlencoded" \
-d "code=코드" -> 이건 리액트 로그로 확인

curl -i -X POST http://localhost:8080/api/auth/google \
-H "Content-Type: application/json" \
-d '{"code": "4/0AQlEd8zxBH8Sa2A0lDAMKpN_rSzAbkHahjCWWbH-1uAlfHVYdM12D104m6i8KBODwnqR_A"}'
