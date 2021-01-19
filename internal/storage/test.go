package storage

import (
	"fmt"

	"github.com/t-bfame/diago/internal/model"
	"github.com/t-bfame/diago/internal/tools"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

const TestBucketName = "Test"

func initStorageTest(db *bolt.DB) error {
	if err := db.Update(createInitBucketFunc(TestBucketName)); err != nil {
		return err
	}
	return nil
}

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
