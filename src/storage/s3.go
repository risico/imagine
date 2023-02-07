package storage

type S3Storage struct{}

var _ FS = &S3Storage{}

func (l *S3Storage) Upload(filename string, data []byte) error {
	return nil
}

func (l *S3Storage) Download(filename string) ([]byte, error) {
	return nil, nil
}

func (l *S3Storage) Delete(filename string) error {
	return nil
}
