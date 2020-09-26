package daoImplBoltDB

import (
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
)

const JobBucketName = "Job"

type JobDao struct {
	db *bolt.DB
}

func NewJobDao(db *bolt.DB) *JobDao {
	jobDao := JobDao{db: db}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(JobBucketName))
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %s", JobBucketName, err)
		}
		return nil
	}); err != nil {
		return nil
	}
	return &jobDao
}

func (jobDao JobDao) AddJob(job *mgr.Job) error {
	if err := jobDao.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(JobBucketName))
		if b == nil {
			return fmt.Errorf("failed to add job due to missing bucket: %s", JobBucketName)
		}
		enc, err := gobEncode(job)
		if err != nil {
			return fmt.Errorf("could not encode Job %s: %s", job.ID, err)
		}
		if err := b.Put([]byte(job.ID), enc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (jobDao JobDao) GetJobByJobId(jobId mgr.JobID) (*mgr.Job, error) {
	var result *mgr.Job
	if err := jobDao.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(JobBucketName))
		if err = gobDecode(&result, b.Get([]byte(jobId))); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return result, nil
}
