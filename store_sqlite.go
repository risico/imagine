package imagine

type sqliteStore struct{}

func (s *sqliteStore) Set(key string, data []byte) error {
    return nil
}

func (s *sqliteStore) Get(key string) (data []byte, ok bool, err error) {
    return nil, false, nil
}

func (s *sqliteStore) Delete(key string) error {
    return nil
}

func (s *sqliteStore) Close() error {
    return nil
}

type SqliteStoreParams struct {}

func NewSqliteStore() Store {
    return &sqliteStore{}
}

// createTable creates the images table if it doesn't exist
func (s *sqliteStore) createTable() error {
    return nil
}

