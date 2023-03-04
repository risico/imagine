package cache

import "sync"

// InMemoryCache caches the files directly in memory
type InMemoryCache struct {
	params *InMemoryCacheParams
	mu     *sync.RWMutex
	cache  map[string][]byte
}

type InMemoryCacheParams struct{}

func NewInMemoryCache(params InMemoryCacheParams) Cacher {
	return &InMemoryCache{
		params: &params,
		cache:  make(map[string][]byte),
		mu:     new(sync.RWMutex),
	}
}

func (s *InMemoryCache) Set(key string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[key] = data
	return nil
}

func (s *InMemoryCache) Get(key string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if data, ok := s.cache[key]; ok {
		return data, ok, nil
	}

	return nil, false, nil
}

func (s *InMemoryCache) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cache, key)

	return nil
}

var _ Cacher = new(InMemoryCache)
