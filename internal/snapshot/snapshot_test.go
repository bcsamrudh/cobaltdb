package snapshot

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.snapshot")
	want := map[string]string{"name": "samrudh", "greeting": "hello world"}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestSaveReplacesExistingSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.snapshot")
	if err := Save(path, map[string]string{"old": "value"}); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"new": "value"}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadMissingSnapshot(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "missing.snapshot"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("Load(missing) = %#v, want empty data", got)
	}
}

func TestLoadRejectsCorruptSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.snapshot")
	if err := os.WriteFile(path, []byte("not-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "decode snapshot") {
		t.Fatalf("Load() error = %v, want decode error", err)
	}
}

func TestLoadRejectsUnknownVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.snapshot")
	if err := os.WriteFile(path, []byte(`{"version":2,"data":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported snapshot version") {
		t.Fatalf("Load() error = %v, want version error", err)
	}
}
