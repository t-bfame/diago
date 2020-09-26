package daoImplBoltDB

import (
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
)

const TestBucketName = "Test"

type TestDao struct {
	db *bolt.DB
}

func NewTestDao(db *bolt.DB) *TestDao {
	testDao := TestDao{db: db}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(TestBucketName))
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %s", TestBucketName, err)
		}
		return nil
	}); err != nil {
		return nil
	}
	return &testDao
}

func (testDao TestDao) AddTest(test *mgr.Test) error {
	if err := testDao.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestBucketName))
		if b == nil {
			return fmt.Errorf("failed to add test due to missing bucket: %s", TestBucketName)
		}
		enc, err := gobEncode(test)
		if err != nil {
			return fmt.Errorf("could not encode Test %s: %s", test.ID, err)
		}
		if err := b.Put([]byte(test.ID), enc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (testDao TestDao) GetTestByTestId(testId mgr.TestID) (*mgr.Test, error) {
	var result *mgr.Test
	if err := testDao.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(TestBucketName))
		if err = gobDecode(&result, b.Get([]byte(testId))); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return result, nil
}
