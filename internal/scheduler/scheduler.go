package scheduler

import (
	"fmt"

	mgr "github.com/t-bfame/diago/internal/manager"
)

// Scheduler lalalal
type Scheduler struct {
	pm *PodManager
}

// Schedule dingdingi
func (s Scheduler) Schedule(ti mgr.TestInstance) (err error) {

	err = s.pm.schedule(ti)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return nil
}

// NewScheduler laalala
func NewScheduler() Scheduler {
	pm := NewPodManager()

	return Scheduler{pm}
}
