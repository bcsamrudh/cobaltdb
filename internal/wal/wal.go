// Package wal implements CobaltDB's append-only write-ahead log.
package wal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// Operation identifies a mutation stored in the log.
type Operation string

const (
	Set    Operation = "SET"
	Delete Operation = "DELETE"
)

// Record is one durable database mutation.
type Record struct {
	Operation Operation `json:"operation"`
	Key       string    `json:"key"`
	Value     string    `json:"value,omitempty"`
}

// Log is a concurrency-safe append-only write-ahead log.
type Log struct {
	mu   sync.Mutex
	file *os.File
}

// Open opens or creates a WAL at path.
func Open(path string) (*Log, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open WAL: %w", err)
	}
	return &Log{file: file}, nil
}

// Append writes and fsyncs a record before returning.
func (l *Log) Append(record Record) error {
	if err := validate(record); err != nil {
		return err
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode WAL record: %w", err)
	}
	encoded = append(encoded, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.file.Write(encoded); err != nil {
		return fmt.Errorf("append WAL record: %w", err)
	}
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("sync WAL: %w", err)
	}
	return nil
}

// Replay reads valid records from the beginning of the log in write order.
func (l *Log) Replay(apply func(Record) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek WAL: %w", err)
	}
	reader := bufio.NewReader(l.file)
	line := 0
	var validBytes int64
	for {
		encoded, readErr := reader.ReadBytes('\n')
		if errors.Is(readErr, io.EOF) {
			if len(encoded) > 0 {
				if err := l.truncate(validBytes); err != nil {
					return err
				}
			}
			break
		}
		if readErr != nil {
			return fmt.Errorf("read WAL: %w", readErr)
		}

		line++
		var record Record
		if err := json.Unmarshal(encoded[:len(encoded)-1], &record); err != nil {
			return fmt.Errorf("decode WAL record at line %d: %w", line, err)
		}
		if err := validate(record); err != nil {
			return fmt.Errorf("invalid WAL record at line %d: %w", line, err)
		}
		if err := apply(record); err != nil {
			return fmt.Errorf("apply WAL record at line %d: %w", line, err)
		}
		validBytes += int64(len(encoded))
	}
	_, err := l.file.Seek(0, io.SeekEnd)
	return err
}

// Close closes the log file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

func validate(record Record) error {
	if record.Key == "" {
		return errors.New("key is empty")
	}
	switch record.Operation {
	case Set, Delete:
		return nil
	default:
		return fmt.Errorf("unknown operation %q", record.Operation)
	}
}

func (l *Log) truncate(size int64) error {
	if err := l.file.Truncate(size); err != nil {
		return fmt.Errorf("truncate incomplete WAL tail: %w", err)
	}
	if err := l.file.Sync(); err != nil {
		return fmt.Errorf("sync truncated WAL: %w", err)
	}
	return nil
}
