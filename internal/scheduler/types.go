package scheduler

import (
	"fmt"

	mgr "github.com/t-bfame/diago/internal/manager"
)

// Event passed between scheduler and manager
type Event interface {
	getJobID() mgr.JobID
	print()
}

// InstanceID Pod instance ID for identification
type InstanceID string

type Message struct {
	ID      mgr.JobID
	Message string
}

func (m Message) getJobID() mgr.JobID {
	return m.ID
}

func (m Message) print() {
	fmt.Println(m.Message)
}
