package scheduler

import (
	mgr "github.com/t-bfame/diago/internal/manager"
)

// Scheduler lalalal
type Scheduler struct {
	pm *PodManager
}

// Schedule dingdingi
func (s Scheduler) Schedule(ti mgr.TestInstance) (id int, err error) {
	return s.pm.schedule(ti)
}

// Unschedule dongdongdong
func (s Scheduler) Unschedule(ti mgr.TestInstance, id int) (err error) {
	return s.pm.unschedule(ti, id)
}

// NewScheduler laalala
func NewScheduler() Scheduler {
	pm := NewPodManager()

	return Scheduler{pm}
}
