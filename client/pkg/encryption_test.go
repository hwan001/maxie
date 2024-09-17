package pkg

import (
	"testing"
)

func TestEncryptData(t *testing.T) {
	// Test data and key
	data := []byte("test data")
	key := []byte("averysecretkey!!")

	// Encrypt the data
	encryptedData, err := EncryptData(data, key)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	
	// Ensure that the encrypted data is not empty
	if encryptedData == "" {
		t.Fatalf("Expected non-empty encrypted data")
	}
}
