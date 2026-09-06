// Package snapshot stores point-in-time copies of CobaltDB data.
package snapshot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const currentVersion = 1

type fileFormat struct {
	Version int               `json:"version"`
	Data    map[string]string `json:"data"`
}

// Load reads a snapshot. A missing snapshot is treated as an empty database.
func Load(path string) (map[string]string, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var snapshot fileFormat
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return nil, err
	}
	if snapshot.Version != currentVersion {
		return nil, fmt.Errorf("unsupported snapshot version %d", snapshot.Version)
	}
	if snapshot.Data == nil {
		return nil, errors.New("snapshot data is missing")
	}
	return snapshot.Data, nil
}

// Save atomically replaces path with a durable snapshot of data.
func Save(path string, data map[string]string) (saveErr error) {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".cobalt-snapshot-*")
	if err != nil {
		return fmt.Errorf("create temporary snapshot: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		temporary.Close()
		if saveErr != nil {
			os.Remove(temporaryPath)
		}
	}()

	if err := temporary.Chmod(0o600); err != nil {
		return fmt.Errorf("set snapshot permissions: %w", err)
	}
	encoder := json.NewEncoder(temporary)
	if err := encoder.Encode(fileFormat{Version: currentVersion, Data: data}); err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync snapshot: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close snapshot: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("publish snapshot: %w", err)
	}

	directoryHandle, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open snapshot directory: %w", err)
	}
	defer directoryHandle.Close()
	if err := directoryHandle.Sync(); err != nil {
		return fmt.Errorf("sync snapshot directory: %w", err)
	}
	return nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("snapshot contains trailing data")
		}
		return fmt.Errorf("decode snapshot trailing data: %w", err)
	}
	return nil
}
