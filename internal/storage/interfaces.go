package storage

import (
	mgr "github.com/t-bfame/diago/internal/manager"
)

type JobDao interface {
	AddJob(job *mgr.Job) error
	GetJobByJobId(jobId mgr.JobID) (*mgr.Job, error)
}

type TestDao interface {
	AddTest(test *mgr.Test) error
	GetTestByTestId(testId mgr.TestID) (*mgr.Test, error)
}

type TestInstanceDao interface {
	AddTestInstance(testInstance *mgr.TestInstance) error
	GetTestInstanceByTestInstanceId(testInstanceId mgr.TestInstanceID) (*mgr.TestInstance, error)
}
