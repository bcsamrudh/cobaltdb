package wal

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAppendAndReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.wal")
	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	records := []Record{
		{Operation: Set, Key: "name", Value: "samrudh"},
		{Operation: Set, Key: "greeting", Value: "hello world"},
		{Operation: Delete, Key: "name"},
	}
	for _, record := range records {
		if err := log.Append(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	var got []Record
	if err := reopened.Replay(func(record Record) error {
		got = append(got, record)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, records) {
		t.Fatalf("Replay() = %#v, want %#v", got, records)
	}
}

func TestReplayRejectsCorruptRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.wal")
	if err := os.WriteFile(path, []byte("not-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	err = log.Replay(func(Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("Replay() error = %v, want corruption error at line 1", err)
	}
}

func TestAppendRejectsInvalidRecord(t *testing.T) {
	log, err := Open(filepath.Join(t.TempDir(), "cobalt.wal"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	if err := log.Append(Record{Operation: "UNKNOWN", Key: "name"}); err == nil {
		t.Fatal("Append() accepted an unknown operation")
	}
}
