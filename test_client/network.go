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

	netstat "github.com/shirou/gopsutil/net"
)

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

func connectToServer(wg *sync.WaitGroup, data *ClientData, mutex *sync.Mutex) {
	defer wg.Done()

	fmt.Println("Connecting to server and requesting client ID...")
	resp, err := http.Get(server_url + "/get-client-id")
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

func sendDataToServer(data *ClientData) {
	url := server_url + "/client-data"

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
