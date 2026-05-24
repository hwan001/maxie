package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var fileDB *sql.DB

type ServerFileRecord struct {
	AgentID    string    `json:"agent_id"`
	AgentName  string    `json:"agent_name,omitempty"`
	FullPath   string    `json:"fullpath"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
	CreatedAt  time.Time `json:"created_at"`
	Hash       string    `json:"hash"`
	DriveType  string    `json:"drive_type"`
	SyncedAt   time.Time `json:"synced_at"`
}

type AgentFileRecord struct {
	FullPath   string `json:"fullpath"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modified_at"`
	CreatedAt  int64  `json:"created_at"`
	Hash       string `json:"hash"`
	DriveType  string `json:"drive_type"`
	SyncedAt   int64  `json:"synced_at"`
}

type PendingAction struct {
	ID        int64     `json:"id"`
	AgentID   string    `json:"agent_id"`
	FullPath  string    `json:"fullpath"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

type DuplicateGroup struct {
	Hash  string             `json:"hash"`
	Size  int64              `json:"size"`
	Count int                `json:"count"`
	Files []ServerFileRecord `json:"files"`
}

type FileQuery struct {
	AgentID   string
	AgentIDs  []string // restrict results to this set of agent IDs (user-scoped)
	DriveType string
	Search    string
	Page      int
	Limit     int
	SortBy    string
	SortDir   string
}

type FileQueryResult struct {
	Files []ServerFileRecord `json:"files"`
	Total int                `json:"total"`
}

func fileDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".fileoptimizer", "files.db")
}

func initFileDB() error {
	path := fileDBPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	var err error
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(10000)&_pragma=synchronous(NORMAL)"
	fileDB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	// Serialize writes to avoid SQLITE_BUSY under concurrent batch pushes.
	fileDB.SetMaxOpenConns(1)

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS files (
			agent_id    TEXT NOT NULL,
			fullpath    TEXT NOT NULL,
			size        INTEGER NOT NULL,
			modified_at INTEGER NOT NULL,
			created_at  INTEGER NOT NULL,
			hash        TEXT NOT NULL,
			drive_type  TEXT NOT NULL,
			synced_at   INTEGER NOT NULL,
			PRIMARY KEY (agent_id, fullpath)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_files_agent ON files(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_files_hash ON files(hash)`,
		`CREATE TABLE IF NOT EXISTS pending_actions (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id    TEXT NOT NULL,
			fullpath    TEXT NOT NULL,
			action      TEXT NOT NULL,
			created_at  INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pending_agent ON pending_actions(agent_id)`,
	}

	for _, stmt := range stmts {
		if _, err := fileDB.Exec(stmt); err != nil {
			return fmt.Errorf("filedb init: %w", err)
		}
	}
	return initUserTables()
}

func upsertFileBatch(agentID string, records []AgentFileRecord) error {
	tx, err := fileDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO files (agent_id, fullpath, size, modified_at, created_at, hash, drive_type, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(agent_id, fullpath) DO UPDATE SET
			size        = excluded.size,
			modified_at = excluded.modified_at,
			hash        = excluded.hash,
			drive_type  = excluded.drive_type,
			synced_at   = excluded.synced_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		if _, err := stmt.Exec(agentID, r.FullPath, r.Size, r.ModifiedAt, r.CreatedAt, r.Hash, r.DriveType, r.SyncedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func queryFiles(q FileQuery) (FileQueryResult, error) {
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Page <= 0 {
		q.Page = 1
	}

	// Safety guard: if no agent scoping provided, return nothing instead of leaking all files.
	if q.AgentID == "" && len(q.AgentIDs) == 0 {
		return FileQueryResult{Files: []ServerFileRecord{}, Total: 0}, nil
	}

	offset := (q.Page - 1) * q.Limit

	where := "1=1"
	args := []interface{}{}
	if q.AgentID != "" {
		where += " AND agent_id = ?"
		args = append(args, q.AgentID)
	} else if len(q.AgentIDs) > 0 {
		placeholders := strings.Repeat("?,", len(q.AgentIDs))
		placeholders = placeholders[:len(placeholders)-1]
		where += " AND agent_id IN (" + placeholders + ")"
		for _, id := range q.AgentIDs {
			args = append(args, id)
		}
	}
	if q.DriveType != "" {
		where += " AND drive_type = ?"
		args = append(args, q.DriveType)
	}
	if q.Search != "" {
		where += " AND fullpath LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	orderCol := "synced_at"
	switch q.SortBy {
	case "size":
		orderCol = "size"
	case "modified_at":
		orderCol = "modified_at"
	case "name":
		orderCol = "fullpath"
	}
	orderDir := "DESC"
	if q.SortDir == "asc" {
		orderDir = "ASC"
	}

	var total int
	countArgs := append([]interface{}{}, args...)
	if err := fileDB.QueryRow("SELECT COUNT(*) FROM files WHERE "+where, countArgs...).Scan(&total); err != nil {
		return FileQueryResult{}, err
	}

	queryArgs := append(args, q.Limit, offset)
	rows, err := fileDB.Query(
		fmt.Sprintf("SELECT agent_id, fullpath, size, modified_at, created_at, hash, drive_type, synced_at FROM files WHERE %s ORDER BY %s %s LIMIT ? OFFSET ?", where, orderCol, orderDir),
		queryArgs...,
	)
	if err != nil {
		return FileQueryResult{}, err
	}
	defer rows.Close()

	var files []ServerFileRecord
	for rows.Next() {
		var r ServerFileRecord
		var modAt, crAt, syncAt int64
		if err := rows.Scan(&r.AgentID, &r.FullPath, &r.Size, &modAt, &crAt, &r.Hash, &r.DriveType, &syncAt); err != nil {
			continue
		}
		r.ModifiedAt = time.Unix(modAt, 0)
		r.CreatedAt = time.Unix(crAt, 0)
		r.SyncedAt = time.Unix(syncAt, 0)
		files = append(files, r)
	}
	if files == nil {
		files = []ServerFileRecord{}
	}
	return FileQueryResult{Files: files, Total: total}, nil
}

func getDuplicateGroups(agentID string, agentIDs []string) ([]DuplicateGroup, error) {
	// Safety guard: if no agent scoping provided, return nothing instead of leaking all files.
	if agentID == "" && len(agentIDs) == 0 {
		return []DuplicateGroup{}, nil
	}

	// Single JOIN query replaces the previous N+1 pattern.
	// d (subquery) finds all hashes that appear more than once, then we JOIN
	// back to get every file row for those hashes in one round-trip.

	// Build WHERE clause based on scoping parameters.
	buildWhere := func() (string, []interface{}) {
		if agentID != "" {
			return "WHERE agent_id = ?", []interface{}{agentID}
		}
		if len(agentIDs) > 0 {
			placeholders := strings.Repeat("?,", len(agentIDs))
			placeholders = placeholders[:len(placeholders)-1]
			args := make([]interface{}, len(agentIDs))
			for i, id := range agentIDs {
				args[i] = id
			}
			return "WHERE agent_id IN (" + placeholders + ")", args
		}
		return "", nil
	}

	innerWhere, innerArgs := buildWhere()
	outerWhere, outerArgs := buildWhere()

	query := fmt.Sprintf(`
		SELECT f.agent_id, f.fullpath, f.size, f.modified_at, f.created_at,
		       f.hash, f.drive_type, f.synced_at, d.cnt
		FROM files f
		JOIN (
			SELECT hash, COUNT(*) AS cnt, MAX(size) AS size
			FROM files
			%s
			GROUP BY hash
			HAVING cnt > 1
		) d ON f.hash = d.hash
		%s
		ORDER BY d.size DESC, f.hash, f.modified_at DESC
	`, innerWhere, outerWhere)

	args := append(innerArgs, outerArgs...)
	rows, err := fileDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("getDuplicateGroups: %w", err)
	}
	defer rows.Close()

	// Build groups map keyed by hash to collect files as we stream rows.
	groupMap := make(map[string]*DuplicateGroup)
	var order []string // preserves insertion order for stable output

	for rows.Next() {
		var r ServerFileRecord
		var modAt, crAt, syncAt int64
		var cnt int

		if err := rows.Scan(
			&r.AgentID, &r.FullPath, &r.Size, &modAt, &crAt,
			&r.Hash, &r.DriveType, &syncAt, &cnt,
		); err != nil {
			continue
		}
		r.ModifiedAt = time.Unix(modAt, 0)
		r.CreatedAt = time.Unix(crAt, 0)
		r.SyncedAt = time.Unix(syncAt, 0)

		g, exists := groupMap[r.Hash]
		if !exists {
			g = &DuplicateGroup{
				Hash:  r.Hash,
				Size:  r.Size,
				Count: cnt,
			}
			groupMap[r.Hash] = g
			order = append(order, r.Hash)
		}
		g.Files = append(g.Files, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getDuplicateGroups rows: %w", err)
	}

	groups := make([]DuplicateGroup, 0, len(order))
	for _, hash := range order {
		groups = append(groups, *groupMap[hash])
	}
	return groups, nil
}

func addPendingAction(agentID, fullpath, action string) error {
	_, err := fileDB.Exec(
		`INSERT INTO pending_actions (agent_id, fullpath, action, created_at) VALUES (?, ?, ?, ?)`,
		agentID, fullpath, action, time.Now().Unix(),
	)
	return err
}

func getPendingActions(agentID string) ([]PendingAction, error) {
	rows, err := fileDB.Query(
		`SELECT id, agent_id, fullpath, action, created_at FROM pending_actions WHERE agent_id = ? ORDER BY id`,
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []PendingAction
	for rows.Next() {
		var a PendingAction
		var crAt int64
		if err := rows.Scan(&a.ID, &a.AgentID, &a.FullPath, &a.Action, &crAt); err != nil {
			continue
		}
		a.CreatedAt = time.Unix(crAt, 0)
		actions = append(actions, a)
	}
	if actions == nil {
		actions = []PendingAction{}
	}
	return actions, nil
}

func clearPendingAction(id int64) error {
	_, err := fileDB.Exec(`DELETE FROM pending_actions WHERE id = ?`, id)
	return err
}

func deleteFileRecord(agentID, fullpath string) error {
	_, err := fileDB.Exec(`DELETE FROM files WHERE agent_id = ? AND fullpath = ?`, agentID, fullpath)
	return err
}

// deleteFilesForDrivePrefix removes all file records whose path starts with the
// given drive prefix (exact match or prefix+"/" + anything underneath).
func deleteFilesForDrivePrefix(agentID, pathPrefix string) error {
	prefix := pathPrefix
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	_, err := fileDB.Exec(
		`DELETE FROM files WHERE agent_id = ? AND (fullpath = ? OR fullpath LIKE ?)`,
		agentID, pathPrefix, prefix+"%",
	)
	return err
}

// ─── User store ───────────────────────────────────────────────────────────────

// initUserTables creates the users and oauth_mappings tables in fileDB.
func initUserTables() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id         TEXT    PRIMARY KEY,
			name       TEXT    NOT NULL DEFAULT '',
			email      TEXT    NOT NULL DEFAULT '',
			picture    TEXT    NOT NULL DEFAULT '',
			is_guest   INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			expires_at INTEGER
		)`,
		// Index for fast expired-guest cleanup.
		`CREATE INDEX IF NOT EXISTS idx_users_guest ON users(is_guest, expires_at)`,
		// oauth_mappings maps (provider, provider_id) → internal user UUID.
		// Keeping it separate lets one user eventually link multiple providers.
		`CREATE TABLE IF NOT EXISTS oauth_mappings (
			provider    TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			user_id     TEXT NOT NULL,
			PRIMARY KEY (provider, provider_id)
		)`,
	}
	for _, stmt := range stmts {
		if _, err := fileDB.Exec(stmt); err != nil {
			return fmt.Errorf("initUserTables: %w", err)
		}
	}
	return nil
}

// upsertOAuthUser finds or creates an internal user for a given OAuth identity.
// On first login a new UUID is minted; subsequent logins update profile fields.
func upsertOAuthUser(provider, providerID, name, email, picture string) (*User, error) {
	var userID string
	err := fileDB.QueryRow(
		`SELECT user_id FROM oauth_mappings WHERE provider = ? AND provider_id = ?`,
		provider, providerID,
	).Scan(&userID)

	switch err {
	case sql.ErrNoRows:
		// First login — create internal user and mapping.
		userID = "u-" + generateUUID()
		if _, err := fileDB.Exec(
			`INSERT INTO users (id, name, email, picture, is_guest, created_at) VALUES (?, ?, ?, ?, 0, ?)`,
			userID, name, email, picture, time.Now().Unix(),
		); err != nil {
			return nil, fmt.Errorf("upsertOAuthUser insert user: %w", err)
		}
		if _, err := fileDB.Exec(
			`INSERT INTO oauth_mappings (provider, provider_id, user_id) VALUES (?, ?, ?)`,
			provider, providerID, userID,
		); err != nil {
			return nil, fmt.Errorf("upsertOAuthUser insert mapping: %w", err)
		}
	case nil:
		// Returning user — refresh profile.
		if _, err := fileDB.Exec(
			`UPDATE users SET name = ?, email = ?, picture = ? WHERE id = ?`,
			name, email, picture, userID,
		); err != nil {
			return nil, fmt.Errorf("upsertOAuthUser update: %w", err)
		}
	default:
		return nil, fmt.Errorf("upsertOAuthUser lookup: %w", err)
	}

	return getUserByID(userID)
}

// createGuestUser mints a guest user with a 24-hour expiry.
func createGuestUser() (*User, error) {
	id := "g-" + generateUUID()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	if _, err := fileDB.Exec(
		`INSERT INTO users (id, name, email, picture, is_guest, created_at, expires_at)
		 VALUES (?, 'Guest', '', '', 1, ?, ?)`,
		id, now.Unix(), expiresAt.Unix(),
	); err != nil {
		return nil, fmt.Errorf("createGuestUser: %w", err)
	}
	return &User{
		ID:        id,
		Name:      "Guest",
		IsGuest:   true,
		CreatedAt: now,
		ExpiresAt: &expiresAt,
	}, nil
}

// getUserByID fetches a user record by its internal UUID.
func getUserByID(id string) (*User, error) {
	var u User
	var createdAt int64
	var expiresAt sql.NullInt64
	var isGuest int
	err := fileDB.QueryRow(
		`SELECT id, name, email, picture, is_guest, created_at, expires_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Picture, &isGuest, &createdAt, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("getUserByID: %w", err)
	}
	u.IsGuest = isGuest == 1
	u.CreatedAt = time.Unix(createdAt, 0)
	if expiresAt.Valid {
		t := time.Unix(expiresAt.Int64, 0)
		u.ExpiresAt = &t
	}
	return &u, nil
}

// cleanupExpiredGuests removes guest users whose expiry has passed, along with
// all of their agents and file records. Safe to call on a ticker.
func cleanupExpiredGuests() {
	rows, err := fileDB.Query(
		`SELECT id FROM users WHERE is_guest = 1 AND expires_at IS NOT NULL AND expires_at < ?`,
		time.Now().Unix(),
	)
	if err != nil {
		log.Printf("cleanupExpiredGuests query: %v", err)
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()

	for _, userID := range ids {
		deleteUserWithAgents(userID)
		log.Printf("cleaned up expired guest %s", userID)
	}
}

// deleteUserWithAgents removes a user and cascades to their agents and files.
func deleteUserWithAgents(userID string) {
	agentMu.Lock()
	var toDelete []string
	for id, a := range agentStore {
		if a.UserID == userID {
			toDelete = append(toDelete, id)
		}
	}
	for _, id := range toDelete {
		delete(agentStore, id)
	}
	agentMu.Unlock()

	if len(toDelete) > 0 {
		go saveAgents()
		for _, agentID := range toDelete {
			if _, err := fileDB.Exec(`DELETE FROM files WHERE agent_id = ?`, agentID); err != nil {
				log.Printf("deleteUserWithAgents files(%s): %v", agentID, err)
			}
			if _, err := fileDB.Exec(`DELETE FROM pending_actions WHERE agent_id = ?`, agentID); err != nil {
				log.Printf("deleteUserWithAgents pending(%s): %v", agentID, err)
			}
		}
	}

	fileDB.Exec(`DELETE FROM oauth_mappings WHERE user_id = ?`, userID)
	fileDB.Exec(`DELETE FROM users WHERE id = ?`, userID)
}
