package persistence

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
)

type Snapshot struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type SnapshotRepository interface {
	CreateSnapshot(description string) (*Snapshot, error)
	RestoreSnapshot(snapshotID string) error
	ListSnapshots() ([]Snapshot, error)
}

type snapshotRepoImpl struct {
	baseDir  string // "snapshots" dir
	stateDir string // "state" dir to backup/restore
}

func NewSnapshotRepository(baseDir, stateDir string) SnapshotRepository {
	return &snapshotRepoImpl{
		baseDir:  baseDir,
		stateDir: stateDir,
	}
}

func (r *snapshotRepoImpl) CreateSnapshot(description string) (*Snapshot, error) {
	id := fmt.Sprintf("%s-%s", time.Now().Format("20060102-150405"), uuid.New().String()[:8])
	snapDir := filepath.Join(r.baseDir, id)

	if err := os.MkdirAll(snapDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot dir: %w", err)
	}

	// 1. Copy state directory
	if err := copyDir(r.stateDir, filepath.Join(snapDir, "state")); err != nil {
		return nil, fmt.Errorf("failed to copy state: %w", err)
	}

	// 2. Save metadata
	snap := &Snapshot{
		ID:          id,
		Description: description,
		CreatedAt:   time.Now(),
	}
	metaPath := filepath.Join(snapDir, "snapshot.json")
	if err := writeJSON(metaPath, snap); err != nil {
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return snap, nil
}

func (r *snapshotRepoImpl) RestoreSnapshot(snapshotID string) error {
	snapDir := filepath.Join(r.baseDir, snapshotID)
	if _, err := os.Stat(snapDir); os.IsNotExist(err) {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	// 1. Safety Backup
	backupID := fmt.Sprintf("backup-pre-restore-%s", time.Now().Format("20060102-150405"))
	backupDir := filepath.Join(r.baseDir, backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create safety backup dir: %w", err)
	}
	if err := copyDir(r.stateDir, filepath.Join(backupDir, "state")); err != nil {
		return fmt.Errorf("failed to create safety backup: %w", err)
	}

	// 2. Clear current state
	if err := os.RemoveAll(r.stateDir); err != nil {
		return fmt.Errorf("failed to clear current state: %w", err)
	}

	// 3. Restore from snapshot
	if err := copyDir(filepath.Join(snapDir, "state"), r.stateDir); err != nil {
		return fmt.Errorf("failed to restore state: %w", err)
	}

	return nil
}

func (r *snapshotRepoImpl) ListSnapshots() ([]Snapshot, error) {
	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, err
	}

	var snapshots []Snapshot
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Ignore safety backups
		if len(entry.Name()) > 6 && entry.Name()[:6] == "backup" {
			continue
		}

		metaPath := filepath.Join(r.baseDir, entry.Name(), "snapshot.json")
		var snap Snapshot
		if err := readJSON(metaPath, &snap); err != nil {
			// Skip invalid snapshots
			continue
		}
		snapshots = append(snapshots, snap)
	}

	// Sort by CreatedAt desc
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.After(snapshots[j].CreatedAt)
	})

	return snapshots, nil
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist.
func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.MkdirAll(dst, si.Mode()); err != nil {
			return err
		}
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err = copyDir(srcPath, dstPath); err != nil {
				return
			}
		} else {
			if err = copyFile(srcPath, dstPath); err != nil {
				return
			}
		}
	}
	return
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() {
		if e := in.Close(); e != nil && err == nil {
			err = e
		}
	}()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil && err == nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	return
}
