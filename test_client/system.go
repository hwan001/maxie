package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
)

func getSysctlValue(key string) (string, error) {
	cmd := exec.Command("sysctl", "-n", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func collectSystemInfo(wg *sync.WaitGroup, data *ClientData, mutex *sync.Mutex) {
	defer wg.Done()

	fmt.Println("Collecting system information...")

	hostInfo, err := host.Info()
	if err != nil {
		fmt.Printf("Error fetching host info: %v\n", err)
		return
	}

	var cpus []CPUInfo

	switch runtime.GOOS {
	case "darwin":
		modelName, _ := getSysctlValue("machdep.cpu.brand_string")

		coreCountStr, _ := getSysctlValue("hw.physicalcpu")
		coreCount, _ := strconv.Atoi(coreCountStr)

		speedMHzStr, _ := getSysctlValue("hw.cpufrequency")
		speedMHz, _ := strconv.Atoi(speedMHzStr)

		cpus = append(cpus, CPUInfo{
			ModelName: modelName,
			Cores:     coreCount,
			SpeedGHz:  float64(speedMHz) / 1000000000.0, // Hz -> GHz 변환
		})
	case "linux", "windows":
		cpuInfo, err := cpu.Info()
		if err != nil {
			fmt.Printf("Error fetching CPU info: %v\n", err)
			return
		}

		for _, ci := range cpuInfo {
			cpus = append(cpus, CPUInfo{
				ModelName: ci.ModelName,
				Cores:     int(ci.Cores),
				SpeedGHz:  ci.Mhz / 1000.0,
			})
		}
	default:
		fmt.Println("Unsupported OS")
		return
	}

	systemInfo := SystemInfo{
		OS:              hostInfo.OS,
		Platform:        hostInfo.Platform,
		PlatformVersion: hostInfo.PlatformVersion,
		CPUs:            cpus,
	}

	mutex.Lock()
	data.SystemInfo = systemInfo
	mutex.Unlock()
}
