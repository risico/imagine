package imagine

type redisStore struct {}

func (r *redisStore) Set(key string, data []byte) error {
    return nil
}

func (r *redisStore) Get(key string) (data []byte, ok bool, err error) {
    return nil, false, nil
}

func (r *redisStore) Delete(key string) error {
    return nil
}

func (r *redisStore) Close() error {
    return nil
}

func NewRedisStore() Store {
    return &redisStore{}
}
