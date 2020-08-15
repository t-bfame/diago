package scheduler

import (
	mgr "github.com/t-bfame/diago/internal/manager"
)

// Event passed between scheduler and manager
type Event interface {
	getJobID() mgr.JobID
}
