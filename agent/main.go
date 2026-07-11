package main

import (
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getlantern/systray"
)

var cfg *Config
var reRegisterCh = make(chan struct{}, 1)
var reRegisterMu sync.Mutex

// scanStateCh carries scan state update strings:
//
//	"scanning"       – a full scan has started
//	"drive:PATH"     – currently scanning a specific drive
//	"idle"           – scan finished
var scanStateCh = make(chan string, 8)

// scanning guards against concurrent full-drive scans.
var scanning atomic.Bool

func main() {
	var err error
	cfg, err = loadConfig()
	if err != nil {
		cfg = &Config{}
	}

	if err := initDB(); err != nil {
		log.Printf("db init failed: %v", err)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIdleIcon())
	systray.SetTooltip("Maxie Agent")

	mTitle := systray.AddMenuItem("Maxie Agent", "")
	mTitle.Disable()

	// Unregistered items
	mRegister := systray.AddMenuItem("Register Agent…", "Register this agent with a server")
	mSepUnreg := systray.AddMenuItem("---", "")
	mSepUnreg.Disable()

	// Registered items
	mStatus := systray.AddMenuItem("", "")
	mStatus.Disable()
	systray.AddSeparator()
	mScan := systray.AddMenuItem("Scan Drives Now", "Scan all configured drives")
	mSettings := systray.AddMenuItem("Settings…", "Configure agent settings")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the agent")

	applyMenuState := func() {
		if cfg.AgentID == "" {
			mRegister.Show()
			mSepUnreg.Show()
			mStatus.Hide()
			mScan.Hide()
			mSettings.Hide()
		} else {
			mStatus.SetTitle("✓  " + cfg.AgentName)
			mStatus.SetTooltip(cfg.ServerURL)
			mRegister.Hide()
			mSepUnreg.Hide()
			mStatus.Show()
			mScan.Show()
			mSettings.Show()
		}
	}

	applyMenuState()

	// Listen for scan state changes and update icon / menu titles accordingly.
	go func() {
		idleIcon := getIdleIcon()
		scanIcon := getScanIcon()
		for state := range scanStateCh {
			switch {
			case state == "idle":
				systray.SetIcon(idleIcon)
				if cfg.AgentID != "" {
					mStatus.SetTitle("✓  " + cfg.AgentName)
				}
				mScan.SetTitle("Scan Drives Now")
				mScan.Enable()

			case state == "scanning":
				systray.SetIcon(scanIcon)
				if cfg.AgentID != "" {
					mStatus.SetTitle("⟳  Scanning…")
				}
				mScan.SetTitle("Scanning…")
				mScan.Disable()

			case strings.HasPrefix(state, "drive:"):
				drivePath := state[len("drive:"):]
				if cfg.AgentID != "" {
					// Trim long paths to keep menu readable
					label := drivePath
					if len(label) > 40 {
						label = "…" + label[len(label)-37:]
					}
					mStatus.SetTitle("⟳  " + label)
				}
			}
		}
	}()

	if cfg.AgentID == "" {
		go func() {
			time.Sleep(500 * time.Millisecond)
			doRegister(applyMenuState, mStatus)
		}()
	} else {
		go startPeriodicSync()
		go startPeriodicScan()
	}

	go func() {
		for {
			select {
			case <-mRegister.ClickedCh:
				doRegister(applyMenuState, mStatus)
			case <-mScan.ClickedCh:
				go scanAllDrives()
			case <-mSettings.ClickedCh:
				go doSettings()
			case <-reRegisterCh:
				applyMenuState()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {}

func doRegister(applyMenuState func(), mStatus *systray.MenuItem) {
	input, ok := showRegisterDialog()
	if !ok {
		return
	}

	resp, err := registerWithServer(input.ServerURL, input.AgentName, input.UserID)
	if err != nil {
		log.Printf("registration failed: %v", err)
		return
	}

	cfg.AgentID = resp.AgentID
	cfg.AgentName = input.AgentName
	cfg.Token = resp.Token
	cfg.ServerURL = input.ServerURL
	cfg.UserID = input.UserID
	if err := saveConfig(cfg); err != nil {
		log.Printf("save config failed: %v", err)
	}

	applyMenuState()
	go startPeriodicSync()
	go startPeriodicScan()
}

func doSettings() {
	action, ok := showSettingsAction()
	if !ok {
		return
	}
	switch action {
	case "Add Drive":
		doAddDrive()
	case "Remove Drive":
		doRemoveDrive()
	case "Change Server URL":
		doChangeServer()
	case "Scan Schedule":
		doScanSchedule()
	case "Link User Code":
		doLinkUserCode()
	case "Agent Info":
		showInfoDialog(cfg)
	}
}

func doScanSchedule() {
	current := cfg.ScanIntervalMinutes
	if current <= 0 {
		current = 10
	}
	mins, ok := showScheduleDialog(current)
	if !ok {
		return
	}
	cfg.ScanIntervalMinutes = mins
	if err := saveConfig(cfg); err != nil {
		log.Printf("save config failed: %v", err)
	}
}

func doAddDrive() {
	input, ok := showAddDriveDialog()
	if !ok {
		return
	}
	cfg.Drives = append(cfg.Drives, DriveEntry{
		Path:      input.Path,
		DriveType: input.DriveType,
		Label:     input.Label,
	})
	if err := saveConfig(cfg); err != nil {
		log.Printf("save config failed: %v", err)
	}
	drive := cfg.Drives[len(cfg.Drives)-1]
	// Sync the updated drive list to the server immediately so the web UI
	// reflects the new drive without waiting for the next heartbeat.
	go notifyServerDrives()
	go func() {
		records := scanDrive(drive)
		if len(records) > 0 {
			pushFiles(records)
		}
	}()
}

func doRemoveDrive() {
	idx, ok := showRemoveDriveDialog(cfg.Drives)
	if !ok {
		return
	}
	cfg.Drives = append(cfg.Drives[:idx], cfg.Drives[idx+1:]...)
	if err := saveConfig(cfg); err != nil {
		log.Printf("save config failed: %v", err)
	}
	// Tell the server immediately so it removes the drive and deletes its
	// file records — without this the heartbeat would restore the drive.
	go notifyServerDrives()
}

func doLinkUserCode() {
	code, ok := showLinkUserCodeDialog(cfg.UserID)
	if !ok {
		return
	}
	if err := linkUserToServer(code); err != nil {
		log.Printf("link user code failed: %v", err)
		return
	}
	cfg.UserID = code
	if err := saveConfig(cfg); err != nil {
		log.Printf("save config failed: %v", err)
	}
}

func doChangeServer() {
	url, ok := showChangeServerDialog(cfg.ServerURL)
	if !ok {
		return
	}
	cfg.ServerURL = url
	if err := saveConfig(cfg); err != nil {
		log.Printf("save config failed: %v", err)
	}
}

// scanAllDrives scans every configured drive and pushes file records to the
// server. Concurrent invocations are silently dropped (one scan at a time).
func scanAllDrives() {
	if !scanning.CompareAndSwap(false, true) {
		return // already in progress
	}
	defer scanning.Store(false)

	scanStateCh <- "scanning"
	defer func() { scanStateCh <- "idle" }()

	var all []FileRecord
	for _, d := range cfg.Drives {
		scanStateCh <- "drive:" + d.Path
		records := scanDrive(d)
		all = append(all, records...)
	}
	syncData()
	if len(all) > 0 {
		go pushFiles(all)
	}
}

func startPeriodicSync() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		syncData()
		go checkPendingActions()
	}
}

func startPeriodicScan() {
	for {
		interval := cfg.ScanIntervalMinutes
		if interval <= 0 {
			interval = 10
		}
		time.Sleep(time.Duration(interval) * time.Minute)
		scanAllDrives()
	}
}
