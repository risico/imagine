package imagine

import "io"

// there might be packages out there that can abstract localhost + s3
type Store interface {
	Set(key string, data []byte) error
	Get(key string) (data []byte, err error)
	Delete(key string) error

	io.Closer
}
