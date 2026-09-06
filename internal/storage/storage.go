// Package storage combines the in-memory store with durable write-ahead logging.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/bcsamrudh/cobaltdb/internal/snapshot"
	"github.com/bcsamrudh/cobaltdb/internal/store"
	"github.com/bcsamrudh/cobaltdb/internal/wal"
)

// Store persists mutations before applying them in memory.
type Store struct {
	mutations sync.Mutex
	memory    *store.Store
	log       *wal.Log
	snapshot  string
}

// Open creates or recovers a persistent Store in dataDirectory.
func Open(dataDirectory string) (*Store, error) {
	if err := os.MkdirAll(dataDirectory, 0o750); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	snapshotPath := filepath.Join(dataDirectory, "cobalt.snapshot")
	entries, err := snapshot.Load(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("recover database: %w", err)
	}
	memory := store.New()
	for key, value := range entries {
		if err := memory.Set(key, value); err != nil {
			return nil, fmt.Errorf("restore snapshot: %w", err)
		}
	}

	log, err := wal.Open(filepath.Join(dataDirectory, "cobalt.wal"))
	if err != nil {
		return nil, err
	}
	persistent := &Store{memory: memory, log: log, snapshot: snapshotPath}
	if err := log.Replay(persistent.apply); err != nil {
		log.Close()
		return nil, fmt.Errorf("recover database: %w", err)
	}
	return persistent, nil
}

// Snapshot atomically saves the current state and resets the covered WAL.
func (s *Store) Snapshot() error {
	s.mutations.Lock()
	defer s.mutations.Unlock()
	return s.snapshotLocked()
}

// Set durably records and applies a value.
func (s *Store) Set(key, value string) error {
	s.mutations.Lock()
	defer s.mutations.Unlock()
	if err := s.log.Append(wal.Record{Operation: wal.Set, Key: key, Value: value}); err != nil {
		return err
	}
	return s.memory.Set(key, value)
}

// Get returns a value and whether it exists.
func (s *Store) Get(key string) (string, bool) {
	return s.memory.Get(key)
}

// Delete durably records and removes an existing key.
func (s *Store) Delete(key string) (bool, error) {
	s.mutations.Lock()
	defer s.mutations.Unlock()
	if !s.memory.Exists(key) {
		return false, nil
	}
	if err := s.log.Append(wal.Record{Operation: wal.Delete, Key: key}); err != nil {
		return false, err
	}
	return s.memory.Delete(key)
}

// Exists reports whether key exists.
func (s *Store) Exists(key string) bool {
	return s.memory.Exists(key)
}

// Close snapshots current state, compacts the WAL, and closes the log.
func (s *Store) Close() error {
	s.mutations.Lock()
	defer s.mutations.Unlock()
	snapshotErr := s.snapshotLocked()
	closeErr := s.log.Close()
	if snapshotErr != nil {
		return snapshotErr
	}
	return closeErr
}

func (s *Store) snapshotLocked() error {
	if err := snapshot.Save(s.snapshot, s.memory.Snapshot()); err != nil {
		return err
	}
	if err := s.log.Reset(); err != nil {
		return fmt.Errorf("compact WAL: %w", err)
	}
	return nil
}

func (s *Store) apply(record wal.Record) error {
	switch record.Operation {
	case wal.Set:
		return s.memory.Set(record.Key, record.Value)
	case wal.Delete:
		_, err := s.memory.Delete(record.Key)
		return err
	default:
		return fmt.Errorf("unsupported operation %q", record.Operation)
	}
}
