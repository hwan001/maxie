package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", dbPath())
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			fullpath    TEXT PRIMARY KEY,
			size        INTEGER NOT NULL,
			modified_at INTEGER NOT NULL,
			created_at  INTEGER NOT NULL,
			hash        TEXT NOT NULL,
			drive_type  TEXT NOT NULL,
			synced_at   INTEGER NOT NULL
		)
	`)
	return err
}

type FileRecord struct {
	FullPath   string
	Size       int64
	ModifiedAt time.Time
	CreatedAt  time.Time
	Hash       string
	DriveType  string
	SyncedAt   time.Time
}

func upsertFile(r FileRecord) error {
	_, err := db.Exec(`
		INSERT INTO files (fullpath, size, modified_at, created_at, hash, drive_type, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(fullpath) DO UPDATE SET
			size        = excluded.size,
			modified_at = excluded.modified_at,
			hash        = excluded.hash,
			drive_type  = excluded.drive_type,
			synced_at   = excluded.synced_at
	`,
		r.FullPath,
		r.Size,
		r.ModifiedAt.Unix(),
		r.CreatedAt.Unix(),
		r.Hash,
		r.DriveType,
		r.SyncedAt.Unix(),
	)
	return err
}

func getAllCachedFiles() ([]FileRecord, error) {
	rows, err := db.Query(`SELECT fullpath, size, modified_at, created_at, hash, drive_type, synced_at FROM files`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []FileRecord
	for rows.Next() {
		var r FileRecord
		var modAt, crAt, syncAt int64
		if err := rows.Scan(&r.FullPath, &r.Size, &modAt, &crAt, &r.Hash, &r.DriveType, &syncAt); err != nil {
			continue
		}
		r.ModifiedAt = time.Unix(modAt, 0)
		r.CreatedAt = time.Unix(crAt, 0)
		r.SyncedAt = time.Unix(syncAt, 0)
		records = append(records, r)
	}
	return records, rows.Err()
}

func getCachedFile(fullpath string) (*FileRecord, error) {
	row := db.QueryRow(`SELECT fullpath, size, modified_at, created_at, hash, drive_type, synced_at FROM files WHERE fullpath = ?`, fullpath)
	var r FileRecord
	var modAt, crAt, syncAt int64
	err := row.Scan(&r.FullPath, &r.Size, &modAt, &crAt, &r.Hash, &r.DriveType, &syncAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.ModifiedAt = time.Unix(modAt, 0)
	r.CreatedAt = time.Unix(crAt, 0)
	r.SyncedAt = time.Unix(syncAt, 0)
	return &r, nil
}

type FileStats struct {
	TotalFiles     int       `json:"total_files"`
	TotalSize      int64     `json:"total_size"`
	DuplicateCount int       `json:"duplicate_count"`
	LastScanned    time.Time `json:"last_scanned"`
}

func computeFileStats(driveType string) FileStats {
	var stats FileStats

	query := `SELECT COUNT(*), COALESCE(SUM(size), 0) FROM files`
	args := []interface{}{}
	if driveType != "" {
		query += ` WHERE drive_type = ?`
		args = append(args, driveType)
	}
	db.QueryRow(query, args...).Scan(&stats.TotalFiles, &stats.TotalSize)

	dupQuery := `SELECT COUNT(*) FROM (SELECT hash FROM files GROUP BY hash HAVING COUNT(*) > 1)`
	db.QueryRow(dupQuery).Scan(&stats.DuplicateCount)

	var maxSynced int64
	db.QueryRow(`SELECT COALESCE(MAX(synced_at), 0) FROM files`).Scan(&maxSynced)
	if maxSynced > 0 {
		stats.LastScanned = time.Unix(maxSynced, 0)
	}

	return stats
}
