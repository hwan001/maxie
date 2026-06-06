//go:build darwin

package main

import (
	"os/exec"
	"strings"
	"time"
)

// scanBluetoothDevices returns Bluetooth devices via system_profiler.
func scanBluetoothDevices() []BluetoothDevice {
	out, err := exec.Command("system_profiler", "SPBluetoothDataType").Output()
	if err != nil {
		return nil
	}

	now := time.Now().Unix()
	var devices []BluetoothDevice

	// system_profiler text output has device names as indented section headers
	// followed by indented key-value pairs. Example:
	//
	//   Devices (Paired, Disconnected):
	//
	//     MyKeyboard:
	//       Address: AA-BB-CC-DD-EE-FF
	//       ...
	//   Devices (Paired, Connected):
	//     MyMouse:
	//       Address: 11-22-33-44-55-66

	var inDevices bool
	var currentName string
	var connected bool

	for _, line := range strings.Split(string(out), "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "Devices (") {
			inDevices = true
			connected = strings.Contains(trimmed, "Connected")
			continue
		}
		if !inDevices {
			continue
		}

		// A device name is at indentation level 4 (two extra spaces beyond "Devices:")
		// and ends with ":"
		if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "=") &&
			!strings.HasPrefix(trimmed, "Address") && !strings.HasPrefix(trimmed, "Minor") {
			// Flush previous device
			if currentName != "" {
				devices = append(devices, BluetoothDevice{
					Name:      currentName,
					Connected: connected,
					ScannedAt: now,
				})
			}
			currentName = strings.TrimSuffix(trimmed, ":")
			continue
		}

		if currentName != "" && strings.HasPrefix(trimmed, "Address:") {
			addr := strings.TrimSpace(strings.TrimPrefix(trimmed, "Address:"))
			addr = strings.ReplaceAll(addr, "-", ":")
			devices = append(devices, BluetoothDevice{
				Name:      currentName,
				Address:   strings.ToLower(addr),
				Connected: connected,
				ScannedAt: now,
			})
			currentName = ""
		}
	}
	// Flush last device if it had no Address line
	if currentName != "" {
		devices = append(devices, BluetoothDevice{
			Name:      currentName,
			Connected: connected,
			ScannedAt: now,
		})
	}
	return devices
}
