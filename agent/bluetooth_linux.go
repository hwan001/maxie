//go:build linux

package main

import (
	"os/exec"
	"strings"
	"time"
)

// scanBluetoothDevices returns paired/known Bluetooth devices via bluetoothctl.
func scanBluetoothDevices() []BluetoothDevice {
	out, err := exec.Command("bluetoothctl", "devices").Output()
	if err != nil {
		return nil
	}

	now := time.Now().Unix()

	// Check connected devices separately
	connectedOut, _ := exec.Command("bluetoothctl", "devices", "Connected").Output()
	connected := map[string]bool{}
	for _, line := range strings.Split(string(connectedOut), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			connected[parts[1]] = true // address
		}
	}

	var devices []BluetoothDevice
	for _, line := range strings.Split(string(out), "\n") {
		// Format: Device AA:BB:CC:DD:EE:FF DeviceName
		parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
		if len(parts) < 3 || parts[0] != "Device" {
			continue
		}
		addr := parts[1]
		name := parts[2]
		devices = append(devices, BluetoothDevice{
			Name:      name,
			Address:   strings.ToLower(addr),
			Connected: connected[addr],
			ScannedAt: now,
		})
	}
	return devices
}
