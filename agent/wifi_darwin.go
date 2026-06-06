//go:build darwin

package main

import (
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const airportBin = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"

// scanWiFiNetworks returns nearby Wi-Fi networks using the airport utility.
func scanWiFiNetworks() []WiFiNetwork {
	out, err := exec.Command(airportBin, "-s").Output()
	if err != nil {
		return nil
	}

	macRe := regexp.MustCompile(`[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}:[0-9a-f]{2}`)
	var networks []WiFiNetwork
	now := time.Now().Unix()

	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i == 0 { // header row
			continue
		}
		loc := macRe.FindStringIndex(line)
		if loc == nil {
			continue
		}
		ssid := strings.TrimSpace(line[:loc[0]])
		rest := strings.Fields(line[loc[0]:])

		net := WiFiNetwork{SSID: ssid, ScannedAt: now}
		if len(rest) >= 1 {
			net.BSSID = rest[0]
		}
		if len(rest) >= 2 {
			net.Signal = rest[1] // RSSI
		}
		if len(rest) >= 3 {
			net.Channel = rest[2]
		}
		if len(rest) >= 6 {
			net.Security = rest[5]
		}
		networks = append(networks, net)
	}
	return networks
}

// getWiFiHistory returns SSIDs of previously connected Wi-Fi networks.
func getWiFiHistory() []string {
	// Requires read access to the airport preferences file.
	out, err := exec.Command("defaults", "read",
		"/Library/Preferences/SystemConfiguration/com.apple.airport.preferences",
		"KnownNetworks").Output()
	if err != nil {
		return nil
	}

	var ssids []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		// Plist key is either "SSID_STR" or "SSIDString" depending on macOS version.
		if strings.Contains(line, "SSID_STR") || strings.Contains(line, "SSIDString") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				ssid := strings.Trim(strings.TrimSpace(parts[1]), `";`)
				if ssid != "" {
					ssids = append(ssids, ssid)
				}
			}
		}
	}
	return ssids
}
