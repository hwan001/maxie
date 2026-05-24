package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ncruces/zenity"
)

type RegisterInput struct {
	ServerURL string
	AgentName string
	UserID    string // user UUID from the web dashboard (optional for guest users)
}

func testServerConnection(url string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url + "/api/agents")
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

func showRegisterDialog() (*RegisterInput, bool) {
	var serverURL string
	for {
		var err error
		serverURL, err = zenity.Entry(
			"Enter server URL (e.g. https://server.domain.com:",
			zenity.Title("File Optimizer — Register Agent"),
			zenity.EntryText("https://"),
		)
		if err != nil || strings.TrimSpace(serverURL) == "" {
			return nil, false
		}
		serverURL = strings.TrimRight(strings.TrimSpace(serverURL), "/")

		if connErr := testServerConnection(serverURL); connErr != nil {
			msg := fmt.Sprintf("✗  Cannot connect to server:\n%v\n\nTry a different URL?", connErr)
			if zenity.Question(msg, zenity.Title("Connection Failed")) != nil {
				return nil, false
			}
			continue
		}
		zenity.Info("✓  Connected successfully", zenity.Title("Connection OK"))
		break
	}

	agentName, err := zenity.Entry(
		"Enter a name for this agent:",
		zenity.Title("File Optimizer — Register Agent"),
		zenity.EntryText(defaultAgentName()),
	)
	if err != nil || strings.TrimSpace(agentName) == "" {
		return nil, false
	}

	// User code is optional — guests or OAuth users paste their UUID here so
	// the server can scope this agent to their account. Skip by pressing Cancel.
	userID := ""
	if code, err := zenity.Entry(
		"Paste your User Code from the web dashboard\n(leave empty or cancel to skip):",
		zenity.Title("File Optimizer — User Code (optional)"),
	); err == nil {
		userID = strings.TrimSpace(code)
	}

	return &RegisterInput{
		ServerURL: serverURL,
		AgentName: strings.TrimSpace(agentName),
		UserID:    userID,
	}, true
}

func showSettingsAction() (string, bool) {
	item, err := zenity.List(
		"Select an action:",
		[]string{"Add Drive", "Remove Drive", "Change Server URL", "Scan Schedule", "Link User Code", "Agent Info"},
		zenity.Title("File Optimizer — Settings"),
	)
	if err != nil {
		return "", false
	}
	return item, true
}

func showLinkUserCodeDialog(current string) (string, bool) {
	prompt := "Paste your User Code from the web dashboard:"
	if current != "" {
		prompt = fmt.Sprintf("Current User Code: %s\n\nPaste a new User Code to replace it:", current)
	}
	code, err := zenity.Entry(
		prompt,
		zenity.Title("File Optimizer — Link User Code"),
	)
	if err != nil || strings.TrimSpace(code) == "" {
		return "", false
	}
	return strings.TrimSpace(code), true
}

func showScheduleDialog(currentMinutes int) (int, bool) {
	type option struct {
		label   string
		minutes int
	}
	opts := []option{
		{"5 minutes", 5},
		{"10 minutes", 10},
		{"30 minutes", 30},
		{"1 hour", 60},
		{"3 hours", 180},
		{"6 hours", 360},
		{"24 hours", 1440},
	}

	labels := make([]string, len(opts))
	for i, o := range opts {
		labels[i] = o.label
	}

	currentLabel := "10 minutes"
	for _, o := range opts {
		if o.minutes == currentMinutes {
			currentLabel = o.label
			break
		}
	}

	chosen, err := zenity.List(
		fmt.Sprintf("Select scan interval (current: %s):", currentLabel),
		labels,
		zenity.Title("File Optimizer — Scan Schedule"),
	)
	if err != nil {
		return 0, false
	}

	for _, o := range opts {
		if o.label == chosen {
			return o.minutes, true
		}
	}
	return 0, false
}

func showChangeServerDialog(current string) (string, bool) {
	url, err := zenity.Entry(
		"Enter new server URL:",
		zenity.Title("File Optimizer — Change Server"),
		zenity.EntryText(current),
	)
	if err != nil || strings.TrimSpace(url) == "" {
		return "", false
	}
	return strings.TrimRight(strings.TrimSpace(url), "/"), true
}

type DriveInput struct {
	Path      string
	DriveType string
	Label     string
}

func showAddDriveDialog() (*DriveInput, bool) {
	path, err := zenity.SelectFile(
		zenity.Title("Select Drive / Folder to Monitor"),
		zenity.Directory(),
	)
	if err != nil || path == "" {
		return nil, false
	}

	driveTypeStr, err := zenity.List(
		"Select drive type:",
		[]string{"local", "google", "naver", "onedrive", "dropbox", "icloud", "other"},
		zenity.Title("Drive Type"),
	)
	if err != nil {
		return nil, false
	}

	label, err := zenity.Entry(
		"Label for this drive (optional):",
		zenity.Title("Drive Label"),
		zenity.EntryText(defaultLabel(path, driveTypeStr)),
	)
	if err != nil {
		label = defaultLabel(path, driveTypeStr)
	}

	return &DriveInput{
		Path:      path,
		DriveType: driveTypeStr,
		Label:     strings.TrimSpace(label),
	}, true
}

func showInfoDialog(c *Config) {
	msg := fmt.Sprintf(
		"Agent: %s\nID: %s\nServer: %s\n\nDrives (%d):",
		c.AgentName, c.AgentID, c.ServerURL, len(c.Drives),
	)
	for _, d := range c.Drives {
		msg += fmt.Sprintf("\n  [%s] %s  (%s)", d.DriveType, d.Label, d.Path)
	}
	zenity.Info(msg, zenity.Title("File Optimizer Agent"))
}

func showRemoveDriveDialog(drives []DriveEntry) (int, bool) {
	if len(drives) == 0 {
		zenity.Info("No drives configured.", zenity.Title("Remove Drive"))
		return 0, false
	}
	labels := make([]string, len(drives))
	for i, d := range drives {
		labels[i] = fmt.Sprintf("[%s] %s (%s)", d.DriveType, d.Label, d.Path)
	}
	chosen, err := zenity.List("Select drive to remove:", labels, zenity.Title("Remove Drive"))
	if err != nil {
		return 0, false
	}
	for i, l := range labels {
		if l == chosen {
			return i, true
		}
	}
	return 0, false
}

func defaultAgentName() string {
	h, _ := os.Hostname()
	if h == "" {
		return "my-agent"
	}
	return h
}

func defaultLabel(path, driveType string) string {
	parts := strings.Split(strings.TrimRight(path, "/\\"), "/")
	if len(parts) == 0 {
		return driveType
	}
	last := parts[len(parts)-1]
	if last == "" && len(parts) > 1 {
		last = parts[len(parts)-2]
	}
	return last
}
