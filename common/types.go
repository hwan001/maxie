package common

type FileInfo struct {
	Path string
	Hash string
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type Config struct {
	ServerURL      string
	Credentials    Credentials
	ScanDir        string
	EncryptionKey  []byte
}