package storage

import (
	mgr "github.com/t-bfame/diago/internal/manager"
)

func AddJob(job *mgr.Job) error {
	return daoFactory.GetJobDao().AddJob(job)
}

func GetJobByJobId(jobId mgr.JobID) (*mgr.Job, error) {
	return daoFactory.GetJobDao().GetJobByJobId(jobId)
}
