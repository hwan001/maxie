package collector

import (
	"fmt"
	"net"
	netstat "github.com/shirou/gopsutil/net"
	"time"
	"regexp"
)

func CollectNetworInfo() {
	// 네트워크 통계 정보 가져오기
	netStats, err := netstat.IOCounters(true)
	if err != nil {
		fmt.Printf("Error fetching network stats: %v\n", err)
		return
	}

	for _, stat := range netStats {
		fmt.Printf("Interface: %s\n", stat.Name)
		fmt.Printf("Bytes Sent: %d\n", stat.BytesSent)
		fmt.Printf("Bytes Received: %d\n", stat.BytesRecv)
		fmt.Printf("Packets Sent: %d\n", stat.PacketsSent)
		fmt.Printf("Packets Received: %d\n", stat.PacketsRecv)
		fmt.Printf("Errors In: %d\n", stat.Errin)
		fmt.Printf("Errors Out: %d\n", stat.Errout)
		fmt.Println()
	}
}

func DefaultNetworkInfo(){
	interfaces, err := net.Interfaces() // 시스템의 모든 네트워크 인터페이스 가져오기

	cidrRegex := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2}$`)

	if err != nil {
		fmt.Printf("Error fetching network interfaces: %v\n", err)
		return
	}

	for _, iface := range interfaces {
		fmt.Printf("Name: %s\n", iface.Name)
		fmt.Printf("MAC: %s\n", iface.HardwareAddr.String())

		addrs, err := iface.Addrs() // 해당 인터페이스의 IP 주소 가져오기
		if err != nil {
			fmt.Printf("Error fetching addresses: %v\n", err)
			continue
		}

		for _, addr := range addrs {
			fmt.Printf(" Address: %s\n", addr.String())
			
			// IP 주소에 대해 정규표현식 검사
			if iface.Name != "lo0" && cidrRegex.MatchString(addr.String()) {
				// IP 주소가 정규표현식 패턴과 일치하는 경우
				fmt.Printf(" -> Matching IPv4 found: %s\n", addr.String())

				// 스캔 함수 호출 (scanIPRange 함수는 따로 정의해야 함)
				scanIPRange(addr.String(), 443) // 주석 처리된 부분은 스캔 함수 예시
			}
		}

		fmt.Println()
	}
}

func scanIPRange(networkCIDR string, port int) {
	// 네트워크 CIDR 범위로부터 IP 대역을 생성
	ip, ipnet, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		fmt.Printf("Invalid CIDR: %v\n", err)
		return
	}

	// CIDR 범위 내의 각 IP 주소를 순회하면서 포트 스캔
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		// 각 IP에 대해 포트 스캔 시도
		address := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
			// 연결 실패: 서버가 활성화되지 않음
			fmt.Printf("IP: %s is not active on port %d\n", ip, port)
		} else {
			// 연결 성공: 서버가 활성화됨
			fmt.Printf("IP: %s is active on port %d\n", ip, port)
			conn.Close()
		}
	}
}

// IP 주소를 1씩 증가시키는 함수
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}