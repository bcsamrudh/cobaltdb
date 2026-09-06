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

func TestReplayDiscardsIncompleteFinalRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.wal")
	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	want := Record{Operation: Set, Key: "name", Value: "samrudh"}
	if err := log.Append(want); err != nil {
		t.Fatal(err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(`{"operation":"SET","key":"partial`); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	recovered, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	var got []Record
	if err := recovered.Replay(func(record Record) error {
		got = append(got, record)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []Record{want}) {
		t.Fatalf("Replay() = %#v, want %#v", got, []Record{want})
	}
	if err := recovered.Append(Record{Operation: Set, Key: "language", Value: "go"}); err != nil {
		t.Fatal(err)
	}
	if err := recovered.Close(); err != nil {
		t.Fatal(err)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), "partial") {
		t.Fatalf("incomplete tail was not removed: %q", contents)
	}
}

func TestReplayRejectsCorruptionBeforeEnd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cobalt.wal")
	contents := "{\"operation\":\"SET\",\"key\":\"a\",\"value\":\"1\"}\n" +
		"not-json\n" +
		"{\"operation\":\"SET\",\"key\":\"b\",\"value\":\"2\"}\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	err = log.Replay(func(Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("Replay() error = %v, want corruption error at line 2", err)
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
