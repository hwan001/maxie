package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// networkCache holds cached results for slow external lookups so they are not
// repeated on every heartbeat tick. Cache entries expire after cacheTTL.
const cacheTTL = 15 * time.Minute

var (
	cacheMu      sync.Mutex
	cachedIP     string
	cachedIPAt   time.Time
	cachedGeo    *GeoLocation
	cachedGeoAt  time.Time
)

// NetworkDevice represents a device discovered on the local network via ARP.
type NetworkDevice struct {
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname,omitempty"`
}

// WiFiNetwork represents a nearby Wi-Fi access point.
type WiFiNetwork struct {
	SSID      string `json:"ssid"`
	BSSID     string `json:"bssid,omitempty"`
	Signal    string `json:"signal,omitempty"`
	Security  string `json:"security,omitempty"`
	Channel   string `json:"channel,omitempty"`
	ScannedAt int64  `json:"scanned_at"`
}

// BluetoothDevice represents a detected Bluetooth device.
type BluetoothDevice struct {
	Name      string `json:"name"`
	Address   string `json:"address,omitempty"`
	Connected bool   `json:"connected"`
	ScannedAt int64  `json:"scanned_at"`
}

// GeoLocation is an approximate location derived from the public IP address.
type GeoLocation struct {
	Country  string  `json:"country,omitempty"`
	Region   string  `json:"region,omitempty"`
	City     string  `json:"city,omitempty"`
	Lat      float64 `json:"lat,omitempty"`
	Lon      float64 `json:"lon,omitempty"`
	Timezone string  `json:"timezone,omitempty"`
	Source   string  `json:"source"`
}

// fetchPublicIP performs the actual HTTP request for the public IP.
func fetchPublicIP() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}

// getPublicIP returns the machine's public (WAN) IP address.
// Results are cached for cacheTTL to avoid hammering the external service.
func getPublicIP() string {
	cacheMu.Lock()
	if cachedIP != "" && time.Since(cachedIPAt) < cacheTTL {
		ip := cachedIP
		cacheMu.Unlock()
		return ip
	}
	cacheMu.Unlock()

	ip := fetchPublicIP()
	if ip != "" {
		cacheMu.Lock()
		cachedIP = ip
		cachedIPAt = time.Now()
		cacheMu.Unlock()
	}
	return ip
}

// getARPNeighbors returns devices currently in the local ARP cache.
func getARPNeighbors() []NetworkDevice {
	cmd := exec.Command("arp", "-a")
	hideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	ipRe := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	macRe := regexp.MustCompile(`[0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}[:\-][0-9a-fA-F]{2}`)

	var devices []NetworkDevice
	seen := map[string]bool{}

	for _, line := range strings.Split(string(out), "\n") {
		ip := ipRe.FindString(line)
		mac := macRe.FindString(line)
		if ip == "" || mac == "" {
			continue
		}
		mac = strings.ToLower(mac)
		if seen[ip] {
			continue
		}
		seen[ip] = true

		hostname := ""
		if idx := strings.Index(line, "("+ip+")"); idx > 0 {
			hostname = strings.TrimSpace(line[:idx])
			if hostname == "?" {
				hostname = ""
			}
		}
		devices = append(devices, NetworkDevice{IP: ip, MAC: mac, Hostname: hostname})
	}
	return devices
}

// fetchGeoLocationIpapi tries ipapi.co (free tier, HTTPS, 1000 req/day).
func fetchGeoLocationIpapi(client *http.Client) *GeoLocation {
	resp, err := client.Get("https://ipapi.co/json/")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Error       bool    `json:"error"`
		CountryName string  `json:"country_name"`
		Region      string  `json:"region"`
		City        string  `json:"city"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Timezone    string  `json:"timezone"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Error {
		return nil
	}
	if result.Latitude == 0 && result.Longitude == 0 {
		return nil
	}
	return &GeoLocation{
		Country:  result.CountryName,
		Region:   result.Region,
		City:     result.City,
		Lat:      result.Latitude,
		Lon:      result.Longitude,
		Timezone: result.Timezone,
		Source:   "ipapi.co",
	}
}

// fetchGeoLocationIpinfo tries ipinfo.io as a fallback (50k req/month free).
func fetchGeoLocationIpinfo(client *http.Client) *GeoLocation {
	resp, err := client.Get("https://ipinfo.io/json")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Country  string `json:"country"`
		Region   string `json:"region"`
		City     string `json:"city"`
		Loc      string `json:"loc"` // "lat,lon"
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Loc == "" {
		return nil
	}

	var lat, lon float64
	if _, err := fmt.Sscanf(result.Loc, "%f,%f", &lat, &lon); err != nil || (lat == 0 && lon == 0) {
		return nil
	}
	return &GeoLocation{
		Country:  result.Country,
		Region:   result.Region,
		City:     result.City,
		Lat:      lat,
		Lon:      lon,
		Timezone: result.Timezone,
		Source:   "ipinfo.io",
	}
}

// fetchGeoLocation tries ipapi.co first, falls back to ipinfo.io.
func fetchGeoLocation() *GeoLocation {
	client := &http.Client{Timeout: 5 * time.Second}
	if geo := fetchGeoLocationIpapi(client); geo != nil {
		return geo
	}
	return fetchGeoLocationIpinfo(client)
}

// getGeoLocation returns approximate location inferred from the public IP.
// Results are cached for cacheTTL to avoid hammering the external service.
func getGeoLocation() *GeoLocation {
	cacheMu.Lock()
	if cachedGeo != nil && time.Since(cachedGeoAt) < cacheTTL {
		geo := cachedGeo
		cacheMu.Unlock()
		return geo
	}
	cacheMu.Unlock()

	geo := fetchGeoLocation()
	if geo != nil {
		cacheMu.Lock()
		cachedGeo = geo
		cachedGeoAt = time.Now()
		cacheMu.Unlock()
	}
	return geo
}
