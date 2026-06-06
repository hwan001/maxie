//go:build linux

package main

import (
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// scanWiFiNetworks returns nearby Wi-Fi networks using nmcli.
func scanWiFiNetworks() []WiFiNetwork {
	// Use multiline mode so BSSID colons don't conflict with the field separator.
	out, err := exec.Command("nmcli",
		"--mode", "multiline",
		"-f", "SSID,BSSID,SIGNAL,SECURITY,CHAN",
		"dev", "wifi", "list").Output()
	if err != nil {
		return nil
	}

	var networks []WiFiNetwork
	now := time.Now().Unix()
	current := WiFiNetwork{ScannedAt: now}
	started := false

	for _, line := range strings.Split(string(out), "\n") {
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch key {
		case "SSID":
			if started {
				networks = append(networks, current)
			}
			current = WiFiNetwork{SSID: val, ScannedAt: now}
			started = true
		case "BSSID":
			current.BSSID = val
		case "SIGNAL":
			current.Signal = val
		case "SECURITY":
			current.Security = val
		case "CHAN":
			current.Channel = val
		}
	}
	if started {
		networks = append(networks, current)
	}
	return networks
}

// getWiFiHistory returns SSIDs of saved Wi-Fi connections via nmcli.
func getWiFiHistory() []string {
	out, err := exec.Command("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show").Output()
	if err != nil {
		return nil
	}

	wifiRe := regexp.MustCompile(`802-11-wireless|wifi`)
	var profiles []string
	for _, line := range strings.Split(string(out), "\n") {
		if !wifiRe.MatchString(line) {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) >= 1 && parts[0] != "" {
			profiles = append(profiles, parts[0])
		}
	}
	return profiles
}
