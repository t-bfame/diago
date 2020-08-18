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

	// Put job in queue for pod group
	err = s.pm.schedule(j, events)

	return events, err
}

// Stop dongdongdong
func (s Scheduler) Stop(j mgr.Job) (err error) {
	return s.pm.unschedule(j)
}

// Register something
func (s Scheduler) Register(group string, instance InstanceID) (chan Incoming, chan Outgoing, error) {
	return s.pm.register(group, instance)
}

// NewScheduler laalala
func NewScheduler() Scheduler {
	pm := NewPodManager()

	return Scheduler{pm}
}
