package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
)

const (
	DatabaseName = "my.db"
)

var (
	db *bolt.DB
)

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
	return nil
}

func createInitBucketFunc(bucketName string) func(tx *bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %s", bucketName, err)
		}
		return nil
	}
}
