package imagine

import (
	"fmt"
	"os"
	"time"

	"github.com/juju/errors"
)

// LocalStoreParams are the parameters for creating a new LocalStore
type LocalStoreParams struct {
    // Path is the path to the directory where the files will be stored
    Path string

    // TTL is the time to live for the file in seconds
    // This is to be set if you want to use this store as a caching mechanism
    TTL time.Duration
}

// localStore is a Store implementation that uses the local filesystem
type localStore struct {
    params *LocalStoreParams

    closeCh chan struct{}
}

// ensure localStore implements Store
var _ Store = new(localStore)

func (l *localStore) Set(filename string, data []byte) error {
	path := fmt.Sprintf("%s/%s", l.params.Path, filename)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return errors.Annotate(err, "storage.Set: could not write file")
	}

	return nil
}

func (l *localStore) Get(filename string) ([]byte, bool, error) {
	path := fmt.Sprintf("%s/%s", l.params.Path, filename)

    // check if file exists
    _, err := os.Stat(path)
    if os.IsNotExist(err) {
        return nil, false, ErrKeyNotFound
    }

	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, false, errors.Annotate(err, "storage.Get: could not read file")
	}

	return dat, true, nil
}

func (l *localStore) Delete(filename string) error {
	path := fmt.Sprintf("%s/%s", l.params.Path, filename)
	return errors.Annotate(os.Remove(path), "storage.Delete: could not delete file")
}

func (l *localStore) Close() error {
    close(l.closeCh)
	return nil
}

func (l *localStore) cleanup() {
    if l.params.TTL == time.Duration(0) {
        return
    }

    go func() {
        ticker := time.NewTicker(l.params.TTL / 2)

        for {
            select {
            case <-l.closeCh:
                ticker.Stop()
                return
            case <-ticker.C:
                // get all files in the directory
                files, err := os.ReadDir(l.params.Path)
                if err != nil {
                    continue
                }

                // delete files that are older than TTL
                for _, file := range files {
                    info, err := file.Info()
                    if err != nil {
                        continue
                    }

                    if time.Since(info.ModTime()) > l.params.TTL {
                        os.Remove(file.Name())
                    }
                }
            }
        }
    }()
}

func NewLocalStorage(params LocalStoreParams) (Store, error) {
    // create the directory if it doesn't exist
    if _, err := os.Stat(params.Path); os.IsNotExist(err) {
        err := os.Mkdir(params.Path, 0755)
        if err != nil {
            return nil, errors.Annotate(err, "storage.NewLocalStorage: could not create directory")
        }
    }

    ls := &localStore{
        params: &params,
        closeCh: make(chan struct{}),
    }

    ls.cleanup()

	return ls, nil
}
