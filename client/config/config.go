package config

import(
	"fileoptimizer/common"
)

func LoadConfig() (*common.Config, error) {
	// AES 키는 16, 24, 또는 32 바이트여야 함
	return &common.Config{
		ServerURL:     "https://example.com",
		Credentials:   common.Credentials{Username: "user", Password: "pass"},
		ScanDir:       "/path/to/scan",
		EncryptionKey: []byte("thisis16byteskey"),  // 16바이트 길이의 키
	}, nil
}