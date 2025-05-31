package dedup

import "sync"

// MemStore mantiene los hashes sólo en RAM (no persiste).
type MemStore struct {
	mu  sync.RWMutex
	set map[string]struct{}
}

// NewMemStore crea un Store sin persistencia: ideal para tests.
func NewMemStore() *MemStore {
	return &MemStore{set: make(map[string]struct{})}
}

func (m *MemStore) Seen(h string) bool {
	m.mu.RLock()
	_, ok := m.set[h]
	m.mu.RUnlock()
	return ok
}

func (m *MemStore) Mark(h string) error {
	m.mu.Lock()
	m.set[h] = struct{}{}
	m.mu.Unlock()
	return nil
}

func (m *MemStore) Close() error { return nil }
