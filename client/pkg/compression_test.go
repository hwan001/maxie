package pkg

import (
	"testing"
	"fileoptimizer/common"
)

func TestCompressData(t *testing.T) {
	// Mock system and file info
	systemInfo := "mock system info"
	fileInfo := []*common.FileInfo{
		{Path: "/path/to/file1", Hash: "12345"},
		{Path: "/path/to/file2", Hash: "67890"},
	}

	// Compress the data
	compressedData, err := CompressData(systemInfo, fileInfo)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that compressed data is non-empty
	if len(compressedData) == 0 {
		t.Fatalf("Expected non-empty compressed data")
	}

	// Test decompression
	decompressedData, err := DecompressData(compressedData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify decompressed data is non-empty
	if len(decompressedData) == 0 {
		t.Fatalf("Expected non-empty decompressed data")
	}
}
