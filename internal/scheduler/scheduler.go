package scheduler

import (
	mgr "github.com/t-bfame/diago/internal/manager"
)

// Scheduler lalalal
type Scheduler struct {
	pm *PodManager
}

// Submit dingdingi
func (s Scheduler) Submit(j mgr.Job) (events chan Event, err error) {
	events = make(chan Event)

	err = s.pm.schedule(j, events)

	return events, err
}

// Stop dongdongdong
func (s Scheduler) Stop(j mgr.Job, id int) (err error) {
	return s.pm.unschedule(j, id)
}

// NewScheduler laalala
func NewScheduler() Scheduler {
	pm := NewPodManager()

	return Scheduler{pm}
}
