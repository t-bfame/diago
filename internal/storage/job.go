package storage

import (
	"fmt"

	"github.com/t-bfame/diago/internal/model"
	"github.com/t-bfame/diago/internal/tools"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

const JobBucketName = "Job"

func initStorageJob(db *bolt.DB) error {
	if err := db.Update(createInitBucketFunc(JobBucketName)); err != nil {
		return err
	}
	return nil
}

func AddJob(job *model.Job) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(JobBucketName))
		if b == nil {
			return fmt.Errorf("missing bucket '%s'", JobBucketName)
		}
		enc, err := tools.GobEncode(job)
		if err != nil {
			return fmt.Errorf("failed to encode Job due to: %s", err)
		}
		if err := b.Put([]byte(job.ID), enc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("job", job).Error("Failed to add Job")
		return err
	}
	return nil
}

func GetJobByJobId(jobId model.JobID) (*model.Job, error) {
	var result *model.Job
	if err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(JobBucketName))
		if err = tools.GobDecode(&result, b.Get([]byte(jobId))); err != nil {
			return fmt.Errorf("failed to decode Job due to: %s", err)
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("jobId", jobId).Error("Failed to GetJobById")
	}
	return result, nil
}
