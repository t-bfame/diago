package scheduler

import (
	"errors"
	"time"

	m "github.com/t-bfame/diago/pkg/model"
	aggregator "github.com/t-bfame/diago/proto-gen/aggregator"
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

type IncomingAgg interface {
	getJobID() m.JobID
	getGroup() string
	getInstanceID() InstanceID
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
	// case *worker.Message_Metrics:
	// 	metrics := msg.GetMetrics()
	// 	timestamp, err := pytypes.Timestamp(metrics.GetTimestamp())
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	inc = Metrics{
	// 		ID:        m.JobID(metrics.GetJobId()),
	// 		Code:      metrics.GetCode(),
	// 		BytesIn:   metrics.GetBytesIn(),
	// 		BytesOut:  metrics.GetBytesOut(),
	// 		Latency:   time.Duration(metrics.GetLatency()),
	// 		Error:     metrics.GetError(),
	// 		Timestamp: timestamp,
	// 	}

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

type AggMetrics struct {
	TestId     string
	InstanceId string
	JobId      string
	Group      string
	// Latencies holds computed request latency metrics.
	Latencies []*durationpb.Duration `json:"latencies"`

	// Histogram, only if requested
	// Histogram *Histogram `json:"buckets,omitempty"`

	// BytesIn holds computed incoming byte metrics.
	BytesIn uint64 `json:"bytes_in"`

	// BytesOut holds computed outgoing byte metrics.
	BytesOut uint64 `json:"bytes_out"`

	// Earliest is the earliest timestamp in a Result set.
	Earliest time.Time `json:"earliest"`

	// Latest is the latest timestamp in a Result set.
	Latest time.Time `json:"latest"`

	// End is the latest timestamp in a Result set plus its latency.
	End time.Time `json:"end"`

	// Requests is the total number of requests executed.
	Requests uint64 `json:"requests"`

	// StatusCodes is a histogram of the responses' status codes.
	StatusCodes map[string]uint64 `json:"status_codes"`

	// Errors is a set of unique errors returned by the targets during the attack.
	Errors []string `json:"errors"`

	// Used for fast lookup of errors in Errors
	errors map[string]struct{}

	// Finish message has been received
	Finished bool
}

func (m AggMetrics) getJobID() m.JobID {
	return m.getJobID()
}

func (m AggMetrics) getInstanceID() InstanceID {
	return m.getInstanceID()
}

func (m AggMetrics) getGroup() string {
	return m.getGroup()
}

// ProtoToIncoming Convert protobufs to Incoming type message
func ProtoToIncomingAgg(msg *aggregator.Message) (IncomingAgg, error) {
	var inc IncomingAgg

	switch msg.Payload.(type) {
	case *aggregator.Message_AggMetrics:
		metrics := msg.GetAggMetrics()
		inc = AggMetrics{
			TestId:      metrics.GetTestId(),
			InstanceId:  metrics.GetInstanceId(),
			JobId:       string(m.JobID(metrics.GetJobId())),
			Group:       metrics.GetGroup(),
			Requests:    metrics.GetRequests(),
			StatusCodes: metrics.GetCodes(),
			BytesIn:     metrics.GetBytesIn(),
			BytesOut:    metrics.GetBytesOut(),
			Latencies:   metrics.GetLatencies(),
			Earliest:    metrics.GetEarliest().AsTime(),
			Latest:      metrics.GetLatest().AsTime(),
			End:         metrics.GetEnd().AsTime(),
			Errors:      metrics.GetErrors(),
			Finished:    metrics.GetFinish(),
		}
	default:
		return nil, errors.New("Invalid Incoming aggregator message type")
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
