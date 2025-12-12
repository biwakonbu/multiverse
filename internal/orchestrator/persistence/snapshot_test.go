package persistence

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSnapshotRepository_CreateRestore(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, "state")
	snapDir := filepath.Join(tmpDir, "snapshots")

	// Create initial state
	err := os.MkdirAll(stateDir, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(stateDir, "file1.txt"), []byte("original content"), 0644)
	assert.NoError(t, err)

	repo := NewSnapshotRepository(snapDir, stateDir)

	// 1. Create Snapshot
	snap, err := repo.CreateSnapshot("initial state")
	assert.NoError(t, err)
	assert.NotEmpty(t, snap.ID)
	assert.Equal(t, "initial state", snap.Description)

	// Verify snapshot storage
	snapContentPath := filepath.Join(snapDir, snap.ID, "state", "file1.txt")
	content, err := os.ReadFile(snapContentPath)
	assert.NoError(t, err)
	assert.Equal(t, "original content", string(content))

	// 2. Modify State
	err = os.WriteFile(filepath.Join(stateDir, "file1.txt"), []byte("modified content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(stateDir, "file2.txt"), []byte("new file"), 0644)
	assert.NoError(t, err)

	// 3. Restore Snapshot
	err = repo.RestoreSnapshot(snap.ID)
	assert.NoError(t, err)

	// Verify state restored
	content, err = os.ReadFile(filepath.Join(stateDir, "file1.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "original content", string(content))

	// Verify new file gone
	_, err = os.Stat(filepath.Join(stateDir, "file2.txt"))
	assert.True(t, os.IsNotExist(err))

	// Verify safety backup created
	entries, _ := os.ReadDir(snapDir)
	foundBackup := false
	for _, e := range entries {
		if len(e.Name()) > 6 && e.Name()[:6] == "backup" {
			foundBackup = true
			break
		}
	}
	assert.True(t, foundBackup, "safety backup should be created")
}

func TestSnapshotRepository_List(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewSnapshotRepository(filepath.Join(tmpDir, "snapshots"), filepath.Join(tmpDir, "state"))
	_ = os.MkdirAll(filepath.Join(tmpDir, "state"), 0755)

	// Create 3 snapshots
	_, _ = repo.CreateSnapshot("snap 1")
	time.Sleep(10 * time.Millisecond)
	_, _ = repo.CreateSnapshot("snap 2")
	time.Sleep(10 * time.Millisecond)
	_, _ = repo.CreateSnapshot("snap 3")

	// List
	snaps, err := repo.ListSnapshots()
	assert.NoError(t, err)
	assert.Len(t, snaps, 3)

	// Check order (desc)
	assert.Equal(t, "snap 3", snaps[0].Description)
	assert.Equal(t, "snap 2", snaps[1].Description)
	assert.Equal(t, "snap 1", snaps[2].Description)
}
