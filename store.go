package imagine

import (
	"errors"
	"io"
)

var ErrKeyNotFound = errors.New("key not found")

// there might be packages out there that can abstract localhost + s3
type Store interface {
	Set(key string, data []byte) error
	Get(key string) (data []byte, ok bool, err error)
	Delete(key string) error

	io.Closer
}
