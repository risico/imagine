package imagine

import (
	"errors"
	"io"
)

var ErrKeyNotFound = errors.New("key not found")

// Store is an interface for a key-value store. This interface is used by both
// the cache and the storage backends.
// Given the huge overlap between them it makes sense to have a common interface.
// For the Cache layer things like key expiration can be set at the store implementation
// level.
type Store interface {
	Set(key string, data []byte) error
	Get(key string) (data []byte, ok bool, err error)
	Delete(key string) error

	io.Closer
}
