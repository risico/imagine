package imagine

import (
	"fmt"
	"os"

	"github.com/juju/errors"
)

type LocalStorage struct {
	path string
}

var _ Storage = new(LocalStorage)

func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{path}
}

func (l *LocalStorage) Set(filename string, data []byte) error {
	path := fmt.Sprintf("%s/%s", l.path, filename)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return errors.Annotate(err, "storage.Set: could not write file")
	}

	return nil
}

func (l *LocalStorage) Get(filename string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s", l.path, filename)
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Annotate(err, "storage.Get: could not read file")
	}
	return dat, nil
}

func (l *LocalStorage) Delete(filename string) error {
	path := fmt.Sprintf("%s/%s", l.path, filename)
	return errors.Annotate(os.Remove(path), "storage.Delete: could not delete file")
}
