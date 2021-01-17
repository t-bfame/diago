package storage

import (
	"encoding/gob"
	"fmt"

	"github.com/t-bfame/diago/internal/metrics"
	"github.com/t-bfame/diago/internal/model"
	"github.com/t-bfame/diago/internal/tools"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

const (
	TestInstanceBucketName             = "TestInstance"
	IdxTestID2TestInstanceIDBucketName = "TestInstance"
)

type IdxTestID2TestInstanceID struct {
	TestId          model.TestID
	TestInstanceIds map[model.TestInstanceID]bool
}

func initStorageTestInstance(db *bolt.DB) error {
	if err := db.Update(createInitBucketFunc(TestInstanceBucketName)); err != nil {
		return err
	}
	if err := db.Update(createInitBucketFunc(IdxTestID2TestInstanceIDBucketName)); err != nil {
		return err
	}
	gob.Register(map[string]*metrics.Metrics{})
	return nil
}

func AddTestInstance(testInstance *model.TestInstance) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		if err := doAddTestInstance(tx, testInstance); err != nil {
			return err
		}
		if err := doAddIndex(tx, testInstance); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testInstance", testInstance).Error("Failed to add TestInstance")
		return err
	}
	return nil
}

func GetTestInstance(testInstanceID model.TestInstanceID) (*model.TestInstance, error) {
	var result *model.TestInstance
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestInstanceBucketName))
		if err := tools.GobDecode(&result, b.Get([]byte(testInstanceID))); err != nil {
			return fmt.Errorf("failed to decode TestInstance due to: %s", err)
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testInstanceID", testInstanceID).Error("Failed to GetTestInstance")
		return nil, err
	}
	return result, nil
}

func GetTestInstances(testInstanceIDs []model.TestInstanceID) ([]*model.TestInstance, error) {
	var result = make([]*model.TestInstance, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		if instances, err := doGetTestInstances(tx, testInstanceIDs); err != nil {
			return err
		} else {
			result = instances
		}
		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetTestInstances")
		return nil, err
	}
	return result, nil
}

func GetTestInstancesByTestID(testID model.TestID) ([]*model.TestInstance, error) {
	var result = make([]*model.TestInstance, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		var index *IdxTestID2TestInstanceID
		if value, err := doGetIndex(tx, testID); err != nil {
			return err
		} else {
			index = value
		}

		if instances, err := doGetTestInstancesByIDMap(tx, index.TestInstanceIds); err != nil {
			return err
		} else {
			result = instances
		}
		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetTestInstancesByTestID")
		return nil, err
	}
	return result, nil
}

func doAddTestInstance(tx *bolt.Tx, testInstance *model.TestInstance) error {
	b := tx.Bucket([]byte(TestInstanceBucketName))
	if b == nil {
		return fmt.Errorf("missing bucket '%s'", TestInstanceBucketName)
	}
	enc, err := tools.GobEncode(testInstance)
	if err != nil {
		return fmt.Errorf("failed to encode TestInstance due to: %s", err)
	}
	if err := b.Put([]byte(testInstance.ID), enc); err != nil {
		return err
	}
	return nil
}

func doGetIndex(tx *bolt.Tx, testID model.TestID) (*IdxTestID2TestInstanceID, error) {
	b := tx.Bucket([]byte(IdxTestID2TestInstanceIDBucketName))
	if b == nil {
		return nil, fmt.Errorf("missing bucket '%s'", IdxTestID2TestInstanceIDBucketName)
	}

	data := b.Get([]byte(testID))
	if data == nil {
		return nil, nil
	}

	var index *IdxTestID2TestInstanceID
	if err := tools.GobDecode(&index, data); err != nil {
		return nil, fmt.Errorf("failed to decode IdxTestID2TestInstanceID due to: %s", err)
	}
	return index, nil
}

func doAddIndex(tx *bolt.Tx, testInstance *model.TestInstance) error {
	testID := testInstance.TestID
	instanceID := testInstance.ID

	b := tx.Bucket([]byte(IdxTestID2TestInstanceIDBucketName))
	if b == nil {
		return fmt.Errorf("missing bucket '%s'", IdxTestID2TestInstanceIDBucketName)
	}

	var index *IdxTestID2TestInstanceID
	if value, err := doGetIndex(tx, testID); err != nil {
		return err
	} else {
		index = value
		if index == nil {
			index = &IdxTestID2TestInstanceID{
				TestId:          testID,
				TestInstanceIds: make(map[model.TestInstanceID]bool),
			}
		}
	}

	index.TestInstanceIds[instanceID] = true

	enc, err := tools.GobEncode(index)
	if err != nil {
		return fmt.Errorf("failed to encode IdxTestID2TestInstanceID due to: %s", err)
	}
	if err := b.Put([]byte(testID), enc); err != nil {
		return err
	}

	return nil
}

func doGetTestInstances(tx *bolt.Tx, testInstanceIDs []model.TestInstanceID) ([]*model.TestInstance, error) {
	instances := make([]*model.TestInstance, 0)
	b := tx.Bucket([]byte(TestInstanceBucketName))
	for _, id := range testInstanceIDs {
		var value *model.TestInstance
		if err := tools.GobDecode(&value, b.Get([]byte(id))); err != nil {
			return nil, fmt.Errorf("failed to decode TestInstance due to: %s", err)
		}
		instances = append(instances, value)
	}
	return instances, nil
}

func doGetTestInstancesByIDMap(tx *bolt.Tx, testInstanceIDs map[model.TestInstanceID]bool) ([]*model.TestInstance, error) {
	instances := make([]*model.TestInstance, 0)
	b := tx.Bucket([]byte(TestInstanceBucketName))
	for id := range testInstanceIDs {
		var value *model.TestInstance
		if err := tools.GobDecode(&value, b.Get([]byte(id))); err != nil {
			return nil, fmt.Errorf("failed to decode TestInstance due to: %s", err)
		}
		instances = append(instances, value)
	}
	return instances, nil
}
