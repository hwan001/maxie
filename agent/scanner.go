package main

import (
	"encoding/hex"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeebo/blake3"
)

// skipDirs are directories that contain generated/third-party artifacts and
// should not be indexed as user files.
var skipDirs = map[string]struct{}{
	"node_modules":  {},
	".git":          {},
	".svn":          {},
	"vendor":        {},
	"__pycache__":   {},
	".venv":         {},
	"venv":          {},
	"env":           {},
	".env":          {},
	"target":        {},
	".gradle":       {},
	"build":         {},
	"dist":          {},
	".next":         {},
	".nuxt":         {},
	".output":       {},
	"Pods":          {},
	".dart_tool":    {},
	".pub-cache":    {},
	"bower_components": {},
	"jspm_packages": {},
	".cargo":        {},
	".cache":        {},
	".idea":         {},
	".vscode":       {},
	"__MACOSX":      {},
	".Trash":        {},
}

func scanDrive(drive DriveEntry) []FileRecord {
	log.Printf("scanning drive: %s (%s)", drive.Label, drive.Path)

	excludeDirs := make(map[string]struct{}, len(skipDirs)+len(drive.ExcludeDirs))
	for k, v := range skipDirs {
		excludeDirs[k] = v
	}
	for _, d := range drive.ExcludeDirs {
		excludeDirs[d] = struct{}{}
	}

	excludeExts := make(map[string]struct{}, len(drive.ExcludeExts))
	for _, ext := range drive.ExcludeExts {
		e := strings.ToLower(ext)
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		excludeExts[e] = struct{}{}
	}

	var records []FileRecord
	filepath.WalkDir(drive.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if _, skip := excludeDirs[d.Name()]; skip {
				return fs.SkipDir
			}
			return nil
		}

		if len(excludeExts) > 0 {
			if _, skip := excludeExts[strings.ToLower(filepath.Ext(path))]; skip {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		size := info.Size()
		modTime := info.ModTime()

		cached, err := getCachedFile(path)
		if err == nil && cached != nil &&
			cached.Size == size &&
			cached.ModifiedAt.Equal(modTime) {
			records = append(records, *cached)
			return nil
		}

		hash, err := hashFile(path)
		if err != nil {
			return nil
		}

		rec := FileRecord{
			FullPath:   path,
			Size:       size,
			ModifiedAt: modTime,
			CreatedAt:  fileCreatedAt(info),
			Hash:       hash,
			DriveType:  drive.DriveType,
			SyncedAt:   time.Now(),
		}
		upsertFile(rec)
		records = append(records, rec)
		return nil
	})
	return records
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := blake3.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
