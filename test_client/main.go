package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "sync"
    "time"

    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/host"
    netstat "github.com/shirou/gopsutil/net"
)

// 데이터 구조 정의
type ClientData struct {
    ClientID          string              `json:"client_id"`
    NetworkInterfaces []NetworkInterface  `json:"network_interfaces"`
    SystemInfo        SystemInfo          `json:"system_info"`
    ActivePorts       []ActivePort        `json:"active_ports"`
}

type NetworkInterface struct {
    Name string `json:"name"`
    IP   string `json:"ip"`
}

type SystemInfo struct {
    OS              string   `json:"os"`
    Platform        string   `json:"platform"`
    PlatformVersion string   `json:"platform_version"`
    CPUs            []CPUInfo `json:"cpus"`
}

type CPUInfo struct {
    ModelName string  `json:"model_name"`
    Cores     int     `json:"cores"`
    SpeedGHz  float64 `json:"speed_ghz"`
}

type ActivePort struct {
    Port         int    `json:"port"`
    LocalAddress string `json:"local_address"`
}


// 네트워크 정보 수집 함수
func collectNetworkInfo(wg *sync.WaitGroup, data *ClientData, mutex *sync.Mutex) {
    defer wg.Done()

    fmt.Println("Collecting network interfaces...")

    interfaces, err := net.Interfaces()
    if err != nil {
        fmt.Printf("Error fetching network interfaces: %v\n", err)
        return
    }

    var networkInterfaces []NetworkInterface

    for _, iface := range interfaces {
        addrs, err := iface.Addrs()
        if err != nil {
            fmt.Printf("Error fetching addresses: %v\n", err)
            continue
        }

        for _, addr := range addrs {
            if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
                networkInterfaces = append(networkInterfaces, NetworkInterface{
                    Name: iface.Name,
                    IP:   ipnet.IP.String(),
                })

                // 기본 포트 스캔
                fmt.Println("Scanning basic ports (22, 80)...")
                scanIPRange(ipnet.IP.String(), []int{22, 80})
            }
        }
    }

    mutex.Lock()
    data.NetworkInterfaces = networkInterfaces
    mutex.Unlock()
}

// 포트 스캔 함수
func scanIPRange(ip string, ports []int) {
    for _, port := range ports {
        address := fmt.Sprintf("%s:%d", ip, port)
        conn, err := net.DialTimeout("tcp", address, time.Second)
        if err != nil {
            fmt.Printf("IP: %s port %d is not active\n", ip, port)
        } else {
            fmt.Printf("IP: %s port %d is active\n", ip, port)
            conn.Close()
        }
    }
}


// 시스템 정보 수집 함수
func collectSystemInfo(wg *sync.WaitGroup, data *ClientData, mutex *sync.Mutex) {
    defer wg.Done()

    fmt.Println("Collecting system information...")

    hostInfo, err := host.Info()
    if err != nil {
        fmt.Printf("Error fetching host info: %v\n", err)
        return
    }

    cpuInfo, err := cpu.Info()
    if err != nil {
        fmt.Printf("Error fetching CPU info: %v\n", err)
        return
    }

    var cpus []CPUInfo
    for _, ci := range cpuInfo {
        cpus = append(cpus, CPUInfo{
            ModelName: ci.ModelName,
            Cores:     int(ci.Cores),
            SpeedGHz:  ci.Mhz / 1000.0,
        })
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


// 현재 사용 중인 포트 정보 수집 함수
func collectActivePorts(wg *sync.WaitGroup, data *ClientData, mutex *sync.Mutex) {
    defer wg.Done()

    fmt.Println("Collecting active ports...")
    netStats, err := netstat.Connections("tcp")
    if err != nil {
        fmt.Printf("Error fetching network stats: %v\n", err)
        return
    }

    var activePorts []ActivePort
    for _, stat := range netStats {
        if stat.Status == "LISTEN" {
            activePorts = append(activePorts, ActivePort{
				Port:         int(stat.Laddr.Port),
				LocalAddress: stat.Laddr.IP,
			})
        }
    }

    mutex.Lock()
    data.ActivePorts = activePorts
    mutex.Unlock()
}

// 서버와 연결하여 고유 코드 발급 함수
func connectToServer(wg *sync.WaitGroup, data *ClientData, mutex *sync.Mutex) {
    defer wg.Done()

    fmt.Println("Connecting to server and requesting client ID...")
    // 서버에 고유 코드 요청 로직 추가 (예: HTTP 요청)

    // 예시: 서버에서 클라이언트 ID를 생성하여 반환한다고 가정
    resp, err := http.Get("http://goserver.666lab.org/get-client-id")
    if err != nil {
        fmt.Printf("Error connecting to server: %v\n", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        var result map[string]string
        json.NewDecoder(resp.Body).Decode(&result)
        clientID, exists := result["client_id"]
        if exists {
            mutex.Lock()
            data.ClientID = clientID
            mutex.Unlock()
            fmt.Printf("Assigned client ID: %s\n", clientID)
        }
    }
}
func main() {
    var wg sync.WaitGroup
    var mutex sync.Mutex

    clientData := &ClientData{}

    // 네트워크 정보 수집 작업
    wg.Add(1)
    go collectNetworkInfo(&wg, clientData, &mutex)

    // 시스템 정보 수집 작업
    wg.Add(1)
    go collectSystemInfo(&wg, clientData, &mutex)

    // 활성화된 포트 정보 수집 작업
    wg.Add(1)
    go collectActivePorts(&wg, clientData, &mutex)

    // 서버와 연결하여 고유 코드 발급 작업
    wg.Add(1)
    go connectToServer(&wg, clientData, &mutex)

    // 모든 작업 완료 대기
    wg.Wait()

    fmt.Println("All tasks completed.")

    // 데이터 서버로 전송
    sendDataToServer(clientData)
}

// 데이터 서버로 전송하는 함수
func sendDataToServer(data *ClientData) {
    url := "https://goserver.666lab.org/client-data"

    jsonData, err := json.Marshal(data)
    if err != nil {
        log.Fatalf("Error marshalling data: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        log.Fatalf("Error creating request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{
        Timeout: time.Second * 10,
    }

    resp, err := client.Do(req)
    if err != nil {
        log.Fatalf("Error sending data to server: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        fmt.Println("Data sent successfully")
    } else {
        fmt.Printf("Failed to send data. Status code: %d\n", resp.StatusCode)
    }
}
