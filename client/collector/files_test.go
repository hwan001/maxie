package collector

import (
	"os"
	"testing"
)

func TestScanFiles(t *testing.T) {
	// Create mock files for testing
	tmpDir := t.TempDir()
	tmpFile1 := tmpDir + "/file1.txt"
	tmpFile2 := tmpDir + "/file2.txt"

	_ = os.WriteFile(tmpFile1, []byte("hello world"), 0644)
	_ = os.WriteFile(tmpFile2, []byte("hello Go"), 0644)

	// Scan the files in the directory
	files := ScanFiles(tmpDir)

	// Check that the correct number of files are scanned
	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}

	// Check that file paths are correct
	expectedFilePaths := []string{tmpFile1, tmpFile2}
	for i, fileInfo := range files {
		if fileInfo.Path != expectedFilePaths[i] {
			t.Errorf("Expected %v, got %v", expectedFilePaths[i], fileInfo.Path)
		}
	}
}
