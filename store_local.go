package imagine

import (
	"fmt"
	"os"

	"github.com/juju/errors"
)

type localStore struct {
	path string
}

var _ Store = new(localStore)

func NewLocalStorage(path string) Store {
	return &localStore{path}
}

func (l *localStore) Set(filename string, data []byte) error {
	path := fmt.Sprintf("%s/%s", l.path, filename)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return errors.Annotate(err, "storage.Set: could not write file")
	}

	return nil
}

func (l *localStore) Get(filename string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s", l.path, filename)
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Annotate(err, "storage.Get: could not read file")
	}
	return dat, nil
}

func (l *localStore) Delete(filename string) error {
	path := fmt.Sprintf("%s/%s", l.path, filename)
	return errors.Annotate(os.Remove(path), "storage.Delete: could not delete file")
}

func (l *localStore) Close() error {
	return nil
}
