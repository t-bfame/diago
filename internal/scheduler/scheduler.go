package scheduler

import (
	"fmt"

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
func (s Scheduler) Stop(j mgr.Job) (err error) {
	return s.pm.unschedule(j)
}

// Register something
func (s Scheduler) Register(group string, instance InstanceID) (chan Incoming, chan Outgoing, error) {
	// return s.pm.register(group, instance)

	inc := make(chan Incoming)
	out := make(chan Outgoing)

	go func() {
		out <- Start{
			ID:         "inst",
			Frequency:  1,
			Duration:   3,
			HTTPMethod: "GET",
			HTTPUrl:    "localhost:3000",
		}
	}()

	go func() {
		for lol := range inc {
			fmt.Println("Incoming:", lol)
		}
	}()

	return inc, out, nil
}

// NewScheduler laalala
func NewScheduler() Scheduler {
	pm := NewPodManager()

	return Scheduler{pm}
}
