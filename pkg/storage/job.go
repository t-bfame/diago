package storage

import (
	"fmt"

	"github.com/t-bfame/diago/pkg/model"
	"github.com/t-bfame/diago/pkg/tools"

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

func DeleteJob(jobID model.JobID) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(JobBucketName))
		if b == nil {
			return fmt.Errorf("missing bucket '%s'", JobBucketName)
		}
		if err := b.Delete([]byte(jobID)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("jobID", jobID).Error("Failed to delete Job")
		return err
	}
	return nil
}

func GetJobByJobId(jobId model.JobID) (*model.Job, error) {
	var result *model.Job
	if err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(JobBucketName))
		data := b.Get([]byte(jobId))
		if data == nil {
			return nil
		}
		if err = tools.GobDecode(&result, data); err != nil {
			return fmt.Errorf("failed to decode Job due to: %s", err)
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("jobId", jobId).Error("Failed to GetJobById")
	}
	return result, nil
}

func GetAllJobs() ([]*model.Job, error) {
	var jobs = make([]*model.Job, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(JobBucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var job *model.Job
			if err := tools.GobDecode(&job, v); err != nil {
				return fmt.Errorf("failed to decode Job due to: %s", err)
			}
			jobs = append(jobs, job)
		}

		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetAllJobs")
		return nil, err
	}
	return jobs, nil
}
