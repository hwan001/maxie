//go:build windows

package main

import (
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// scanWiFiNetworks returns nearby Wi-Fi access points using netsh.
func scanWiFiNetworks() []WiFiNetwork {
	cmd := exec.Command("netsh", "wlan", "show", "networks", "mode=bssid")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var networks []WiFiNetwork
	var current *WiFiNetwork
	now := time.Now().Unix()

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)

		// New SSID block (exclude lines with "BSSID")
		if strings.HasPrefix(line, "SSID") && !strings.Contains(line, "BSSID") {
			if current != nil {
				networks = append(networks, *current)
			}
			if kv := wifiKV(line); kv[1] != "" {
				current = &WiFiNetwork{SSID: kv[1], ScannedAt: now}
			}
			continue
		}
		if current == nil {
			continue
		}
		switch {
		case strings.HasPrefix(line, "BSSID 1"):
			if kv := wifiKV(line); kv[1] != "" {
				current.BSSID = kv[1]
			}
		case strings.HasPrefix(line, "Signal"):
			if kv := wifiKV(line); kv[1] != "" {
				current.Signal = kv[1]
			}
		case strings.HasPrefix(line, "Authentication"):
			if kv := wifiKV(line); kv[1] != "" {
				current.Security = kv[1]
			}
		case strings.HasPrefix(line, "Channel"):
			if kv := wifiKV(line); kv[1] != "" {
				current.Channel = kv[1]
			}
		}
	}
	if current != nil {
		networks = append(networks, *current)
	}
	return networks
}

// getWiFiHistory returns the list of Wi-Fi profiles saved on the machine.
func getWiFiHistory() []string {
	cmd := exec.Command("netsh", "wlan", "show", "profiles")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var profiles []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "All User Profile") {
			if kv := wifiKV(line); kv[1] != "" {
				profiles = append(profiles, kv[1])
			}
		}
	}
	return profiles
}

// wifiKV splits "Key : Value" into [Key, Value].
func wifiKV(line string) [2]string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return [2]string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
	}
	return [2]string{strings.TrimSpace(line), ""}
}
