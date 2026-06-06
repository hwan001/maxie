//go:build windows

package main

import (
	"os/exec"
	"strings"
	"time"
)

// scanBluetoothDevices returns Bluetooth devices visible to Windows via PowerShell.
func scanBluetoothDevices() []BluetoothDevice {
	// Each line: FriendlyName|Status
	script := `Get-PnpDevice -Class Bluetooth | ` +
		`ForEach-Object { $_.FriendlyName + '|' + $_.Status }`

	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return nil
	}

	now := time.Now().Unix()
	var devices []BluetoothDevice

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		name := strings.TrimSpace(parts[0])
		if name == "" {
			continue
		}
		status := ""
		if len(parts) == 2 {
			status = strings.TrimSpace(parts[1])
		}
		devices = append(devices, BluetoothDevice{
			Name:      name,
			Connected: strings.EqualFold(status, "OK"),
			ScannedAt: now,
		})
	}
	return devices
}
