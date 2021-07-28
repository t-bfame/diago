package scheduler

import (
	"errors"
	"log"
	"time"

	pytypes "github.com/golang/protobuf/ptypes"
	m "github.com/t-bfame/diago/pkg/model"
	worker "github.com/t-bfame/diago/proto-gen/worker"
)

// InstanceID Pod instance ID for identification
type InstanceID string

// Event for internal communication
type Event interface {
	getJobID() m.JobID
}

// Incoming messages from workers to leader
type Incoming interface {
	getJobID() m.JobID
}

// Finish message
type Finish struct {
	ID m.JobID
}

// Metrics message
type Metrics struct {
	ID        m.JobID
	Code      uint32
	BytesIn   uint64
	BytesOut  uint64
	Latency   time.Duration
	Error     string
	Timestamp time.Time
}

func (m Finish) getJobID() m.JobID {
	return m.ID
}

func (m Metrics) getJobID() m.JobID {
	return m.ID
}

// ProtoToIncoming Convert protobufs to Incoming type message
func ProtoToIncoming(msg *worker.Message) (Incoming, error) {
	var inc Incoming

	switch msg.Payload.(type) {
	case *worker.Message_Metrics:
		metrics := msg.GetMetrics()
		timestamp, err := pytypes.Timestamp(metrics.GetTimestamp())
		if err != nil {
			log.Fatal(err)
		}
		inc = Metrics{
			ID:        m.JobID(metrics.GetJobId()),
			Code:      metrics.GetCode(),
			BytesIn:   metrics.GetBytesIn(),
			BytesOut:  metrics.GetBytesOut(),
			Latency:   time.Duration(metrics.GetLatency()),
			Error:     metrics.GetError(),
			Timestamp: timestamp,
		}

	case *worker.Message_Finish:
		finish := msg.GetFinish()
		inc = Finish{
			ID: m.JobID(finish.GetJobId()),
		}

	default:
		return nil, errors.New("Invalid Incoming message type")
	}

	return inc, nil
}

// Outgoing messages to worker from leader
type Outgoing interface {
	getJobID() m.JobID
	ToProto() *worker.Message
}

// Start message
type Start struct {
	JobID                   m.JobID
	Frequency               uint64
	Duration                uint64
	HTTPMethod              string
	HTTPUrl                 string
	HTTPBody                string
	PersistResponseSampling m.SamplingRate
}

// Stop message
type Stop struct {
	ID m.JobID
}

func (m Start) getJobID() m.JobID {
	return m.JobID
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
					Body:   &m.HTTPBody,
				},
				PersistResponseSamplingRate: &worker.SamplingRate{
					Period: m.PersistResponseSampling.Period,
				},
			},
		},
	}
}

func (m Stop) getJobID() m.JobID {
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
