package storage

import (
	"bytes"
	"fmt"

	"github.com/t-bfame/diago/pkg/model"
	"github.com/t-bfame/diago/pkg/tools"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

// This is the boltDB bucket name for storing "model/Test".
const TestBucketName = "Test"

// Initializes boltDB for "model/Test" storage.
func initStorageTest(db *bolt.DB) error {
	if err := db.Update(createInitBucketFunc(TestBucketName)); err != nil {
		return err
	}
	return nil
}

// Add a "model/Test" to the storage.
func AddTest(test *model.Test) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestBucketName))
		if b == nil {
			return fmt.Errorf("missing bucket '%s'", TestBucketName)
		}
		enc, err := tools.GobEncode(test)
		if err != nil {
			return fmt.Errorf("failed to encode Test due to: %s", err)
		}
		if err := b.Put([]byte(test.ID), enc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("test", test).Error("Failed to add Test")
		return err
	}
	return nil
}

// Delete a "model/Test" with the specified TestID from the storage.
func DeleteTest(testID model.TestID) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestBucketName))
		if b == nil {
			return fmt.Errorf("missing bucket '%s'", TestBucketName)
		}
		if err := b.Delete([]byte(testID)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testID", testID).Error("Failed to delete Test")
		return err
	}
	return nil
}

// Retrieve a "model/Test" with the specified TestID from the storage.
func GetTestByTestId(testId model.TestID) (*model.Test, error) {
	var result *model.Test
	if err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(TestBucketName))
		data := b.Get([]byte(testId))
		if data == nil {
			return nil
		}
		if err = tools.GobDecode(&result, data); err != nil {
			return fmt.Errorf("failed to decode Test due to: %s", err)
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("TestId", testId).Error("Failed to GetTestByTestId")
		return nil, err
	}
	return result, nil
}

// Retrieve all "model/Test" stored in the storage.
func GetAllTests() ([]*model.Test, error) {
	var tests = make([]*model.Test, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestBucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var test *model.Test
			if err := tools.GobDecode(&test, v); err != nil {
				return fmt.Errorf("failed to decode Test due to: %s", err)
			}
			tests = append(tests, test)
		}

		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetAllTests")
		return nil, err
	}
	return tests, nil
}

// Retrieve all "model/Test" with the specified JobID prefix from the storage.
func GetAllTestsWithPrefix(prefixStr string) ([]*model.Test, error) {
	var tests = make([]*model.Test, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestBucketName))
		c := b.Cursor()

		prefix := []byte(prefixStr)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var test *model.Test
			if err := tools.GobDecode(&test, v); err != nil {
				return fmt.Errorf("failed to decode Test due to: %s", err)
			}
			tests = append(tests, test)
		}

		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetAllTestsWithPrefix")
		return nil, err
	}
	return tests, nil
}
