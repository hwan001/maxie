package main

import (
	"fmt"
	"fileoptimizer/client/collector"
)

func main() {
	fmt.Println("Starting client...")

	scanDir := "/Users/hwan/github/woghks7209/fileoptimizer/test_test"
	files := collector.ScanFiles(scanDir)  
  	collector.DuplicateFiles(files)

	fmt.Println("Client finished successfully.")
}
