package collector

import (
	"os"
	"runtime"
)

type SystemInfo struct {
	OS           string
	Hostname     string
	CPUs         int
	Architecture string
}

func CollectSystemInfo() *SystemInfo {
	hostname, _ := os.Hostname()

	return &SystemInfo{
		OS:           runtime.GOOS,
		Hostname:     hostname,
		CPUs:         runtime.NumCPU(),
		Architecture: runtime.GOARCH,
	}
}
