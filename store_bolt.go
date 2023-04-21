package imagine

import (
	"github.com/juju/errors"
	bolt "go.etcd.io/bbolt"
)


// boltStore is a Store implementation that uses BoltDB
type boltStore struct {
	params *BoltStoreParams
	db     *bolt.DB
}


// BoltStoreParams are the parameters for creating a new BoltStore
type BoltStoreParams struct {
	Path string
}

// ensure boltStore implements Store
var _ Store = new(boltStore)

// NewBoltStore creates a new BoltStore
func NewBoltStore(params BoltStoreParams) (Store, error) {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open(params.Path, 0600, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &boltStore{
		params: &params,
		db:     db,
	}, nil
}

func (b *boltStore) Set(filename string, data []byte) error {
	return nil
}


func (b *boltStore) Get(filename string) ([]byte, bool, error) {
	return nil, false, nil
}

func (b *boltStore) Delete(filename string) error {
	return nil
}

func (b *boltStore) Close() error {
	b.db.Close()
	return nil
}
