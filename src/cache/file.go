package cache

// FileCache stores the cache data in a local db
type FileCache struct{}

// NewFileCache returns a new FileCache
func NewFileCache() *FileCache {
	return &FileCache{}
}

func (f *FileCache) Set(key string, value []byte) error {
	return nil
}

func (f *FileCache) Get(key string) ([]byte, bool, error) {
	return nil, false, nil
}

func (f *FileCache) Delete(key string) error {
	return nil
}

var _ Cacher = new(FileCache)
