// Package storage implements a storage API used to store Diago objects.
// It uses boltDB as the underlying storage management.
package storage

import (
	"context"
	"fmt"

	"github.com/boltdb/bolt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	db          *bolt.DB
	mongoClient *mongo.Client
)

func ConnectToMongoDB(ctx context.Context, dbString string) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbString))
	if err != nil {
		return err
	}
	mongoClient = client
	return nil
}

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
	if err := initStorageUser(db); err != nil {
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
