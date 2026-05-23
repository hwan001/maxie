package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	netstat "github.com/shirou/gopsutil/net"
)

type NetworkInterface struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type CPUInfo struct {
	ModelName string  `json:"model_name"`
	Cores     int32   `json:"cores"`
	SpeedGHz  float64 `json:"speed_ghz"`
}

type SystemInfo struct {
	OS              string    `json:"os"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	CPUs            []CPUInfo `json:"cpus"`
}

type ActivePort struct {
	Port         int    `json:"port"`
	LocalAddress string `json:"local_address"`
}

type AgentData struct {
	ClientID          string             `json:"client_id"`
	NetworkInterfaces []NetworkInterface `json:"network_interfaces"`
	SystemInfo        SystemInfo         `json:"system_info"`
	ActivePorts       []ActivePort       `json:"active_ports"`
}

type RegisterRequest struct {
	Name      string    `json:"name"`
	ServerURL string    `json:"server_url"`
	Data      AgentData `json:"client_data"`
}

type RegisterResponse struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

type HeartbeatRequest struct {
	ClientData          AgentData    `json:"client_data"`
	Drives              []DriveEntry `json:"drives"`
	FileStats           FileStats    `json:"file_stats"`
	ScanIntervalMinutes int          `json:"scan_interval_minutes"`
}

func collectData() *AgentData {
	data := &AgentData{}
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()
		hostname, _ := os.Hostname()
		cpus, _ := cpu.Info()
		var cpuInfos []CPUInfo
		for _, c := range cpus {
			cpuInfos = append(cpuInfos, CPUInfo{
				ModelName: c.ModelName,
				Cores:     c.Cores,
				SpeedGHz:  c.Mhz / 1000,
			})
		}
		mu.Lock()
		data.ClientID = hostname
		data.SystemInfo = SystemInfo{
			OS:       runtime.GOOS,
			Platform: runtime.GOARCH,
			CPUs:     cpuInfos,
		}
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ifaces, err := net.Interfaces()
		if err != nil {
			return
		}
		var result []NetworkInterface
		for _, iface := range ifaces {
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					result = append(result, NetworkInterface{Name: iface.Name, IP: ipnet.IP.String()})
				}
			}
		}
		mu.Lock()
		data.NetworkInterfaces = result
		mu.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		conns, err := netstat.Connections("tcp")
		if err != nil {
			return
		}
		var ports []ActivePort
		for _, c := range conns {
			if c.Status == "LISTEN" {
				ports = append(ports, ActivePort{Port: int(c.Laddr.Port), LocalAddress: c.Laddr.IP})
			}
		}
		mu.Lock()
		data.ActivePorts = ports
		mu.Unlock()
	}()

	wg.Wait()
	return data
}

func registerWithServer(serverURL, name string) (*RegisterResponse, error) {
	data := collectData()
	hostname, _ := os.Hostname()
	data.ClientID = hostname

	req := RegisterRequest{
		Name:      name,
		ServerURL: serverURL,
		Data:      *data,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(serverURL+"/api/agent/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var result RegisterResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}

type pushFileRecord struct {
	FullPath   string `json:"fullpath"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modified_at"`
	CreatedAt  int64  `json:"created_at"`
	Hash       string `json:"hash"`
	DriveType  string `json:"drive_type"`
	SyncedAt   int64  `json:"synced_at"`
}

func pushFiles(files []FileRecord) {
	if cfg.AgentID == "" || cfg.Token == "" || len(files) == 0 {
		return
	}

	const batchSize = 500
	client := &http.Client{Timeout: 30 * time.Second}

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}
		chunk := files[i:end]

		records := make([]pushFileRecord, len(chunk))
		for j, f := range chunk {
			records[j] = pushFileRecord{
				FullPath:   f.FullPath,
				Size:       f.Size,
				ModifiedAt: f.ModifiedAt.Unix(),
				CreatedAt:  f.CreatedAt.Unix(),
				Hash:       f.Hash,
				DriveType:  f.DriveType,
				SyncedAt:   f.SyncedAt.Unix(),
			}
		}

		body, err := json.Marshal(map[string]interface{}{"files": records})
		if err != nil {
			continue
		}

		req, err := http.NewRequest("POST", cfg.ServerURL+"/api/agent/files", bytes.NewBuffer(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Agent-Token", cfg.Token)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("pushFiles error: %v", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("pushFiles: server returned %d", resp.StatusCode)
		}
	}
}

// notifyServerDrives pushes the agent's current drive list to the server so
// that any drives removed locally are reflected on the server immediately
// (and their file records are deleted).
func notifyServerDrives() {
	if cfg.AgentID == "" || cfg.Token == "" {
		return
	}

	body, err := json.Marshal(map[string]interface{}{"drives": cfg.Drives})
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/agent/drives", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", cfg.Token)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("notifyServerDrives error: %v", err)
		return
	}
	resp.Body.Close()
}

func checkPendingActions() {
	if cfg.AgentID == "" || cfg.Token == "" {
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", cfg.ServerURL+"/api/agent/pending-actions", nil)
	if err != nil {
		return
	}
	req.Header.Set("X-Agent-Token", cfg.Token)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var result struct {
		Actions []struct {
			ID       int64  `json:"id"`
			FullPath string `json:"fullpath"`
			Action   string `json:"action"`
		} `json:"actions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	for _, action := range result.Actions {
		switch action.Action {
		case "delete":
			if err := os.Remove(action.FullPath); err != nil {
				log.Printf("delete failed %s: %v", action.FullPath, err)
				continue
			}
			db.Exec(`DELETE FROM files WHERE fullpath = ?`, action.FullPath)
		}
		confirmAction(action.ID, action.FullPath)
	}
}

func confirmAction(id int64, fullpath string) {
	body, _ := json.Marshal(map[string]interface{}{"id": id, "fullpath": fullpath})

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/agent/confirm-action", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", cfg.Token)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func syncData() {
	if cfg.AgentID == "" || cfg.Token == "" {
		return
	}

	data := collectData()
	hostname, _ := os.Hostname()
	data.ClientID = hostname

	stats := computeFileStats("")

	interval := cfg.ScanIntervalMinutes
	if interval <= 0 {
		interval = 10
	}

	hb := HeartbeatRequest{
		ClientData: *data,
		Drives:     cfg.Drives,
		FileStats: FileStats{
			TotalFiles:     stats.TotalFiles,
			TotalSize:      stats.TotalSize,
			DuplicateCount: stats.DuplicateCount,
			LastScanned:    stats.LastScanned,
		},
		ScanIntervalMinutes: interval,
	}

	body, err := json.Marshal(hb)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/agent/heartbeat", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", cfg.Token)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var hbResp struct {
			Drives              []DriveEntry `json:"drives"`
			ScanIntervalMinutes int          `json:"scan_interval_minutes"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&hbResp); err == nil {
			// Always update drives from server — an empty slice means the server
			// has removed all drives (e.g. user removed the last one via web).
			cfg.Drives = hbResp.Drives
			if hbResp.ScanIntervalMinutes > 0 {
				cfg.ScanIntervalMinutes = hbResp.ScanIntervalMinutes
			}
			saveConfig(cfg)
		}
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("heartbeat 401: token invalidated, attempting re-registration")
		cfg.AgentID = ""
		cfg.Token = ""
		saveConfig(cfg)

		if cfg.ServerURL != "" && cfg.AgentName != "" {
			regResp, err := registerWithServer(cfg.ServerURL, cfg.AgentName)
			if err != nil {
				log.Printf("auto re-registration failed: %v", err)
				select {
				case reRegisterCh <- struct{}{}:
				default:
				}
				return
			}
			cfg.AgentID = regResp.AgentID
			cfg.Token = regResp.Token
			saveConfig(cfg)
			log.Printf("re-registered as %s", cfg.AgentID)
		}

		select {
		case reRegisterCh <- struct{}{}:
		default:
		}
	}
}
