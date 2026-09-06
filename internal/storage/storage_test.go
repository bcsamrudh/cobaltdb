package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecovery(t *testing.T) {
	directory := t.TempDir()
	first, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Set("name", "samrudh"); err != nil {
		t.Fatal(err)
	}
	if err := first.Set("language", "go"); err != nil {
		t.Fatal(err)
	}
	if _, err := first.Delete("name"); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	recovered, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	if recovered.Exists("name") {
		t.Fatal("deleted key was restored")
	}
	if value, found := recovered.Get("language"); !found || value != "go" {
		t.Fatalf("Get(language) = %q, %v; want go, true", value, found)
	}
}

func TestDeleteMissingKeyIsNotPersisted(t *testing.T) {
	directory := t.TempDir()
	database, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := database.Delete("missing")
	if err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Fatal("Delete(missing) = true, want false")
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryFromIncompleteWALTail(t *testing.T) {
	directory := t.TempDir()
	database, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Set("language", "go"); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	file, err := os.OpenFile(filepath.Join(directory, "cobalt.wal"), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(`{"operation":"DELETE","key":`); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	recovered, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	if value, found := recovered.Get("language"); !found || value != "go" {
		t.Fatalf("Get(language) = %q, %v; want go, true", value, found)
	}
}

func TestSnapshotCompactsWALAndReplaysNewerRecords(t *testing.T) {
	directory := t.TempDir()
	database, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Set("name", "samrudh"); err != nil {
		t.Fatal(err)
	}
	if err := database.Snapshot(); err != nil {
		t.Fatal(err)
	}
	walPath := filepath.Join(directory, "cobalt.wal")
	info, err := os.Stat(walPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("WAL size after snapshot = %d, want 0", info.Size())
	}
	if err := database.Set("language", "go"); err != nil {
		t.Fatal(err)
	}
	if err := database.log.Close(); err != nil {
		t.Fatal(err)
	}

	recovered, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	if value, found := recovered.Get("name"); !found || value != "samrudh" {
		t.Fatalf("snapshot value = %q, %v; want samrudh, true", value, found)
	}
	if value, found := recovered.Get("language"); !found || value != "go" {
		t.Fatalf("WAL value = %q, %v; want go, true", value, found)
	}
}

func TestOpenRejectsCorruptSnapshot(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "cobalt.snapshot"), []byte("not-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(directory); err == nil {
		t.Fatal("Open() accepted a corrupt snapshot")
	}
}
