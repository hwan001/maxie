package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Path string
	Hash string
}

var ignoreExtensions = []string{".pdb", ".dll", ".exe", ".user", ".suo"}
var ignoreDirs = map[string]bool{
    "bin":  true,
    "obj":  true,
    ".vs":  true,
    "Debug": true,
    "Release": true,
	".git": true,
	"node_modules": true,
}

func CheckExtension(filePath string) bool {
    for _, ext := range ignoreExtensions {
        if strings.HasSuffix(filePath, ext) {
            return true
        }
    }
    return false
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

func ScanFiles(root string) []*FileInfo {
	var files []*FileInfo
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && ignoreDirs[d.Name()] {
			fmt.Printf("Skip(dir) : %s\n", path)
			return filepath.SkipDir
		}

        if !d.IsDir() && CheckExtension(path) {
            fmt.Printf("Skip(file): %s\n", path)
            return nil
        }

		if !d.IsDir() {
			hash := hashFile(path)
			// fmt.Printf("%s\n", path)
			files = append(files, &FileInfo{Path: path, Hash: hash})
		}

		return nil
	})
	if err != nil {
		fmt.Println("Error scanning files:", err)
	}
	return files
}

func DuplicateFiles(files []*FileInfo) map[string][]string {
	fileMap := make(map[string][]string)
	duplicateMap := make(map[string][]string)
	
	for _, file := range files {
		fileMap[file.Hash] = append(fileMap[file.Hash], file.Path)
	}

	for hash, paths := range fileMap {
		fmt.Printf("Hash: %s\n", hash)

		for i, path := range paths {
			if len(paths) > 1 {
				fmt.Printf("  path(%d) : %s \n", i, path)
				duplicateMap[hash] = paths
			}
		}
	}

	return duplicateMap
}


func collectNetworkInfo(wg *sync.WaitGroup, data *ClientData, mutex *sync.Mutex) {
	defer wg.Done()

	fmt.Println("Collecting network interfaces...")

	files := ScanFiles("root/path/")
	filescanningdata := DuplicateFiles(files) // 중복 파일 데이터 생성

	mutex.Lock()
	data.NetworkInterfaces = filescanningdata
	mutex.Unlock()
}
