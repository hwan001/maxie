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
