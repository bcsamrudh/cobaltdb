package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	s := New()
	if err := s.Set("name", "samrudh"); err != nil {
		t.Fatal(err)
	}
	value, exists := s.Get("name")
	if !exists || value != "samrudh" {
		t.Fatalf("Get(name) = %q, %v; want %q, true", value, exists, "samrudh")
	}
}

func TestSetOverwritesExistingValue(t *testing.T) {
	s := New()
	if err := s.Set("language", "rust"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("language", "go"); err != nil {
		t.Fatal(err)
	}
	value, exists := s.Get("language")
	if !exists || value != "go" {
		t.Fatalf("Get(language) = %q, %v; want %q, true", value, exists, "go")
	}
}

func TestGetNonexistentKey(t *testing.T) {
	s := New()
	value, exists := s.Get("missing")
	if exists || value != "" {
		t.Fatalf("Get(missing) = %q, %v; want empty value, false", value, exists)
	}
}

func TestDelete(t *testing.T) {
	s := New()
	if err := s.Set("language", "go"); err != nil {
		t.Fatal(err)
	}
	deleted, err := s.Delete("language")
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatal("Delete(language) = false, want true")
	}
	if _, exists := s.Get("language"); exists {
		t.Fatal("language still exists after deletion")
	}
	deleted, err = s.Delete("language")
	if err != nil {
		t.Fatal(err)
	}
	if deleted {
		t.Fatal("deleting a missing key returned true")
	}
}

func TestExists(t *testing.T) {
	s := New()
	if s.Exists("name") {
		t.Fatal("missing name unexpectedly exists")
	}
	if err := s.Set("name", "samrudh"); err != nil {
		t.Fatal(err)
	}
	if !s.Exists("name") {
		t.Fatal("stored name does not exist")
	}
	if _, err := s.Delete("name"); err != nil {
		t.Fatal(err)
	}
	if s.Exists("name") {
		t.Fatal("deleted name still exists")
	}
}

func TestConcurrentAccess(t *testing.T) {
	const workers = 100
	const operations = 1_000

	s := New()
	var wg sync.WaitGroup
	wg.Add(workers)
	for worker := 0; worker < workers; worker++ {
		go func(worker int) {
			defer wg.Done()
			key := fmt.Sprintf("worker-%d", worker)
			for operation := 0; operation < operations; operation++ {
				if err := s.Set(key, fmt.Sprintf("value-%d", operation)); err != nil {
					t.Errorf("Set(%s) failed: %v", key, err)
					return
				}
				if _, exists := s.Get(key); !exists {
					t.Errorf("%s disappeared after Set", key)
					return
				}
				if !s.Exists(key) {
					t.Errorf("Exists(%s) = false after Set", key)
					return
				}
			}
			deleted, err := s.Delete(key)
			if err != nil {
				t.Errorf("Delete(%s) failed: %v", key, err)
				return
			}
			if !deleted {
				t.Errorf("Delete(%s) = false, want true", key)
			}
		}(worker)
	}
	wg.Wait()

	for worker := 0; worker < workers; worker++ {
		key := fmt.Sprintf("worker-%d", worker)
		if s.Exists(key) {
			t.Errorf("%s still exists after concurrent test", key)
		}
	}
}

func TestSnapshotIsIndependent(t *testing.T) {
	s := New()
	if err := s.Set("name", "samrudh"); err != nil {
		t.Fatal(err)
	}
	snapshot := s.Snapshot()
	snapshot["name"] = "changed"
	snapshot["extra"] = "value"

	if value, _ := s.Get("name"); value != "samrudh" {
		t.Fatalf("snapshot mutation changed stored value to %q", value)
	}
	if s.Exists("extra") {
		t.Fatal("snapshot mutation added a key to the store")
	}
}
