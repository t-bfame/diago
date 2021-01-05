package scheduler

import (
	m "github.com/t-bfame/diago/internal/model"
)

// Scheduler lalalal
type Scheduler struct {
	pm *PodManager
}

// Submit dingdingi
func (s Scheduler) Submit(j m.Job) (events chan Event, err error) {
	events = make(chan Event, 2)

	// Put job in queue for pod group
	err = s.pm.schedule(j, events)

	return events, err
}

// Stop dongdongdong
func (s Scheduler) Stop(j m.Job) (err error) {
	return s.pm.unschedule(j)
}

// Register something
func (s Scheduler) Register(group string, instance InstanceID, frequency uint64) (chan Incoming, chan Outgoing, error) {
	return s.pm.register(group, instance, frequency)
}

// NewScheduler laalala
func NewScheduler() Scheduler {
	pm := NewPodManager()

	return Scheduler{pm}
}
