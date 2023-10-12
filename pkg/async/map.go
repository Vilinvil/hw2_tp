package async

import "sync"

type Map[keyT comparable, valT any] struct {
	storage map[keyT]valT
	rwMu    sync.RWMutex
}

func New[keyT comparable, valT any](capacity int) *Map[keyT, valT] {
	return &Map[keyT, valT]{storage: make(map[keyT]valT, capacity)}
}

// Insert - insert element with key and val in map if it not exists. Replace element if element with same key already exist
func (m *Map[keyT, valT]) Insert(key keyT, val valT) {
	m.rwMu.Lock()
	m.storage[key] = val
	m.rwMu.Unlock()
}

// Get - get element by key from map and return this element
func (m *Map[keyT, valT]) Get(key keyT) (valT, bool) {
	m.rwMu.RLock()
	val, ok := m.storage[key]
	m.rwMu.RUnlock()

	return val, ok
}
