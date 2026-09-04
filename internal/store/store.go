// Package store provides CobaltDB's in-memory key-value storage engine.
package store

import "sync"

// Store is a concurrency-safe in-memory key-value store.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

// New creates an empty Store.
func New() *Store {
	return &Store{data: make(map[string]string)}
}

// Set associates key with value, replacing any existing value.
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Get returns the value associated with key and whether the key exists.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

// Delete removes key and reports whether it existed.
func (s *Store) Delete(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[key]; !exists {
		return false, nil
	}
	delete(s.data, key)
	return true, nil
}

// Exists reports whether key exists.
func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists
}
