package scheduler

// Scheduler lalalal
type Scheduler struct {
	pm *PodManager
}

// Schedule dingdingi
func (s Scheduler) Schedule() (err error) {

	s.pm.schedule()

	return nil
}

// NewScheduler laalala
func NewScheduler() Scheduler {
	pm := NewPodManager()

	return Scheduler{pm}
}
