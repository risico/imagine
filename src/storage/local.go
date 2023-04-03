package storage

type LocalStorage struct{}

var _ Storage = new(LocalStorage)

func (l *LocalStorage) Set(filename string, data []byte) error {
	return nil
}

func (l *LocalStorage) Get(filename string) ([]byte, error) {
	return nil, nil
}

func (l *LocalStorage) Delete(filename string) error {
	return nil
}
