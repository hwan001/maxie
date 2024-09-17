package handler

import (
	"os"
)

func MoveFile(srcPath, destPath string) error {
	return os.Rename(srcPath, destPath)
}

func DeleteFile(path string) error {
	return os.Remove(path)
}
