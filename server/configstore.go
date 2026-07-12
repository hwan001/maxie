package main

import (
	"database/sql"
	"fmt"
	"time"
)

// Well-known app_config keys. Values are opaque strings; each caller owns its
// own encoding (e.g. the admin password is stored as a bcrypt hash, never plain).
const (
	cfgKeyAdminPasswordHash = "admin_password_hash"
)

// initConfigTable creates the app_config key-value store in fileDB.
//
// app_config holds server-wide settings that must survive restarts and do not
// belong to any single user or agent — starting with the admin password hash
// (MAX-10) and later OIDC provider settings and role config (MAX-11).
func initConfigTable() error {
	stmt := `CREATE TABLE IF NOT EXISTS app_config (
		key        TEXT PRIMARY KEY,
		value      TEXT NOT NULL DEFAULT '',
		updated_at INTEGER NOT NULL
	)`
	if _, err := fileDB.Exec(stmt); err != nil {
		return fmt.Errorf("initConfigTable: %w", err)
	}
	return nil
}

// getConfig returns the value for key. found is false when the key is absent
// (which is not an error — callers use it to detect first-time setup).
func getConfig(key string) (value string, found bool, err error) {
	err = fileDB.QueryRow(`SELECT value FROM app_config WHERE key = ?`, key).Scan(&value)
	switch err {
	case sql.ErrNoRows:
		return "", false, nil
	case nil:
		return value, true, nil
	default:
		return "", false, fmt.Errorf("getConfig(%s): %w", key, err)
	}
}

// setConfig upserts a config value, stamping updated_at.
func setConfig(key, value string) error {
	if _, err := fileDB.Exec(
		`INSERT INTO app_config (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, time.Now().Unix(),
	); err != nil {
		return fmt.Errorf("setConfig(%s): %w", key, err)
	}
	return nil
}
