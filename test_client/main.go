package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mutex sync.Mutex

	fmt.Println("server : " + server_url)

	clientData := &ClientData{}

	// 네트워크 정보 수집 작업
	wg.Add(1)
	go collectNetworkInfo(&wg, clientData, &mutex)

	// 시스템 정보 수집 작업
	wg.Add(1)
	go collectSystemInfo(&wg, clientData, &mutex)

	// 활성화된 포트 정보 수집 작업
	wg.Add(1)
	go collectActivePorts(&wg, clientData, &mutex)

	// 서버와 연결하여 고유 코드 발급 작업
	wg.Add(1)
	go connectToServer(&wg, clientData, &mutex)

	// 모든 작업 완료 대기
	wg.Wait()

	fmt.Println("All tasks completed.")

	// 데이터 서버로 전송
	sendDataToServer(clientData)
}
