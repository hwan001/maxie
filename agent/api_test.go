package main

import (
	"path/filepath"
	"testing"
)

// TestIsPathWithinConfiguredDrives verifies the remote-delete guard: only files
// that are real descendants of a configured drive may be deleted. Drive roots,
// parent-traversal escapes, and sibling paths that merely share a name prefix
// must all be rejected.
func TestIsPathWithinConfiguredDrives(t *testing.T) {
	root := t.TempDir()
	driveA := filepath.Join(root, "driveA")
	driveB := filepath.Join(root, "driveB")
	outside := filepath.Join(root, "outside")

	cfg = &Config{Drives: []DriveEntry{{Path: driveA}, {Path: driveB}}}
	t.Cleanup(func() { cfg = &Config{} })

	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{"file directly under drive", filepath.Join(driveA, "file.txt"), true},
		{"nested file under drive", filepath.Join(driveA, "sub", "deep", "f.bin"), true},
		{"file under second drive", filepath.Join(driveB, "x.log"), true},
		{"file outside all drives", filepath.Join(outside, "evil.txt"), false},
		{"drive root itself", driveA, false},
		{"parent traversal escape", filepath.Join(driveA, "..", "outside", "evil.txt"), false},
		{"sibling name-prefix trick", driveA + "-evil" + string(filepath.Separator) + "file.txt", false},
		{"absolute unrelated path", "/etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPathWithinConfiguredDrives(tt.target); got != tt.want {
				t.Errorf("isPathWithinConfiguredDrives(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

// TestIsPathWithinConfiguredDrivesNoDrives ensures nothing is deletable when no
// drives are configured.
func TestIsPathWithinConfiguredDrivesNoDrives(t *testing.T) {
	cfg = &Config{}
	if isPathWithinConfiguredDrives("/anything/at/all.txt") {
		t.Error("expected false when no drives are configured")
	}
}
