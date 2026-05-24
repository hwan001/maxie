package main

import (
	"sync"
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
	IsGuest bool   `json:"is_guest,omitempty"`
	jwt.RegisteredClaims
}

type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Picture   string     `json:"picture"`
	IsGuest   bool       `json:"is_guest,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

var jwtBlacklist = map[string]time.Time{}

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

type DriveEntry struct {
	Path        string   `json:"path"`
	DriveType   string   `json:"drive_type"` // google, naver, local, etc.
	Label       string   `json:"label"`
	ExcludeDirs []string `json:"exclude_dirs,omitempty"`
	ExcludeExts []string `json:"exclude_exts,omitempty"`
}

type FileStats struct {
	TotalFiles     int       `json:"total_files"`
	TotalSize      int64     `json:"total_size"`
	DuplicateCount int       `json:"duplicate_count"`
	LastScanned    time.Time `json:"last_scanned"`
}

type AgentRecord struct {
	AgentID             string       `json:"agent_id"`
	Name                string       `json:"name"`
	Token               string       `json:"token"`
	ServerURL           string       `json:"server_url"`
	RegisteredAt        time.Time    `json:"registered_at"`
	LastSeen            time.Time    `json:"last_seen"`
	ClientData          ClientData   `json:"client_data"`
	Drives              []DriveEntry `json:"drives"`
	FileStats           FileStats    `json:"file_stats"`
	ScanIntervalMinutes int          `json:"scan_interval_minutes"`
	// UserID is the internal user UUID that owns this agent. Empty for legacy
	// agents registered before user isolation was introduced.
	UserID string `json:"user_id,omitempty"`
	// DeletedPaths tracks drive paths explicitly removed via the web UI.
	// Heartbeats from the agent will not re-add these paths.
	DeletedPaths []string `json:"deleted_paths,omitempty"`
}

var agentStore = map[string]*AgentRecord{}
var agentMu sync.RWMutex
