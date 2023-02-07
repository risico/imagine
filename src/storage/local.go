package storage

type LocalStorage struct{}

var _ FS = &LocalStorage{}

func (l *LocalStorage) Upload(filename string, data []byte) error {
	return nil
}

func (l *LocalStorage) Download(filename string) ([]byte, error) {
	return nil, nil
}

func (l *LocalStorage) Delete(filename string) error {
	return nil
}
