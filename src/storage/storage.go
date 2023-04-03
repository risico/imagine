package storage

// there might be packages out there that can abstract localhost + s3
type Storage interface {
	Set(filename string, data []byte) error
	Get(filename string) (data []byte, err error)
	Delete(filename string) error
}
