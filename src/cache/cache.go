package cache

type Cacher interface {
	Set(key string, data []byte) error
	Get(key string) ([]byte, bool, error)
	Delete(key string) error
}
