package cache

type Cacher interface {
	Set(key string, data any) error
	Get(key string) (any, error)
	Delete(key string) error
}
