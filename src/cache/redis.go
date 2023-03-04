package cache

type RedisCache struct {
	params *RedisCacheParams
}

type RedisCacheParams struct {
	radixClient *radix.Client
}

func NewRedisCache(params RedisCacheParams) *RedisCache {
	return &RedisCache{
		params: &params,
	}
}

func (r *RedisCache) Set(key string, value []byte) error {
	return nil
}

func (r *RedisCache) Get(key string) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *RedisCache) Delete(key string) error {
	return nil
}

var _ Cacher = new(RedisCache)
