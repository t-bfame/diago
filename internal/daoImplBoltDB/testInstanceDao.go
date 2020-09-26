package daoImplBoltDB

import (
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
)

const TestInstanceBucketName = "TestInstance"

type TestInstanceDao struct {
	db *bolt.DB
}

func NewTestInstanceDao(db *bolt.DB) *TestInstanceDao {
	testInstanceDao := TestInstanceDao{db: db}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(TestInstanceBucketName))
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %s", TestInstanceBucketName, err)
		}
		return nil
	}); err != nil {
		return nil
	}
	return &testInstanceDao
}

func (testInstanceDao TestInstanceDao) AddTestInstance(testInstance *mgr.TestInstance) error {
	if err := testInstanceDao.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestInstanceBucketName))
		if b == nil {
			return fmt.Errorf("failed to add testInstance due to missing bucket: %s", TestInstanceBucketName)
		}
		enc, err := gobEncode(testInstance)
		if err != nil {
			return fmt.Errorf("could not encode TestInstance %s: %s", testInstance.ID, err)
		}
		if err := b.Put([]byte(testInstance.ID), enc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (testInstanceDao TestInstanceDao) GetTestInstanceByTestInstanceId(testInstanceID mgr.TestInstanceID) (*mgr.TestInstance, error) {
	var result *mgr.TestInstance
	if err := testInstanceDao.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(TestInstanceBucketName))
		if err = gobDecode(&result, b.Get([]byte(testInstanceID))); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return result, nil
}
