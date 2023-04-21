package imagine

import (
	"sync"
	"time"
)

type MemoryStoreParams struct {
	TTL time.Duration
}

// InMemoryStorage is a storage implementation that stores data in memory
// This should only be used for testing
type MemoryStore struct {
	params *MemoryStoreParams
	mu     *sync.RWMutex
	cache  map[string][]byte
}

var _ Store = new(MemoryStore)

func NewInMemoryStorage(params MemoryStoreParams) Store {
	return &MemoryStore{
		params: &params,
		cache:  make(map[string][]byte),
		mu:     new(sync.RWMutex),
	}
}

func (m *MemoryStore) Set(key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[key] = data
	return nil
}


func (m *MemoryStore) Get(key string) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if data, ok := m.cache[key]; ok {
		return data, true, nil
	}

	return nil, false, nil
}

func (m *MemoryStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cache, key)

	return nil
}

func (m *MemoryStore) Close() error {
	return nil
}
