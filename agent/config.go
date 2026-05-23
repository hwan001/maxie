package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type DriveEntry struct {
	Path        string   `json:"path"`
	DriveType   string   `json:"drive_type"` // google, naver, local, etc.
	Label       string   `json:"label"`
	ExcludeDirs []string `json:"exclude_dirs,omitempty"`
	ExcludeExts []string `json:"exclude_exts,omitempty"`
}

type Config struct {
	AgentID             string       `json:"agent_id"`
	AgentName           string       `json:"agent_name"`
	Token               string       `json:"token"`
	ServerURL           string       `json:"server_url"`
	Drives              []DriveEntry `json:"drives"`
	ScanIntervalMinutes int          `json:"scan_interval_minutes"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".fileoptimizer")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

func dbPath() string {
	return filepath.Join(configDir(), "cache.db")
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return &Config{}, nil
	}
	var c Config
	return &c, json.Unmarshal(data, &c)
}

func saveConfig(c *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
