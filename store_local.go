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
	fmt.Printf("[LocalStore] Attempting to write %d bytes to %s\n", len(data), path)
	
	// Check if directory exists
	if _, err := os.Stat(l.params.Path); os.IsNotExist(err) {
		fmt.Printf("[LocalStore] Directory %s does not exist, attempting to create\n", l.params.Path)
		if err := os.MkdirAll(l.params.Path, 0755); err != nil {
			fmt.Printf("[LocalStore] Failed to create directory: %v\n", err)
			return errors.Annotate(err, "storage.Set: could not create directory")
		}
	}
	
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		fmt.Printf("[LocalStore] Failed to write file: %v\n", err)
		return errors.Annotate(err, "storage.Set: could not write file")
	}
	
	fmt.Printf("[LocalStore] Successfully wrote file to %s\n", path)
	return nil
}

func (l *localStore) Get(filename string) ([]byte, bool, error) {
	path := fmt.Sprintf("%s/%s", l.params.Path, filename)
	fmt.Printf("[LocalStore] Get called for path: %s\n", path)

    // check if file exists
    _, err := os.Stat(path)
    if os.IsNotExist(err) {
        fmt.Printf("[LocalStore] File does not exist: %s\n", path)
        return nil, false, ErrKeyNotFound
    }

	dat, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("[LocalStore] Failed to read file: %v\n", err)
		return nil, false, errors.Annotate(err, "storage.Get: could not read file")
	}

	fmt.Printf("[LocalStore] Successfully read %d bytes from %s\n", len(dat), path)
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
    fmt.Printf("[LocalStore] Creating new LocalStorage with path: %s, TTL: %v\n", params.Path, params.TTL)
    
    // create the directory if it doesn't exist
    if _, err := os.Stat(params.Path); os.IsNotExist(err) {
        fmt.Printf("[LocalStore] Directory %s does not exist, creating\n", params.Path)
        err := os.MkdirAll(params.Path, 0755)
        if err != nil {
            fmt.Printf("[LocalStore] Failed to create directory: %v\n", err)
            return nil, errors.Annotate(err, "storage.NewLocalStorage: could not create directory")
        }
        fmt.Printf("[LocalStore] Successfully created directory %s\n", params.Path)
    } else {
        fmt.Printf("[LocalStore] Directory %s already exists\n", params.Path)
    }

    ls := &localStore{
        params: &params,
        closeCh: make(chan struct{}),
    }

    ls.cleanup()

	return ls, nil
}
