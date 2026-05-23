package main

import (
	"database/sql"
	"fmt"
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
	return nil
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
	offset := (q.Page - 1) * q.Limit

	where := "1=1"
	args := []interface{}{}
	if q.AgentID != "" {
		where += " AND agent_id = ?"
		args = append(args, q.AgentID)
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

func getDuplicateGroups(agentID string) ([]DuplicateGroup, error) {
	where := ""
	args := []interface{}{}
	if agentID != "" {
		where = "WHERE agent_id = ?"
		args = append(args, agentID)
	}

	rows, err := fileDB.Query(
		fmt.Sprintf("SELECT hash, size, COUNT(*) as cnt FROM files %s GROUP BY hash HAVING cnt > 1 ORDER BY size DESC", where),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []DuplicateGroup
	for rows.Next() {
		var g DuplicateGroup
		if err := rows.Scan(&g.Hash, &g.Size, &g.Count); err != nil {
			continue
		}
		groups = append(groups, g)
	}

	for i, g := range groups {
		fileWhere := "hash = ?"
		fileArgs := []interface{}{g.Hash}
		if agentID != "" {
			fileWhere += " AND agent_id = ?"
			fileArgs = append(fileArgs, agentID)
		}
		frows, err := fileDB.Query(
			fmt.Sprintf("SELECT agent_id, fullpath, size, modified_at, created_at, hash, drive_type, synced_at FROM files WHERE %s ORDER BY modified_at DESC", fileWhere),
			fileArgs...,
		)
		if err != nil {
			continue
		}
		for frows.Next() {
			var r ServerFileRecord
			var modAt, crAt, syncAt int64
			if err := frows.Scan(&r.AgentID, &r.FullPath, &r.Size, &modAt, &crAt, &r.Hash, &r.DriveType, &syncAt); err != nil {
				continue
			}
			r.ModifiedAt = time.Unix(modAt, 0)
			r.CreatedAt = time.Unix(crAt, 0)
			r.SyncedAt = time.Unix(syncAt, 0)
			groups[i].Files = append(groups[i].Files, r)
		}
		frows.Close()
	}

	if groups == nil {
		groups = []DuplicateGroup{}
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
