package storage

// there might be packages out there that can abstract localhost + s3
type FS interface {
	Upload(filename string, data []byte) error
	Download(filename string) (data []byte, err error)
	Delete(filename string) error
}
