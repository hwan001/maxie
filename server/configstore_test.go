package main

import "testing"

func TestConfigStore(t *testing.T) {
	// Sandbox fileDB into a temp HOME so we never touch the real ~/.maxie/files.db.
	t.Setenv("HOME", t.TempDir())
	if err := initFileDB(); err != nil {
		t.Fatalf("initFileDB: %v", err)
	}

	// Absent key → found=false, no error (first-time setup signal).
	if _, found, err := getConfig(cfgKeyAdminPasswordHash); err != nil || found {
		t.Fatalf("expected absent key, got found=%v err=%v", found, err)
	}

	// Set then read back.
	if err := setConfig(cfgKeyAdminPasswordHash, "hash-v1"); err != nil {
		t.Fatalf("setConfig: %v", err)
	}
	v, found, err := getConfig(cfgKeyAdminPasswordHash)
	if err != nil || !found || v != "hash-v1" {
		t.Fatalf("get after set: v=%q found=%v err=%v", v, found, err)
	}

	// Upsert overwrites the existing value.
	if err := setConfig(cfgKeyAdminPasswordHash, "hash-v2"); err != nil {
		t.Fatalf("setConfig overwrite: %v", err)
	}
	if v, _, _ := getConfig(cfgKeyAdminPasswordHash); v != "hash-v2" {
		t.Fatalf("expected hash-v2 after overwrite, got %q", v)
	}
}
