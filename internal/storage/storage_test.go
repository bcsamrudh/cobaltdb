package storage

import "testing"

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
