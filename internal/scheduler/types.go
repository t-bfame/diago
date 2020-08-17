package scheduler

import (
	"errors"

	mgr "github.com/t-bfame/diago/internal/manager"
	worker "github.com/t-bfame/diago/proto-gen/worker"
)

// InstanceID Pod instance ID for identification
type InstanceID string

// Event for internal communication
type Event interface {
	getJobID() mgr.JobID
}

// Incoming messages from workers to leader
type Incoming interface {
	getJobID() mgr.JobID
}

// Finish message
type Finish struct {
	ID mgr.JobID
}

// Metrics message
type Metrics struct {
	ID       mgr.JobID
	Code     uint32
	BytesIn  uint64
	BytesOut uint64
	Latency  uint64
	Error    string
}

func (m Finish) getJobID() mgr.JobID {
	return m.ID
}

func (m Metrics) getJobID() mgr.JobID {
	return m.ID
}

// ProtoToIncoming Convert protobufs to Incoming type message
func ProtoToIncoming(msg *worker.Message) (Incoming, error) {
	var inc Incoming

	switch msg.Payload.(type) {
	case *worker.Message_Metrics:
		metrics := msg.GetMetrics()
		inc = Metrics{
			ID:       mgr.JobID(metrics.GetJobId()),
			Code:     metrics.GetCode(),
			BytesIn:  metrics.GetBytesIn(),
			BytesOut: metrics.GetBytesOut(),
			Latency:  metrics.GetLatency(),
			Error:    metrics.GetError(),
		}

	case *worker.Message_Finish:
		finish := msg.GetFinish()
		inc = Finish{
			ID: mgr.JobID(finish.GetJobId()),
		}

	default:
		return nil, errors.New("Invalid Incoming message type")
	}

	return inc, nil
}

// Outgoing messages to worker from leader
type Outgoing interface {
	getJobID() mgr.JobID
	ToProto() *worker.Message
}

// Start message
type Start struct {
	ID         mgr.JobID
	Frequency  uint64
	Duration   uint64
	HTTPMethod string
	HTTPUrl    string
}

// Stop message
type Stop struct {
	ID mgr.JobID
}

func (m Start) getJobID() mgr.JobID {
	return m.ID
}

// ToProto convert Outgoing to protobuf messages
func (m Start) ToProto() *worker.Message {
	return &worker.Message{
		Payload: &worker.Message_Start{
			Start: &worker.Start{
				JobId:     string(m.getJobID()),
				Frequency: m.Frequency,
				Duration:  m.Duration,
				Request: &worker.HTTPRequest{
					Method: m.HTTPMethod,
					Url:    m.HTTPUrl,
				},
			},
		},
	}
}

func (m Stop) getJobID() mgr.JobID {
	return m.ID
}

// ToProto convert Outgoing to protobuf messages
func (m Stop) ToProto() *worker.Message {
	return &worker.Message{
		Payload: &worker.Message_Stop{
			Stop: &worker.Stop{
				JobId: string(m.getJobID()),
			},
		},
	}
}
