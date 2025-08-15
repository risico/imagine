package imagine

import (
	"database/sql"
	"fmt"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
	"github.com/juju/errors"
)

type sqliteStore struct {
	db        *sql.DB
	tableName string
}

type SQLiteStoreParams struct {
	Path      string
	TableName string
}

func NewSQLiteStorage(params SQLiteStoreParams) (Store, error) {
	if params.Path == "" {
		return nil, errors.New("SQLite path is required")
	}
	
	if params.TableName == "" {
		params.TableName = "imagine_images"
	}
	
	db, err := sql.Open("sqlite3", params.Path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	
	store := &sqliteStore{
		db:        db,
		tableName: params.TableName,
	}
	
	// Create table if it doesn't exist
	if err := store.createTable(); err != nil {
		db.Close()
		return nil, errors.Trace(err)
	}
	
	fmt.Printf("[SQLiteStore] Initialized with path: %s, table: %s\n", params.Path, params.TableName)
	
	return store, nil
}

func (s *sqliteStore) createTable() error {
	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			hash TEXT PRIMARY KEY,
			data BLOB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, s.tableName)
	
	_, err := s.db.Exec(createSQL)
	if err != nil {
		return errors.Trace(err)
	}
	
	// Create index on accessed_at for TTL cleanup if needed
	indexSQL := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_accessed 
		ON %s(accessed_at)
	`, s.tableName, s.tableName)
	
	_, err = s.db.Exec(indexSQL)
	return errors.Trace(err)
}

func (s *sqliteStore) Set(key string, data []byte) error {
	fmt.Printf("[SQLiteStore] Set called for key: %s, data size: %d bytes\n", key, len(data))
	
	// Use REPLACE to handle both insert and update
	query := fmt.Sprintf(`
		REPLACE INTO %s (hash, data, created_at, accessed_at) 
		VALUES (?, ?, ?, ?)
	`, s.tableName)
	
	now := time.Now()
	_, err := s.db.Exec(query, key, data, now, now)
	
	if err != nil {
		fmt.Printf("[SQLiteStore] Error setting key %s: %v\n", key, err)
		return errors.Trace(err)
	}
	
	fmt.Printf("[SQLiteStore] Successfully stored key: %s\n", key)
	return nil
}

func (s *sqliteStore) Get(key string) (data []byte, found bool, err error) {
	fmt.Printf("[SQLiteStore] Get called for key: %s\n", key)
	
	query := fmt.Sprintf(`
		SELECT data FROM %s WHERE hash = ?
	`, s.tableName)
	
	err = s.db.QueryRow(query, key).Scan(&data)
	
	if err == sql.ErrNoRows {
		fmt.Printf("[SQLiteStore] Key not found: %s\n", key)
		return nil, false, ErrKeyNotFound
	}
	
	if err != nil {
		fmt.Printf("[SQLiteStore] Error getting key %s: %v\n", key, err)
		return nil, false, errors.Trace(err)
	}
	
	// Update accessed_at time
	updateQuery := fmt.Sprintf(`
		UPDATE %s SET accessed_at = ? WHERE hash = ?
	`, s.tableName)
	s.db.Exec(updateQuery, time.Now(), key)
	
	fmt.Printf("[SQLiteStore] Found key %s, data size: %d bytes\n", key, len(data))
	return data, true, nil
}

func (s *sqliteStore) Delete(key string) error {
	fmt.Printf("[SQLiteStore] Delete called for key: %s\n", key)
	
	query := fmt.Sprintf(`
		DELETE FROM %s WHERE hash = ?
	`, s.tableName)
	
	_, err := s.db.Exec(query, key)
	
	if err != nil {
		fmt.Printf("[SQLiteStore] Error deleting key %s: %v\n", key, err)
		return errors.Trace(err)
	}
	
	fmt.Printf("[SQLiteStore] Successfully deleted key: %s\n", key)
	return nil
}

func (s *sqliteStore) Close() error {
	fmt.Printf("[SQLiteStore] Closing database connection\n")
	return s.db.Close()
}

// Cleanup removes entries older than the TTL
func (s *sqliteStore) Cleanup(ttl time.Duration) error {
	if ttl <= 0 {
		return nil // No cleanup if TTL is not set
	}
	
	cutoff := time.Now().Add(-ttl)
	query := fmt.Sprintf(`
		DELETE FROM %s WHERE accessed_at < ?
	`, s.tableName)
	
	result, err := s.db.Exec(query, cutoff)
	if err != nil {
		return errors.Trace(err)
	}
	
	deleted, _ := result.RowsAffected()
	if deleted > 0 {
		fmt.Printf("[SQLiteStore] Cleaned up %d old entries\n", deleted)
	}
	
	return nil
}