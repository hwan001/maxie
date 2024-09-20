package main


import (
	"time"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuthConfig = &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	RedirectURL:  redirectURI,
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

type Claims struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

// 샘플 사용자 구조체
type User struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Picture  string `json:"picture"`
}

// 사용자 정보 저장소 (예: 메모리 기반 저장소)
var users = map[string]User{}

// JWT 블랙리스트 (토큰 무효화를 위해 사용)
var jwtBlacklist = map[string]time.Time{}


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


// ClientData 구조체는 클라이언트 측과 동일하게 정의
type ClientData struct {
    ClientID          string             `json:"client_id"`
    NetworkInterfaces []NetworkInterface `json:"network_interfaces"`
    SystemInfo        SystemInfo         `json:"system_info"`
    ActivePorts       []ActivePort       `json:"active_ports"`
}

type NetworkInterface struct {
    Name string `json:"name"`
    IP   string `json:"ip"`
}

type SystemInfo struct {
    OS              string    `json:"os"`
    Platform        string    `json:"platform"`
    PlatformVersion string    `json:"platform_version"`
    CPUs            []CPUInfo `json:"cpus"`
}

type CPUInfo struct {
    ModelName string  `json:"model_name"`
    Cores     int     `json:"cores"`
    SpeedGHz  float64 `json:"speed_ghz"`
}

type ActivePort struct {
    Port         int    `json:"port"`
    LocalAddress string `json:"local_address"`
}
