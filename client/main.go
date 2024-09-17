package main

import (
	"fmt"
	//"log"
	//"fileoptimizer/client/auth"
	"fileoptimizer/client/collector"
	//"fileoptimizer/client/pkg"
	//"fileoptimizer/client/config"
)

func main() {
	fmt.Println("Starting client...")

	// 스캔할 디렉토리 설정 (예시로 /path/to/scan 디렉토리를 스캔)
	scanDir := "/Users/hwan/github/woghks7209/fileoptimizer/test"

	// 파일 스캔
	files := collector.ScanFiles(scanDir)
  fmt.Println(files)
  
	// 중복 파일 제거
  //collector.RemoveDuplicateFiles(files)

  // Collect system information
	systemInfo := collector.CollectSystemInfo()

  fmt.Println(systemInfo)

  //collector.CollectNetworInfo()
  collector.DefaultNetworkInfo()

	//  // Load config
	//  config, err := config.LoadConfig()
	//  if err != nil {
	//  	log.Fatalf("Failed to load config: %v", err)
	//  }
	//
	//  // Authenticate with server
	//  token, err := auth.Authenticate(config.ServerURL, config.Credentials)  // config.Credentials로 수정
	//  if err != nil {
	//  	log.Fatalf("Authentication failed: %v", err)
	//  }
	//
	//  // Collect system information
	//  systemInfo := collector.CollectSystemInfo()
	//  fileInfo := collector.ScanFiles(config.ScanDir)
	//
	//  // Compress the data
	//  data, err := pkg.CompressData(systemInfo, fileInfo)
	//  if err != nil {
	//  	log.Fatalf("Failed to compress data: %v", err)
	//  }
	//
	//  // **Encryption key length validation**
	//  if len(config.EncryptionKey) != 16 && len(config.EncryptionKey) != 24 && len(config.EncryptionKey) != 32 {
	//  	log.Fatalf("Invalid encryption key size: %d. Must be 16, 24, or 32 bytes", len(config.EncryptionKey))
	//  }
	//
	//  // Encrypt the data
	//  encryptedData, err := pkg.EncryptData(data, config.EncryptionKey)
	//  if err != nil {
	//  	log.Fatalf("Failed to encrypt data: %v", err)
	//  }
	//
	//  // Send data to server
	//  err = pkg.SendData(config.ServerURL, encryptedData, token)
	//  if err != nil {
	//  	log.Fatalf("Failed to send data: %v", err)
	//  }

	fmt.Println("Client finished successfully.")
}
