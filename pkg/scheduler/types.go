package scheduler

import (
	"errors"
	"time"

	m "github.com/t-bfame/diago/pkg/model"
	worker "github.com/t-bfame/diago/proto-gen/worker"
	"google.golang.org/protobuf/types/known/durationpb"
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
	ID          m.JobID
	NMetrics    uint64
	StatusCodes map[string]uint64
	BytesIn     uint64
	BytesOut    uint64
	Latencies   []time.Duration
	Earliest    time.Time
	Latest      time.Time
	End         time.Time
	Errors      []string
	Timestamp   time.Time
}

func convertDurationPbsToDurations(l []*durationpb.Duration) []time.Duration {
	var ret []time.Duration
	for _, d := range l {
		ret = append(ret, d.AsDuration())
	}
	return ret
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
		inc = Metrics{
			ID:          m.JobID(metrics.GetJobId()),
			NMetrics:    metrics.GetNMetrics(),
			StatusCodes: metrics.GetCodes(),
			BytesIn:     metrics.GetBytesIn(),
			BytesOut:    metrics.GetBytesOut(),
			Latencies:   convertDurationPbsToDurations(metrics.GetLatencies()),
			Earliest:    metrics.GetEarliest().AsTime(),
			Latest:      metrics.GetLatest().AsTime(),
			End:         metrics.GetEnd().AsTime(),
			Errors:      metrics.GetErrors(),
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
	TestID                  m.TestID
	TestInstanceID          m.TestInstanceID
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
				JobId:          string(m.getJobID()),
				TestId:         string(m.TestID),
				TestInstanceId: string(m.TestInstanceID),
				Frequency:      m.Frequency,
				Duration:       m.Duration,
				Request: &worker.HTTPRequest{
					Method: m.HTTPMethod,
					Url:    m.HTTPUrl,
					Body:   m.HTTPBody,
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
