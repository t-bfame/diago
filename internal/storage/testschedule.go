package storage

import (
	"fmt"

	"github.com/t-bfame/diago/internal/model"
	"github.com/t-bfame/diago/internal/tools"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

const (
	TestScheduleBucketName             = "TestSchedule"
	IdxTestID2TestScheduleIDBucketName = "TestScheduleIdx"
)

type IdxTestID2TestScheduleID struct {
	TestId          model.TestID
	TestScheduleIds map[model.TestScheduleID]bool
}

func initStorageTestSchedule(db *bolt.DB) error {
	if err := db.Update(createInitBucketFunc(TestScheduleBucketName)); err != nil {
		return err
	}
	if err := db.Update(createInitBucketFunc(IdxTestID2TestScheduleIDBucketName)); err != nil {
		return err
	}
	return nil
}

func AddTestSchedule(testSchedule *model.TestSchedule) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		if err := doAddTestSchedule(tx, testSchedule); err != nil {
			return err
		}
		if err := doAddTestScheduleIndex(tx, testSchedule); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testSchedule", testSchedule).Error("Failed to add TestSchedule")
		return err
	}
	return nil
}

func DeleteTestSchedule(testScheduleID model.TestScheduleID) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestScheduleBucketName))
		if b == nil {
			return fmt.Errorf("missing bucket '%s'", TestScheduleBucketName)
		}
		if err := b.Delete([]byte(testScheduleID)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testScheduleID", testScheduleID).Error("Failed to delete TestSchedule")
		return err
	}
	return nil
}

func GetTestSchedule(testScheduleID model.TestScheduleID) (*model.TestSchedule, error) {
	var result *model.TestSchedule
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestScheduleBucketName))
		data := b.Get([]byte(testScheduleID))
		if data == nil {
			return nil
		}
		if err := tools.GobDecode(&result, data); err != nil {
			return fmt.Errorf("failed to decode TestSchedule due to: %s", err)
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("testScheduleID", testScheduleID).Error("Failed to GetTestSchedule")
		return nil, err
	}
	return result, nil
}

func GetAllTestSchedules() ([]*model.TestSchedule, error) {
	var schedules = make([]*model.TestSchedule, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TestScheduleBucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var schedule *model.TestSchedule
			if err := tools.GobDecode(&schedule, v); err != nil {
				return fmt.Errorf("failed to decode TestSchedule due to: %s", err)
			}
			schedules = append(schedules, schedule)
		}

		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetAllTestSchedules")
		return nil, err
	}
	return schedules, nil
}

func GetTestSchedules(testScheduleIDs []model.TestScheduleID) ([]*model.TestSchedule, error) {
	var result = make([]*model.TestSchedule, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		if instances, err := doGetTestSchedules(tx, testScheduleIDs); err != nil {
			return err
		} else {
			result = instances
		}
		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetTestSchedules")
		return nil, err
	}
	return result, nil
}

func GetTestSchedulesByTestID(testID model.TestID) ([]*model.TestSchedule, error) {
	var result = make([]*model.TestSchedule, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		var index *IdxTestID2TestScheduleID
		if value, err := doGetTestScheduleIndex(tx, testID); err != nil {
			return err
		} else {
			index = value
		}

		if instances, err := doGetTestSchedulesByIDMap(tx, index.TestScheduleIds); err != nil {
			return err
		} else {
			result = instances
		}
		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetTestSchedulesByTestID")
		return nil, err
	}
	return result, nil
}

func doAddTestSchedule(tx *bolt.Tx, testSchedule *model.TestSchedule) error {
	b := tx.Bucket([]byte(TestScheduleBucketName))
	if b == nil {
		return fmt.Errorf("missing bucket '%s'", TestScheduleBucketName)
	}
	enc, err := tools.GobEncode(testSchedule)
	if err != nil {
		return fmt.Errorf("failed to encode TestSchedule due to: %s", err)
	}
	if err := b.Put([]byte(testSchedule.ID), enc); err != nil {
		return err
	}
	return nil
}

func doGetTestScheduleIndex(tx *bolt.Tx, testID model.TestID) (*IdxTestID2TestScheduleID, error) {
	b := tx.Bucket([]byte(IdxTestID2TestScheduleIDBucketName))
	if b == nil {
		return nil, fmt.Errorf("missing bucket '%s'", IdxTestID2TestScheduleIDBucketName)
	}

	data := b.Get([]byte(testID))
	if data == nil {
		return nil, nil
	}

	var index *IdxTestID2TestScheduleID
	if err := tools.GobDecode(&index, data); err != nil {
		return nil, fmt.Errorf("failed to decode IdxTestID2TestScheduleID due to: %s", err)
	}
	return index, nil
}

func doAddTestScheduleIndex(tx *bolt.Tx, testSchedule *model.TestSchedule) error {
	testID := testSchedule.TestID
	scheduleID := testSchedule.ID

	b := tx.Bucket([]byte(IdxTestID2TestScheduleIDBucketName))
	if b == nil {
		return fmt.Errorf("missing bucket '%s'", IdxTestID2TestScheduleIDBucketName)
	}

	var index *IdxTestID2TestScheduleID
	if value, err := doGetTestScheduleIndex(tx, testID); err != nil {
		return err
	} else {
		index = value
		if index == nil {
			index = &IdxTestID2TestScheduleID{
				TestId:          testID,
				TestScheduleIds: make(map[model.TestScheduleID]bool),
			}
		}
	}

	index.TestScheduleIds[scheduleID] = true

	enc, err := tools.GobEncode(index)
	if err != nil {
		return fmt.Errorf("failed to encode IdxTestID2TestScheduleID due to: %s", err)
	}
	if err := b.Put([]byte(testID), enc); err != nil {
		return err
	}

	return nil
}

func doGetTestSchedules(tx *bolt.Tx, testScheduleIDs []model.TestScheduleID) ([]*model.TestSchedule, error) {
	schedules := make([]*model.TestSchedule, 0)
	b := tx.Bucket([]byte(TestScheduleBucketName))
	for _, id := range testScheduleIDs {
		var value *model.TestSchedule
		data := b.Get([]byte(id))
		if data == nil {
			continue
		}
		if err := tools.GobDecode(&value, data); err != nil {
			return nil, fmt.Errorf("failed to decode TestSchedule due to: %s", err)
		}
		schedules = append(schedules, value)
	}
	return schedules, nil
}

func doGetTestSchedulesByIDMap(tx *bolt.Tx, testScheduleIDs map[model.TestScheduleID]bool) ([]*model.TestSchedule, error) {
	schedules := make([]*model.TestSchedule, 0)
	b := tx.Bucket([]byte(TestScheduleBucketName))
	for id := range testScheduleIDs {
		var value *model.TestSchedule
		data := b.Get([]byte(id))
		if data == nil {
			continue
		}
		if err := tools.GobDecode(&value, data); err != nil {
			return nil, fmt.Errorf("failed to decode TestSchedule due to: %s", err)
		}
		schedules = append(schedules, value)
	}
	return schedules, nil
}
