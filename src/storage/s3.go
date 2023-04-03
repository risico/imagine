package storage

type S3Storage struct{}

var _ Storage = new(S3Storage)

func (l *S3Storage) Set(filename string, data []byte) error {
	return nil
}

func (l *S3Storage) Get(filename string) ([]byte, error) {
	return nil, nil
}

func (l *S3Storage) Delete(filename string) error {
	return nil
}
