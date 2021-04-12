package storage

import (
	"encoding/gob"
	"fmt"

	"github.com/t-bfame/diago/pkg/metrics"
	"github.com/t-bfame/diago/pkg/model"
	"github.com/t-bfame/diago/pkg/tools"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

const (
	// This is the boltDB bucket name for storing "model/TestInstance".
	TestInstanceBucketName = "TestInstance"
	// This is the boltDB bucket name for storing "IdxTestID2TestInstanceID".
	IdxTestID2TestInstanceIDBucketName = "TestInstanceIdx"
)

// IdxTestID2TestInstanceID stores mapping from one TestID to multiple TestInstanceID.
type IdxTestID2TestInstanceID struct {
	TestId          model.TestID
	TestInstanceIds map[model.TestInstanceID]bool
}

// Initialize boltDB for "model/TestInstance" storage.
func initStorageTestInstance(db *bolt.DB) error {
	if err := db.Update(createInitBucketFunc(TestInstanceBucketName)); err != nil {
		return err
	}
	if err := db.Update(createInitBucketFunc(IdxTestID2TestInstanceIDBucketName)); err != nil {
		return err
	}
	gob.Register(map[string]*metrics.Metrics{})
	gob.Register(map[model.ChaosID]model.ChaosResult{})
	return nil
}

// Add a "model/TestInstance" to the storage.
func AddTestInstance(testInstance *model.TestInstance) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		if err := doAddTestInstance(tx, testInstance); err != nil {
			return err
		}
		if err := doAddTestInstanceIndex(tx, testInstance); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testInstance", testInstance).Error("Failed to add TestInstance")
		return err
	}
	return nil
}

// Delete a "model/TestInstance" with the specified TestInstanceID from the storage.
func DeleteTestInstance(testInstanceID model.TestInstanceID) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		var testID model.TestID
		if instance, err := doGetTestInstance(tx, testInstanceID); err != nil {
			return err
		} else {
			testID = instance.TestID
		}

		if err := doDeleteTestInstance(tx, testInstanceID); err != nil {
			return err
		}

		if err := doRemoveTestInstanceIndex(tx, testID, testInstanceID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testInstanceID", testInstanceID).Error("Failed to delete TestInstance")
		return err
	}
	return nil
}

// Retrieve a "model/TestInstance" with the specified TestInstanceID from the storage.
func GetTestInstance(testInstanceID model.TestInstanceID) (*model.TestInstance, error) {
	var result *model.TestInstance
	if err := db.View(func(tx *bolt.Tx) error {
		if instance, err := doGetTestInstance(tx, testInstanceID); err != nil {
			return err
		} else {
			result = instance
			return nil
		}
	}); err != nil {
		log.WithError(err).WithField("testInstanceID", testInstanceID).Error("Failed to GetTestInstance")
		return nil, err
	}
	return result, nil
}

// Retrieve all "model/TestInstance" stored in the storage.
func GetAllTestInstances() ([]*model.TestInstance, error) {
	var instances = make([]*model.TestInstance, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestInstanceBucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var instance *model.TestInstance
			if err := tools.GobDecode(&instance, v); err != nil {
				return fmt.Errorf("failed to decode TestInstance due to: %s", err)
			}
			instances = append(instances, instance)
		}

		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetAllTestInstances")
		return nil, err
	}
	return instances, nil
}

// Retrieve an array of "model/TestInstance" with the specified array of TestInstanceID from the storage.
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

// Retrieve all "model/TestInstance" with specified TestID from the storage.
func GetTestInstancesByTestID(testID model.TestID) ([]*model.TestInstance, error) {
	var result = make([]*model.TestInstance, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		var index *IdxTestID2TestInstanceID
		if value, err := doGetTestInstanceIndex(tx, testID); err != nil {
			return err
		} else {
			index = value
		}

		if index == nil {
			return nil
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

// Internal function used to retrieve a "model/TestInstance" with the specified TestInstanceID using a provided boltDB transaction.
func doGetTestInstance(tx *bolt.Tx, testInstanceID model.TestInstanceID) (*model.TestInstance, error) {
	var result *model.TestInstance
	b := tx.Bucket([]byte(TestInstanceBucketName))
	data := b.Get([]byte(testInstanceID))
	if data == nil {
		return nil, nil
	}
	if err := tools.GobDecode(&result, data); err != nil {
		return nil, fmt.Errorf("failed to decode TestInstance due to: %s", err)
	}
	return result, nil
}

// Internal function used to add a "model/TestInstance" using the provided boltDB transaction.
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

// Internal function used to delete a "model/TestInstance" with the specified TestInstanceID using the provided boltDB transaction.
func doDeleteTestInstance(tx *bolt.Tx, testInstanceID model.TestInstanceID) error {
	b := tx.Bucket([]byte(TestInstanceBucketName))
	if b == nil {
		return fmt.Errorf("missing bucket '%s'", TestInstanceBucketName)
	}
	if err := b.Delete([]byte(testInstanceID)); err != nil {
		return err
	}
	return nil
}

// Internal function used to retrieve a IdxTestID2TestInstanceID with the specified TestID from the storage.
func doGetTestInstanceIndex(tx *bolt.Tx, testID model.TestID) (*IdxTestID2TestInstanceID, error) {
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

// Internal function used to add a IdxTestID2TestInstanceID to the storage.
func doAddTestInstanceIndex(tx *bolt.Tx, testInstance *model.TestInstance) error {
	testID := testInstance.TestID
	instanceID := testInstance.ID

	b := tx.Bucket([]byte(IdxTestID2TestInstanceIDBucketName))
	if b == nil {
		return fmt.Errorf("missing bucket '%s'", IdxTestID2TestInstanceIDBucketName)
	}

	var index *IdxTestID2TestInstanceID
	if value, err := doGetTestInstanceIndex(tx, testID); err != nil {
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

// Internal function used to remove a index mapping from the specified TestID to the specified TestInstanceID from the storage.
func doRemoveTestInstanceIndex(tx *bolt.Tx, testID model.TestID, instanceID model.TestInstanceID) error {
	b := tx.Bucket([]byte(IdxTestID2TestInstanceIDBucketName))
	if b == nil {
		return fmt.Errorf("missing bucket '%s'", IdxTestID2TestInstanceIDBucketName)
	}

	var index *IdxTestID2TestInstanceID
	if value, err := doGetTestInstanceIndex(tx, testID); err != nil {
		return err
	} else {
		index = value
		if index == nil {
			return nil
		}
	}

	delete(index.TestInstanceIds, instanceID)

	enc, err := tools.GobEncode(index)
	if err != nil {
		return fmt.Errorf("failed to encode IdxTestID2TestInstanceID due to: %s", err)
	}
	if err := b.Put([]byte(testID), enc); err != nil {
		return err
	}

	return nil
}

// Internal function used to get all "model/TestInstance" with specified array of TestInstanceID stored in the storage using the provided boltDB transaction.
func doGetTestInstances(tx *bolt.Tx, testInstanceIDs []model.TestInstanceID) ([]*model.TestInstance, error) {
	instances := make([]*model.TestInstance, 0)
	b := tx.Bucket([]byte(TestInstanceBucketName))
	for _, id := range testInstanceIDs {
		var value *model.TestInstance
		data := b.Get([]byte(id))
		if data == nil {
			continue
		}
		if err := tools.GobDecode(&value, data); err != nil {
			return nil, fmt.Errorf("failed to decode TestInstance due to: %s", err)
		}
		instances = append(instances, value)
	}
	return instances, nil
}

// Internal function used to retrieve a list of "model/TestInstance" with TestInstanceID in the specified map.
func doGetTestInstancesByIDMap(tx *bolt.Tx, testInstanceIDs map[model.TestInstanceID]bool) ([]*model.TestInstance, error) {
	instances := make([]*model.TestInstance, 0)
	b := tx.Bucket([]byte(TestInstanceBucketName))
	for id := range testInstanceIDs {
		var value *model.TestInstance
		data := b.Get([]byte(id))
		if data == nil {
			continue
		}
		if err := tools.GobDecode(&value, data); err != nil {
			return nil, fmt.Errorf("failed to decode TestInstance due to: %s", err)
		}
		instances = append(instances, value)
	}
	return instances, nil
}
