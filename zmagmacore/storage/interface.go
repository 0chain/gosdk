// DEPRECATED: This package is deprecated and will be removed in a future release.
package storage

import (
	"sync"

	"github.com/dgraph-io/badger/v3"

	"github.com/0chain/gosdk/zmagmacore/errors"
)

type (
	// Storage represents the main storage based on badger.DB.
	Storage struct {
		db *badger.DB

		singleton sync.Once // used for opening connection only once
	}

	// Value represent value that can be stored as encoded bytes.
	Value interface {
		Encode() []byte
	}
)

var (
	// storageInst represents singleton storage.
	storageInst = &Storage{}
)

// Open opens singleton connection to storage.
//
// If an error occurs during execution, the program terminates with code 2 and the error will be written in os.Stderr.
//
// Should be used only once while application is starting.
func Open(path string) {
	storageInst.singleton.Do(func() {
		opts := badger.DefaultOptions(path)
		opts.Logger = nil

		db, err := badger.Open(opts)
		if err != nil {
			errors.ExitErr("error while opening storage: %v", err, 2)
		}

		storageInst.db = db
	})
}

// GetStorage returns current storage implementation.
func GetStorage() *Storage {
	return storageInst
}

// Del deletes entry by the key.
func (s *Storage) Del(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Get retrieves entry by the key.
func (s *Storage) Get(key []byte) (value []byte, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		var item *badger.Item
		item, err = txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			value = val
			return nil
		})
	})

	return
}

// Set sets encoded Value with provided key.
func (s *Storage) Set(key []byte, value Value) error {
	return s.db.Update(func(txn *badger.Txn) error {
		blob := value.Encode()
		return txn.Set(key, blob)
	})
}

// SetWithRetries sets encoded Value with provided key with retries.
func (s *Storage) SetWithRetries(key []byte, value Value, numRetries int) error {
	var err error
	for i := 0; i < numRetries; i++ {
		if err = s.Set(key, value); err != nil {
			continue
		}
		break
	}

	return err
}

// Iterate iterates all elements with provided prefix and processes it with the handler.
func (s *Storage) Iterate(handler func(item *badger.Item) error, prefix []byte) error {
	return s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := handler(it.Item()); err != nil {
				return err
			}
		}
		return nil
	})
}

// NewTransaction creates new badger.Txn.
//
// For read-only transactions set update flag to false.
func (s *Storage) NewTransaction(update bool) *badger.Txn {
	return s.db.NewTransaction(update)
}

// Close closes a DB.
func (s *Storage) Close() error {
	return s.db.Close()
}
