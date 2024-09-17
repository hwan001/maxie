package collector

import (
	"crypto/sha256"
	"fmt"
	"fileoptimizer/common" // common 패키지를 사용
	"io"
	"os"
	"path/filepath"
)

func ScanFiles(directory string) []*common.FileInfo { // common.FileInfo를 사용
	var files []*common.FileInfo
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			hash := hashFile(path)
			files = append(files, &common.FileInfo{Path: path, Hash: hash})
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error scanning files:", err)
	}
	return files
}

func hashFile(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func RemoveDuplicateFiles(files []*common.FileInfo) {
	fileMap := make(map[string][]string)

	// 파일을 해시값 기준으로 그룹화
	for _, file := range files {
		fileMap[file.Hash] = append(fileMap[file.Hash], file.Path)
	}

	// 각 그룹에서 중복 파일 삭제
	for hash, paths := range fileMap {
		if len(paths) > 1 {
			// 첫 번째 파일을 제외하고 나머지 파일 삭제
			fmt.Printf("Hash: %s\n", hash)
			for i, path := range paths {
				if i == 0 {
					// 첫 번째 파일은 유지
					fmt.Printf(" Keeping file: %s\n", path)
				} else {
					// 중복 파일 삭제
					err := os.Remove(path)
					if err != nil {
						fmt.Printf(" Failed to delete file: %s, error: %v\n", path, err)
					} else {
						fmt.Printf(" Deleted file: %s\n", path)
					}
				}
			}
		}
	}
}
