package pkg

import (
	"errors"
	"bytes"
	"net/http"
	"time"
)

// 서버와 연결 시도
func connectToServer(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Attempting to connect to server...")

	// 예시로 간단한 네트워크 연결 시도
	conn, err := net.DialTimeout("tcp", "server-address:port", time.Second*5)
	if err != nil {
		fmt.Println("Failed to connect to server")
	} else {
		fmt.Println("Connected to server")
		defer conn.Close()
	}

	// 서버에서 클라이언트별 고유 코드 발급 로직
	clientID := "unique-client-id"
	fmt.Printf("Received Client ID: %s\n", clientID)
}
