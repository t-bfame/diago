// Package storage implements a storage API used to store Diago objects.
// It uses boltDB as the underlying storage management.
package storage

import (
	"fmt"

	"github.com/boltdb/bolt"
)

var (
	db *bolt.DB
)

// Initializes storage file
func InitDatabase(dbName string) error {
	value, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return err
	}

	db = value

	if err := initStorageJob(db); err != nil {
		return err
	}
	if err := initStorageTest(db); err != nil {
		return err
	}
	if err := initStorageTestInstance(db); err != nil {
		return err
	}
	if err := initStorageTestSchedule(db); err != nil {
		return err
	}
	return nil
}

// Internal helper function used to generate a function creating a bucket with given name
func createInitBucketFunc(bucketName string) func(tx *bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %s", bucketName, err)
		}
		return nil
	}
}
